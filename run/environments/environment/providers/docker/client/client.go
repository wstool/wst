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
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/wstool/wst/app"
	"io"
)

type Client interface {
	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerWait(
		ctx context.Context,
		containerID string,
		condition container.WaitCondition,
	) (<-chan container.WaitResponse, <-chan error)
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error)
	NetworkRemove(ctx context.Context, networkID string) error
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
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &dockerClient{
		cli: cli,
	}, nil
}

type dockerClient struct {
	cli *client.Client
}

// ContainerCreate creates a new Docker container based on the specified options.
func (d dockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	return d.cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

// ContainerInspect returns detailed information about the specified container.
func (d dockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return d.cli.ContainerInspect(ctx, containerID)
}

// ContainerLogs fetches the logs of a container.
func (d dockerClient) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	return d.cli.ContainerLogs(ctx, container, options)
}

// ContainerRemove removes a container specified by the container ID.
func (d dockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return d.cli.ContainerRemove(ctx, containerID, options)
}

// ContainerStart starts a container specified by the container ID.
func (d dockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return d.cli.ContainerStart(ctx, containerID, options)
}

// ContainerStop stops a running container with a specified timeout.
func (d dockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return d.cli.ContainerStop(ctx, containerID, options)
}

// ContainerWait waits for a container to reach a certain condition.
func (d dockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	return d.cli.ContainerWait(ctx, containerID, condition)
}

// ImagePull pulls an image from a docker registry.
func (d dockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	return d.cli.ImagePull(ctx, refStr, options)
}

// NetworkCreate creates a new network with the specified options.
func (d dockerClient) NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error) {
	return d.cli.NetworkCreate(ctx, name, options)
}

// NetworkRemove removes a network by a network ID.
func (d dockerClient) NetworkRemove(ctx context.Context, networkID string) error {
	return d.cli.NetworkRemove(ctx, networkID)
}
