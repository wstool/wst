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
	"io"
	"net/netip"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	moby "github.com/moby/moby/client"
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/environments/environment/providers/docker/client"
	"github.com/wstool/wst/run/environments/task"
	"github.com/wstool/wst/run/resources"
)

type Maker interface {
	Make(config *types.DockerEnvironment) (environment.Environment, error)
}

type dockerMaker struct {
	*environment.CommonMaker
	clientMaker client.Maker
}

func CreateMaker(fnd app.Foundation, resourcesMaker resources.Maker) Maker {
	return &dockerMaker{
		CommonMaker: environment.CreateCommonMaker(fnd, resourcesMaker),
		clientMaker: client.CreateMaker(fnd),
	}
}

func (m *dockerMaker) Make(config *types.DockerEnvironment) (environment.Environment, error) {
	cli, err := m.clientMaker.Make()
	if err != nil {
		return nil, errors.Errorf("failed to create docker client: %v", err)
	}

	containerEnv, err := m.MakeContainerEnvironment(&types.ContainerEnvironment{
		Ports:     config.Ports,
		Resources: config.Resources,
		Registry:  config.Registry,
	})
	if err != nil {
		return nil, errors.Errorf("failed to create docker environment: %v", err)
	}

	return &dockerEnvironment{
		ContainerEnvironment: *containerEnv,
		cli:                  cli,
		namePrefix:           config.NamePrefix,
		tasks:                make(map[string]*dockerTask),
		waitTickDuration:     100 * time.Millisecond,
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

// closeOutput stops the log collection of the task if it is active.
func (e *dockerEnvironment) closeOutput(dockTask *dockerTask) error {
	hasError := false
	if dockTask.outputCollector != nil {
		if err := dockTask.outputCollector.Close(); err != nil {
			e.Fnd.Logger().Errorf("Failed to close output for container %s: %v", dockTask.containerName, err)
			hasError = true
		}
		dockTask.outputCollector = nil
	}
	if dockTask.logs != nil {
		if err := dockTask.logs.Close(); err != nil {
			e.Fnd.Logger().Errorf("Failed to close logs for container %s: %v", dockTask.containerName, err)
			hasError = true
		}
		dockTask.logs = nil
	}
	if hasError {
		return errors.Errorf("failed to close output")
	}
	return nil
}

func (e *dockerEnvironment) destroyTask(ctx context.Context, dockTask *dockerTask) error {
	hasError := false
	if e.closeOutput(dockTask) != nil {
		hasError = true
	}
	containerId := dockTask.containerId
	// Stop the container
	_, err := e.cli.ContainerStop(ctx, containerId, moby.ContainerStopOptions{})
	if err != nil && !cerrdefs.IsNotFound(err) {
		e.Fnd.Logger().Errorf("failed to stop container %s: %v", containerId, err)
		hasError = true
	}
	// Remove the container
	_, err = e.cli.ContainerRemove(ctx, containerId, moby.ContainerRemoveOptions{Force: true})
	if err != nil && !cerrdefs.IsNotFound(err) {
		e.Fnd.Logger().Errorf("failed to remove container %s: %v", containerId, err)
		hasError = true
	}
	if hasError {
		return errors.Errorf("failed to destroy container %s", containerId)
	}
	return nil
}

func (e *dockerEnvironment) Destroy(ctx context.Context) error {
	if e.Fnd.DryRun() {
		e.tasks = make(map[string]*dockerTask)
		e.networkName = ""
		return nil
	}
	hasError := false
	for _, dockTask := range e.tasks {
		if err := e.destroyTask(ctx, dockTask); err != nil {
			hasError = true
		}
	}

	// Clear the tasks map after successful cleanup
	e.tasks = make(map[string]*dockerTask)

	// Delete network
	if e.networkName != "" {
		if _, err := e.cli.NetworkRemove(ctx, e.networkName, moby.NetworkRemoveOptions{}); err != nil {
			e.Fnd.Logger().Errorf("failed to remove network %s: %v", e.networkName, err)
			hasError = true
		}
		e.networkName = ""
	}

	if hasError {
		return errors.New("Destroying docker environment failed")
	}
	return nil
}

// containerState returns the current state of the container.
func (e *dockerEnvironment) containerState(ctx context.Context, containerID string) (*container.State, error) {
	resp, err := e.cli.ContainerInspect(ctx, containerID, moby.ContainerInspectOptions{})
	if err != nil {
		return nil, errors.Errorf("failed to inspect container: %v", err)
	}
	if resp.Container.State == nil {
		return nil, errors.Errorf("container state is not available")
	}
	return resp.Container.State, nil
}

// Function to create network if it doesn't exist
func (e *dockerEnvironment) ensureNetwork(ctx context.Context, dryRun bool) error {
	e.networkMutex.Lock()
	defer e.networkMutex.Unlock()

	if e.networkName != "" {
		return nil
	}
	networkName := e.namePrefix
	if networkName == "" {
		networkName = "wst"
	}
	if !dryRun {
		_, err := e.cli.NetworkCreate(ctx, networkName, moby.NetworkCreateOptions{
			Driver: "bridge",
		})
		if err != nil {
			return errors.Errorf("failed to create network %s - %v", networkName, err)
		}
	}
	e.networkName = networkName
	return nil
}

// ensureImage pulls the image if it is not available locally and waits for the pull to finish.
func (e *dockerEnvironment) ensureImage(ctx context.Context, imageName string) error {
	_, err := e.cli.ImageInspect(ctx, imageName)
	if err == nil {
		e.Fnd.Logger().Debugf("Docker image %s is available locally", imageName)
		return nil
	}
	if !cerrdefs.IsNotFound(err) {
		return errors.Errorf("failed to inspect Docker image %s - %v", imageName, err)
	}
	e.Fnd.Logger().Debugf("Pulling Docker image %s", imageName)
	pullOut, err := e.cli.ImagePull(ctx, imageName, moby.ImagePullOptions{})
	if err != nil {
		return errors.Errorf("failed to pull Docker image %s - %v", imageName, err)
	}
	defer pullOut.Close()
	if err = pullOut.Wait(ctx); err != nil {
		return errors.Errorf("failed to finish pulling Docker image %s - %v", imageName, err)
	}
	return nil
}

// createContainer creates the container and replaces an existing container with the same name.
func (e *dockerEnvironment) createContainer(
	ctx context.Context,
	containerName string,
	options moby.ContainerCreateOptions,
) (string, error) {
	containerResp, err := e.cli.ContainerCreate(ctx, options)
	if err != nil && cerrdefs.IsConflict(err) {
		e.Fnd.Logger().Debugf("Removing existing container %s", containerName)
		_, removeErr := e.cli.ContainerRemove(ctx, containerName, moby.ContainerRemoveOptions{Force: true})
		if removeErr != nil {
			return "", errors.Errorf("failed to remove existing container %s: %v", containerName, removeErr)
		}
		containerResp, err = e.cli.ContainerCreate(ctx, options)
	}
	if err != nil {
		return "", errors.Errorf("failed to create Docker container %s: %v", containerName, err)
	}
	return containerResp.ID, nil
}

// waitForContainer waits until the container is running or fails if it exits.
func (e *dockerEnvironment) waitForContainer(ctx context.Context, dockTask *dockerTask) error {
	for {
		state, err := e.containerState(ctx, dockTask.containerId)
		if err != nil {
			return errors.Errorf("failed checking of container %s %s readiness: %v",
				dockTask.containerName, dockTask.containerId, err)
		}
		if state.Running {
			dockTask.containerReady = true
			return nil
		}
		if state.Status == container.StateExited || state.Status == container.StateDead {
			return errors.Errorf("container %s %s exited with code %d: %s",
				dockTask.containerName, dockTask.containerId, state.ExitCode, state.Error)
		}
		select {
		case <-ctx.Done():
			return errors.Errorf("timed out waiting for container %s %s to be ready",
				dockTask.containerName, dockTask.containerId)
		case <-time.After(e.waitTickDuration):
		}
	}
}

// startOutput opens the container log stream and collects the demultiplexed stdout and stderr for the whole life
// of the task. The stream is not bound to the action context because the output is read by later actions.
func (e *dockerEnvironment) startOutput(ctx context.Context, dockTask *dockerTask) error {
	logs, err := e.cli.ContainerLogs(context.WithoutCancel(ctx), dockTask.containerId, moby.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return errors.Errorf("failed to get container %s logs: %v", dockTask.containerName, err)
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	go func() {
		_, copyErr := stdcopy.StdCopy(stdoutWriter, stderrWriter, logs)
		_ = stdoutWriter.CloseWithError(copyErr)
		_ = stderrWriter.CloseWithError(copyErr)
	}()

	collector := e.OutputMaker.MakeCollector(dockTask.containerId)
	if err = collector.Start(stdoutReader, stderrReader); err != nil {
		_ = logs.Close()
		return errors.Errorf("failed to start output collection of container %s: %v", dockTask.containerName, err)
	}
	dockTask.outputCollector = collector
	dockTask.logs = logs
	return nil
}

func (e *dockerEnvironment) RunTask(ctx context.Context, ss *environment.ServiceSettings, cmd *environment.Command) (task.Task, error) {
	sandboxContainerConfig := ss.ContainerConfig
	if sandboxContainerConfig == nil {
		return nil, errors.New("container config is not set")
	}
	imageName := sandboxContainerConfig.Image()
	var command []string
	var executable string
	if cmd != nil && cmd.Name != "" {
		executable = cmd.Name
		command = append([]string{cmd.Name}, cmd.Args...)
	} else {
		executable = ss.Name
	}

	dryRun := e.Fnd.DryRun()

	if err := e.ensureNetwork(ctx, dryRun); err != nil {
		return nil, err
	}

	// Make sure the image is available
	if !dryRun {
		if err := e.ensureImage(ctx, imageName); err != nil {
			return nil, err
		}
	}

	// Docker container config
	serverPort := strconv.Itoa(int(ss.ServerPort))
	containerPort, err := network.ParsePort(serverPort + "/tcp")
	if err != nil {
		return nil, errors.Errorf("invalid server port %s for service %s: %v", serverPort, ss.Name, err)
	}
	containerConfig := &container.Config{
		Image:        imageName,
		Cmd:          command,
		ExposedPorts: network.PortSet{containerPort: struct{}{}},
	}

	// Prepare host config with Port bindings
	hostUrl := ""
	var hostConfig *container.HostConfig
	if ss.Public {
		hostPort := strconv.Itoa(int(ss.Port))
		hostConfig = &container.HostConfig{
			PortBindings: network.PortMap{
				containerPort: []network.PortBinding{{HostIP: netip.IPv4Unspecified(), HostPort: hostPort}},
			},
		}
		hostUrl = "://localhost:" + hostPort
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
	for _, cert := range ss.Certificates {
		if cert.CertificateSourceFilePath != "" {
			binds = append(binds, fmt.Sprintf("%s:%s", cert.CertificateSourceFilePath, cert.CertificateFilePath))
		}
		if cert.PrivateKeySourceFilePath != "" {
			binds = append(binds, fmt.Sprintf("%s:%s", cert.PrivateKeySourceFilePath, cert.PrivateKeyFilePath))
		}
	}
	hostConfig.Binds = binds

	// Create network config
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			e.networkName: {},
		},
	}

	containerName := fmt.Sprintf("%s-%s", e.networkName, ss.Name)
	dockTask := &dockerTask{
		containerName:       containerName,
		containerExecutable: executable,
		containerPublicUrl:  hostUrl,
		containerPrivateUrl: fmt.Sprintf("://%s:%s", containerName, serverPort),
		containerReady:      false,
	}

	if dryRun {
		dockTask.containerId = "container"
		dockTask.containerReady = true
		e.tasks[ss.FullName] = dockTask
		return dockTask, nil
	}

	// Create the Docker container
	containerId, err := e.createContainer(ctx, containerName, moby.ContainerCreateOptions{
		Config:           containerConfig,
		HostConfig:       hostConfig,
		NetworkingConfig: networkingConfig,
		Name:             containerName,
	})
	if err != nil {
		return nil, err
	}
	dockTask.containerId = containerId
	// Register the task before starting so that it is cleaned up by Destroy even if the start fails.
	e.tasks[ss.FullName] = dockTask

	// Start the Docker container
	_, err = e.cli.ContainerStart(ctx, containerId, moby.ContainerStartOptions{})
	if err != nil {
		return nil, errors.Errorf("failed to start Docker container %s %s: %v", containerName, containerId, err)
	}

	if err = e.waitForContainer(ctx, dockTask); err != nil {
		return nil, err
	}

	if err = e.startOutput(ctx, dockTask); err != nil {
		return nil, err
	}

	return dockTask, nil
}

func convertTask(target task.Task) (*dockerTask, error) {
	dockTask, ok := target.(*dockerTask)
	if !ok || dockTask == nil {
		return nil, errors.Errorf("task in not a Docker task")
	}
	return dockTask, nil
}

func (e *dockerEnvironment) ExecTaskCommand(
	ctx context.Context,
	ss *environment.ServiceSettings,
	target task.Task,
	cmd *environment.Command,
	oc output.Collector,
) error {
	dockTask, err := convertTask(target)
	if err != nil {
		return err
	}
	if e.Fnd.DryRun() {
		return nil
	}

	var stdout, stderr io.Writer = io.Discard, io.Discard
	if oc != nil && !reflect.ValueOf(oc).IsNil() {
		stdout = oc.StdoutWriter()
		stderr = oc.StderrWriter()
	}

	execResp, err := e.cli.ExecCreate(ctx, dockTask.containerId, moby.ExecCreateOptions{
		Cmd:          append([]string{cmd.Name}, cmd.Args...),
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return errors.Errorf("failed to create exec for command %s in container %s: %v",
			cmd.Name, dockTask.containerName, err)
	}
	attachResp, err := e.cli.ExecAttach(ctx, execResp.ID, moby.ExecAttachOptions{})
	if err != nil {
		return errors.Errorf("failed to attach exec for command %s in container %s: %v",
			cmd.Name, dockTask.containerName, err)
	}
	_, copyErr := stdcopy.StdCopy(stdout, stderr, attachResp.Reader)
	attachResp.Close()
	if copyErr != nil {
		return errors.Errorf("failed to read output of command %s in container %s: %v",
			cmd.Name, dockTask.containerName, copyErr)
	}
	inspectResp, err := e.cli.ExecInspect(ctx, execResp.ID, moby.ExecInspectOptions{})
	if err != nil {
		return errors.Errorf("failed to inspect exec for command %s in container %s: %v",
			cmd.Name, dockTask.containerName, err)
	}
	if inspectResp.ExitCode != 0 {
		return errors.Errorf("command %s in container %s exited with code %d",
			cmd.Name, dockTask.containerName, inspectResp.ExitCode)
	}
	return nil
}

func (e *dockerEnvironment) ExecTaskSignal(ctx context.Context, ss *environment.ServiceSettings, target task.Task, signal os.Signal) error {
	dockTask, err := convertTask(target)
	if err != nil {
		return err
	}
	signalName, ok := environment.SignalName(signal)
	if !ok {
		return errors.Errorf("signal %v is not supported in Docker environment", signal)
	}
	if e.Fnd.DryRun() {
		return nil
	}
	_, err = e.cli.ContainerKill(ctx, dockTask.containerId, moby.ContainerKillOptions{Signal: signalName})
	if err != nil {
		return errors.Errorf("failed to send signal %s to container %s: %v", signalName, dockTask.containerName, err)
	}
	return nil
}

func (e *dockerEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	dockTask, err := convertTask(target)
	if err != nil {
		return nil, err
	}
	if e.Fnd.DryRun() {
		return &app.DummyReaderCloser{}, nil
	}
	if dockTask.outputCollector == nil {
		return nil, errors.Errorf("output of container %s is not available", dockTask.containerName)
	}
	return dockTask.outputCollector.Reader(ctx, outputType)
}

func (e *dockerEnvironment) RootPath(workspace string) string {
	return ""
}

func (e *dockerEnvironment) Mkdir(serviceName string, path string, perm os.FileMode) error {
	// Currently it is a user responsibility to make sure that directory exists in the container
	return nil
}

func (e *dockerEnvironment) ServiceLocalAddress(serviceName string, servicePort, serverPort int32) string {
	// The server has to listen on all interfaces so that it is reachable through the published port.
	return fmt.Sprintf("0.0.0.0:%d", serverPort)
}

func (e *dockerEnvironment) ServiceLocalPort(servicePort, serverPort int32) int32 {
	return serverPort
}

func (e *dockerEnvironment) ServicePrivateAddress(serviceName string, servicePort, serverPort int32) string {
	return serviceName
}

type dockerTask struct {
	containerName       string
	containerId         string
	containerExecutable string
	containerReady      bool
	containerPublicUrl  string
	containerPrivateUrl string
	outputCollector     output.Collector
	logs                io.Closer
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

func (t *dockerTask) PublicUrl(scheme string) string {
	return scheme + t.containerPublicUrl
}

func (t *dockerTask) PrivateUrl(scheme string) string {
	return scheme + t.containerPrivateUrl
}
