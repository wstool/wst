// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"context"
	"fmt"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/environments/environment/providers/docker/client"
	apitypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/task"
)

type Maker interface {
	Make(config *types.DockerEnvironment) (environment.Environment, error)
}

type dockerMaker struct {
	*environment.CommonMaker
	clientMaker client.Maker
}

func CreateMaker(fnd app.Foundation) Maker {
	return &dockerMaker{
		CommonMaker: environment.CreateCommonMaker(fnd),
		clientMaker: client.CreateMaker(fnd),
	}
}

func (m *dockerMaker) Make(config *types.DockerEnvironment) (environment.Environment, error) {
	cli, err := m.clientMaker.Make()
	if err != nil {
		return nil, errors.Errorf("failed to create docker client: %v", err)
	}

	return &dockerEnvironment{
		ContainerEnvironment: *m.MakeContainerEnvironment(&types.ContainerEnvironment{
			Ports:    config.Ports,
			Registry: config.Registry,
		}),
		cli:              cli,
		namePrefix:       config.NamePrefix,
		tasks:            make(map[string]*dockerTask),
		waitTickDuration: 1 * time.Second,
	}, nil
}

type dockerEnvironment struct {
	environment.ContainerEnvironment
	cli              client.Client
	namePrefix       string
	networkName      string
	networkMutex     sync.Mutex
	tasks            map[string]*dockerTask
	waitTickDuration time.Duration
}

func (e *dockerEnvironment) Init(ctx context.Context) error {
	return nil
}

func (e *dockerEnvironment) Destroy(ctx context.Context) error {
	if e.Fnd.DryRun() {
		return nil
	}
	hasError := false
	for _, dockTask := range e.tasks {
		containerId := dockTask.containerId
		// Stop the container
		err := e.cli.ContainerStop(ctx, containerId, container.StopOptions{})
		if err != nil {
			e.Fnd.Logger().Errorf("failed to stop container %s: %v", containerId, err)
			hasError = true
			continue
		}
		// Remove the container
		err = e.cli.ContainerRemove(ctx, containerId, container.RemoveOptions{})
		if err != nil {
			e.Fnd.Logger().Errorf("failed to remove container %s: %v", containerId, err)
			hasError = true
		}
	}

	// Clear the tasks map after successful cleanup
	e.tasks = make(map[string]*dockerTask)

	// Delete network
	if err := e.cli.NetworkRemove(ctx, e.networkName); err != nil {
		e.Fnd.Logger().Errorf("failed to remove network %s: %v", e.networkName, err)
		hasError = true
	}

	if hasError {
		return errors.New("Destroying docker environment failed")
	}
	return nil
}

func (e *dockerEnvironment) isContainerReady(ctx context.Context, containerID string) (bool, error) {
	resp, err := e.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, errors.Errorf("failed to inspect container: %v", err)
	}

	return resp.State.Running, nil
}

// Function to create network if it doesn't exist
func (e *dockerEnvironment) ensureNetwork(ctx context.Context, dryRun bool) error {
	e.networkMutex.Lock()
	defer e.networkMutex.Unlock()

	if e.networkName != "" {
		return nil
	}
	e.networkName = e.namePrefix
	if !dryRun {
		_, err := e.cli.NetworkCreate(ctx, e.networkName, apitypes.NetworkCreate{
			Driver: "bridge",
		})
		if err != nil {
			return errors.Errorf("failed to create network %s - %v", e.networkName, err)
		}
	}
	return nil
}

