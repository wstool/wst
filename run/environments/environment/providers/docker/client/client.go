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

package client

import (
	"context"
	"fmt"
	moby "github.com/moby/moby/client"
	"github.com/wstool/wst/app"
)

type Client interface {
	ContainerCreate(ctx context.Context, options moby.ContainerCreateOptions) (moby.ContainerCreateResult, error)
	ContainerInspect(
		ctx context.Context,
		containerID string,
		options moby.ContainerInspectOptions,
	) (moby.ContainerInspectResult, error)
	ContainerLogs(
		ctx context.Context,
		containerID string,
		options moby.ContainerLogsOptions,
	) (moby.ContainerLogsResult, error)
	ContainerRemove(
		ctx context.Context,
		containerID string,
		options moby.ContainerRemoveOptions,
	) (moby.ContainerRemoveResult, error)
	ContainerStart(
		ctx context.Context,
		containerID string,
		options moby.ContainerStartOptions,
	) (moby.ContainerStartResult, error)
	ContainerStop(
		ctx context.Context,
		containerID string,
		options moby.ContainerStopOptions,
	) (moby.ContainerStopResult, error)
	ContainerWait(ctx context.Context, containerID string, options moby.ContainerWaitOptions) moby.ContainerWaitResult
	ImagePull(ctx context.Context, refStr string, options moby.ImagePullOptions) (moby.ImagePullResponse, error)
	NetworkCreate(ctx context.Context, name string, options moby.NetworkCreateOptions) (moby.NetworkCreateResult, error)
	NetworkRemove(
		ctx context.Context,
		networkID string,
		options moby.NetworkRemoveOptions,
	) (moby.NetworkRemoveResult, error)
}

type Maker interface {
	Make() (Client, error)
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{fnd: fnd}
}

type nativeMaker struct {
	fnd app.Foundation
}

func (m *nativeMaker) Make() (Client, error) {
	cli, err := moby.New(moby.FromEnv, moby.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &dockerClient{
		cli: cli,
	}, nil
}

type dockerClient struct {
	cli *moby.Client
}

// ContainerCreate creates a new Docker container based on the specified options.
func (d dockerClient) ContainerCreate(ctx context.Context, options moby.ContainerCreateOptions) (moby.ContainerCreateResult, error) {
	return d.cli.ContainerCreate(ctx, options)
}

// ContainerInspect returns detailed information about the specified container.
func (d dockerClient) ContainerInspect(ctx context.Context, containerID string, options moby.ContainerInspectOptions) (moby.ContainerInspectResult, error) {
	return d.cli.ContainerInspect(ctx, containerID, options)
}

// ContainerLogs fetches the logs of a container.
func (d dockerClient) ContainerLogs(ctx context.Context, containerID string, options moby.ContainerLogsOptions) (moby.ContainerLogsResult, error) {
	return d.cli.ContainerLogs(ctx, containerID, options)
}

// ContainerRemove removes a container specified by the container ID.
func (d dockerClient) ContainerRemove(ctx context.Context, containerID string, options moby.ContainerRemoveOptions) (moby.ContainerRemoveResult, error) {
	return d.cli.ContainerRemove(ctx, containerID, options)
}

// ContainerStart starts a container specified by the container ID.
func (d dockerClient) ContainerStart(ctx context.Context, containerID string, options moby.ContainerStartOptions) (moby.ContainerStartResult, error) {
	return d.cli.ContainerStart(ctx, containerID, options)
}

// ContainerStop stops a running container with a specified timeout.
func (d dockerClient) ContainerStop(ctx context.Context, containerID string, options moby.ContainerStopOptions) (moby.ContainerStopResult, error) {
	return d.cli.ContainerStop(ctx, containerID, options)
}

// ContainerWait waits for a container to reach a certain condition.
func (d dockerClient) ContainerWait(ctx context.Context, containerID string, options moby.ContainerWaitOptions) moby.ContainerWaitResult {
	return d.cli.ContainerWait(ctx, containerID, options)
}

// ImagePull pulls an image from a docker registry.
func (d dockerClient) ImagePull(ctx context.Context, refStr string, options moby.ImagePullOptions) (moby.ImagePullResponse, error) {
	return d.cli.ImagePull(ctx, refStr, options)
}

// NetworkCreate creates a new network with the specified options.
func (d dockerClient) NetworkCreate(ctx context.Context, name string, options moby.NetworkCreateOptions) (moby.NetworkCreateResult, error) {
	return d.cli.NetworkCreate(ctx, name, options)
}

// NetworkRemove removes a network by a network ID.
func (d dockerClient) NetworkRemove(ctx context.Context, networkID string, options moby.NetworkRemoveOptions) (moby.NetworkRemoveResult, error) {
	return d.cli.NetworkRemove(ctx, networkID, options)
}
