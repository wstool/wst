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
	apitypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
	"os"
	"time"

	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/task"
	"github.com/bukka/wst/run/services"
)

type Maker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd: fnd,
	}
}

func (m *Maker) Make(config *types.DockerEnvironment) (environment.Environment, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &dockerEnvironment{
		ContainerEnvironment: *environment.NewContainerEnvironment(&config.ContainerEnvironment),
		cli:                  cli,
		namePrefix:           config.NamePrefix,
	}, nil
}

type dockerEnvironment struct {
	environment.ContainerEnvironment
	cli        *client.Client
	namePrefix string
	tasks      map[string]*dockerTask
}

func (e *dockerEnvironment) Init(ctx context.Context) error {
	return nil
}

func (e *dockerEnvironment) Destroy(ctx context.Context) error {
	for _, dockTask := range e.tasks {
		containerId := dockTask.containerId
		// Stop the container
		err := e.cli.ContainerStop(ctx, containerId, container.StopOptions{})
		if err != nil {
			return fmt.Errorf("failed to stop container %s: %w", containerId, err)
		}
		// Remove the container
		err = e.cli.ContainerRemove(ctx, containerId, container.RemoveOptions{})
		if err != nil {
			return fmt.Errorf("failed to remove container %s: %w", containerId, err)
		}
	}

	// Clear the tasks map after successful cleanup
	e.tasks = make(map[string]*dockerTask)

	return nil
}

func (e *dockerEnvironment) isContainerReady(ctx context.Context, containerID string) (bool, error) {
	resp, err := e.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	return resp.State.Running, nil
}

func (e *dockerEnvironment) RunTask(ctx context.Context, service services.Service, cmd *environment.Command) (task.Task, error) {
	sandboxContainerConfig, err := service.Sandbox().ContainerConfig()
	if err != nil {
		return nil, err
	}
	imageName := sandboxContainerConfig.Image()
	var command []string
	if cmd.Name != "" {
		command = append([]string{cmd.Name}, cmd.Args...)
	}

	// 1. Pull the Docker image if not already present
	pullOut, err := e.cli.ImagePull(ctx, imageName, apitypes.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull Docker image %s: %w", imageName, err)
	}
	defer pullOut.Close()

	// Docker container config
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   command,
	}

	// Example: Construct host configuration, such as port bindings
	hostConfig := &container.HostConfig{
		// Example: Port bindings
		PortBindings: nat.PortMap{
			// TODO: make configurable
			"80/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
		},
	}

	// 2. Create the Docker container

	containerName := fmt.Sprintf("%s-%s", e.namePrefix, service.Name())
	// TODO: create network
	containerResp, err := e.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker container for service %s: %w", service.Name(), err)
	}

	// 3. Start the Docker container
	err = e.cli.ContainerStart(ctx, containerResp.ID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start Docker container %s: %w", containerResp.ID, err)
	}

	containerId := containerResp.ID
	// Construct your dockerTask with necessary details
	dockTask := &dockerTask{
		containerName:  containerName,
		containerId:    containerId,
		containerReady: false,
	}

	e.tasks[service.FullName()] = dockTask

	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for container to be ready")
		case <-tick:
			ready, err := e.isContainerReady(context.Background(), containerId)
			if err != nil {
				return nil, fmt.Errorf("failed checking of container readiness: %v\n", err)
			}
			if ready {
				dockTask.containerReady = true
				return dockTask, nil
			}
		}
	}
}

func (e *dockerEnvironment) ExecTaskCommand(ctx context.Context, service services.Service, target task.Task, cmd *environment.Command) error {
	return fmt.Errorf("executing command is not currently supported in Docker environment")
}

func (e *dockerEnvironment) ExecTaskSignal(ctx context.Context, service services.Service, target task.Task, signal os.Signal) error {
	return fmt.Errorf("executing signal is not currently supported in Kubernetes environment")
}

func (e *dockerEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	containerID := target.Id()
	reader, err := e.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: outputType == output.Stdout || outputType == output.Any,
		ShowStderr: outputType == output.Stderr || outputType == output.Any,
		Follow:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return reader, nil
}

func (e *dockerEnvironment) RootPath(service services.Service) string {
	return ""
}

type dockerTask struct {
	containerName  string
	containerId    string
	containerReady bool
	containerUrl   string
}

func (t *dockerTask) Id() string {
	return t.containerId
}

func (t *dockerTask) Name() string {
	return t.containerName
}

func (t *dockerTask) Type() providers.Type {
	return providers.DockerType
}

func (t *dockerTask) BaseUrl() string {
	return t.containerUrl
}