func (e *dockerEnvironment) RunTask(ctx context.Context, ss *environment.ServiceSettings, cmd *environment.Command) (task.Task, error) {
	sandboxContainerConfig := ss.ContainerConfig
	if sandboxContainerConfig == nil {
		return nil, errors.New("container config is not set")
	}
	imageName := sandboxContainerConfig.Image()
	var command []string
	if cmd.Name != "" {
		command = append([]string{cmd.Name}, cmd.Args...)
	}

	dryRun := e.Fnd.DryRun()

	if err := e.ensureNetwork(ctx, dryRun); err != nil {
		return nil, err
	}

	// Pull the Docker image if not already present
	if !dryRun {
		pullOut, err := e.cli.ImagePull(ctx, imageName, apitypes.ImagePullOptions{})
		if err != nil {
			return nil, errors.Errorf("failed to pull Docker image %s - %v", imageName, err)
		}
		defer pullOut.Close()
	}

	// Docker container config
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   command,
	}

	// Prepare host config with Port bindings
	serverPort := strconv.Itoa(int(ss.Port))
	hostUrl := ""
	var hostConfig *container.HostConfig
	if ss.Public {
		hostPort := strconv.Itoa(int(e.ReservePort()))
		portMapName := nat.Port(serverPort + "/tcp")
		hostConfig = &container.HostConfig{
			PortBindings: nat.PortMap{
				portMapName: []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: hostPort}},
			},
		}
		hostUrl = "http://localhost:" + hostPort
	} else {
		hostConfig = &container.HostConfig{}
	}

	// Bind configs and scripts to the host config
	wsConfigPaths := ss.WorkspaceConfigPaths
	wsScriptPaths := ss.WorkspaceScriptPaths
	binds := make([]string, 0, len(wsConfigPaths)+len(wsScriptPaths))
	for configName, envConfigPath := range ss.EnvironmentConfigPaths {
		wsConfigPath, found := wsConfigPaths[configName]
		if !found {
			return nil, errors.Errorf("failed to bind config %s for service %s", configName, ss.Name)
		}
		binds = append(binds, fmt.Sprintf("%s:%s", wsConfigPath, envConfigPath))
	}
	for scriptName, envScriptPath := range ss.EnvironmentScriptPaths {
		wsScriptPath, found := wsScriptPaths[scriptName]
		if !found {
			return nil, errors.Errorf("failed to bind script %s for service %s", scriptName, ss.Name)
		}
		binds = append(binds, fmt.Sprintf("%s:%s", wsScriptPath, envScriptPath))
	}
	hostConfig.Binds = binds

	// Create network config
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			e.networkName: {},
		},
	}

	// Create the Docker container
	containerName := fmt.Sprintf("%s-%s", e.namePrefix, ss.Name)
	var containerId string
	if !dryRun {
		containerResp, err := e.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, containerName)
		if err != nil {
			return nil, errors.Errorf("failed to create Docker container for service %s: %v", ss.Name, err)
		}
		containerId = containerResp.ID

		// Start the Docker container
		err = e.cli.ContainerStart(ctx, containerId, container.StartOptions{})
		if err != nil {
			removeErr := e.cli.ContainerRemove(ctx, containerId, container.RemoveOptions{})
			if removeErr != nil {
				e.Fnd.Logger().Errorf("failed to remove container %s: %v", containerId, removeErr)
			}
			return nil, errors.Errorf("failed to start Docker container %s %s: %v",
				containerName, containerResp.ID, err)
		}
	} else {
		containerId = "container"
	}
	// Construct your dockerTask with necessary details
	dockTask := &dockerTask{
		containerName:       containerName,
		containerId:         containerId,
		containerExecutable: cmd.Name,
		containerPublicUrl:  hostUrl,
		containerPrivateUrl: fmt.Sprintf("http://%s:%s", containerName, serverPort),
		containerReady:      false,
	}

	e.tasks[ss.FullName] = dockTask

	if dryRun {
		dockTask.containerReady = true
		return dockTask, nil
	}

	statusCh, errCh := e.cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, errors.Errorf("failed waiting on container %s %s to run: %v",
				containerName, containerId, err)
		}
	case <-ctx.Done():
		return nil, errors.Errorf("timed out waiting for container %s %s to be ready",
			containerName, containerId)
	case <-statusCh:
		ready, err := e.isContainerReady(ctx, containerId)
		if err != nil {
			return nil, errors.Errorf("failed checking of container %s %s readiness: %v",
				containerName, containerId, err)
		}
		if ready {
			dockTask.containerReady = true
			return dockTask, nil
		}
	}

	tick := time.Tick(e.waitTickDuration)

	for {
		select {
		case <-ctx.Done():
			return nil, errors.Errorf("timed out waiting for container %s %s to be ready",
				containerName, containerId)
		case <-tick:
			ready, err := e.isContainerReady(ctx, containerId)
			if err != nil {
				return nil, errors.Errorf("failed checking of container %s %s readiness: %v",
					containerName, containerId, err)
			}
			if ready {
				dockTask.containerReady = true
				return dockTask, nil
			}
		}
	}
}

func (e *dockerEnvironment) ExecTaskCommand(ctx context.Context, ss *environment.ServiceSettings, target task.Task, cmd *environment.Command) error {
	return errors.Errorf("executing command is not currently supported in Docker environment")
}

func (e *dockerEnvironment) ExecTaskSignal(ctx context.Context, ss *environment.ServiceSettings, target task.Task, signal os.Signal) error {
	return errors.Errorf("executing signal is not currently supported in Kubernetes environment")
}

func (e *dockerEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	if e.Fnd.DryRun() {
		return &app.DummyReaderCloser{}, nil
	}

	containerID := target.Id()
	reader, err := e.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: outputType == output.Stdout || outputType == output.Any,
		ShowStderr: outputType == output.Stderr || outputType == output.Any,
		Follow:     true,
	})
	if err != nil {
		return nil, errors.Errorf("failed to get container logs: %v", err)
	}

	return reader, nil
}

func (e *dockerEnvironment) RootPath(workspace string) string {
	return ""
}

type dockerTask struct {
	containerName       string
	containerId         string
	containerExecutable string
	containerReady      bool
	containerPublicUrl  string
	containerPrivateUrl string
}

func (t *dockerTask) Pid() int {
	return 1
}

func (t *dockerTask) Id() string {
	return t.containerId
}

func (t *dockerTask) Name() string {
	return t.containerName
}

func (t *dockerTask) Executable() string {
	return t.containerExecutable
}

func (t *dockerTask) Type() providers.Type {
	return providers.DockerType
}

func (t *dockerTask) PublicUrl() string {
	return t.containerPublicUrl
}

func (t *dockerTask) PrivateUrl() string {
	return t.containerPrivateUrl
}
