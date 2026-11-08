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
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"iter"
	"net"
	"net/netip"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/api/types/network"
	moby "github.com/moby/moby/client"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	outputMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/output"
	dockerClientMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/providers/docker/client"
	resourcesMocks "github.com/wstool/wst/mocks/generated/run/resources"
	certificatesMocks "github.com/wstool/wst/mocks/generated/run/resources/certificates"
	scriptsMocks "github.com/wstool/wst/mocks/generated/run/resources/scripts"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/environments/task"
	"github.com/wstool/wst/run/resources"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/resources/scripts"
	"github.com/wstool/wst/run/sandboxes/containers"
)

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	resourcesMaker := resourcesMocks.NewMockMaker(t)
	tests := []struct {
		name           string
		fnd            app.Foundation
		resourcesMaker resources.Maker
	}{
		{
			name:           "create maker",
			fnd:            fndMock,
			resourcesMaker: resourcesMaker,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateMaker(tt.fnd, resourcesMaker)
			maker, ok := got.(*dockerMaker)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, maker.Fnd)
			assert.Equal(t, tt.resourcesMaker, maker.ResourcesMaker)
		})
	}
}

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name             string
		config           *types.DockerEnvironment
		setupMocks       func(*testing.T, *dockerClientMocks.MockMaker, *resourcesMocks.MockMaker) (*dockerClientMocks.MockClient, *resources.Resources)
		getExpectedEnv   func(fndMock *appMocks.MockFoundation, cli *dockerClientMocks.MockClient, rscs *resources.Resources) *dockerEnvironment
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful docker environment maker creation with resources",
			config: &types.DockerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				NamePrefix: "test",
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{
						"web_ssl": {
							Certificate: "-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----",
							PrivateKey:  "-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----",
						},
					},
					Scripts: map[string]types.Script{
						"init_script": {
							Content: "#!/bin/bash\necho 'Docker init'",
							Path:    "/docker/init.sh",
							Mode:    "0755",
						},
					},
				},
			},
			setupMocks: func(t *testing.T, m *dockerClientMocks.MockMaker, r *resourcesMocks.MockMaker) (*dockerClientMocks.MockClient, *resources.Resources) {
				c := dockerClientMocks.NewMockClient(t)
				m.On("Make").Return(c, nil)

				expectedResources := &resources.Resources{
					Certificates: map[string]certificates.Certificate{
						"web_ssl": certificatesMocks.NewMockCertificate(t),
					},
					Scripts: map[string]scripts.Script{
						"init_script": scriptsMocks.NewMockScript(t),
					},
				}

				// With fixed Docker maker implementation, it should pass the actual resources
				r.On("Make", types.Resources{
					Certificates: map[string]types.Certificate{
						"web_ssl": {
							Certificate: "-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----",
							PrivateKey:  "-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----",
						},
					},
					Scripts: map[string]types.Script{
						"init_script": {
							Content: "#!/bin/bash\necho 'Docker init'",
							Path:    "/docker/init.sh",
							Mode:    "0755",
						},
					},
				}).Return(expectedResources, nil)

				return c, expectedResources
			},
			getExpectedEnv: func(
				fndMock *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
				rscs *resources.Resources,
			) *dockerEnvironment {
				return &dockerEnvironment{
					ContainerEnvironment: environment.ContainerEnvironment{
						CommonEnvironment: environment.CommonEnvironment{
							Fnd:  fndMock,
							Used: false,
							Ports: environment.Ports{
								Start: 8000,
								Used:  8000,
								End:   8500,
							},
							EnvResources: rscs,
						},
						Registry: environment.ContainerRegistry{
							Auth: environment.ContainerRegistryAuth{
								Username: "u1",
								Password: "p1",
							},
						},
					},
					cli:              cli,
					namePrefix:       "test",
					networkName:      "",
					tasks:            make(map[string]*dockerTask),
					waitTickDuration: 100 * time.Millisecond,
				}
			},
		},
		{
			name: "successful docker environment maker creation with empty resources",
			config: &types.DockerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				NamePrefix: "test",
				Resources:  types.Resources{},
			},
			setupMocks: func(t *testing.T, m *dockerClientMocks.MockMaker, r *resourcesMocks.MockMaker) (*dockerClientMocks.MockClient, *resources.Resources) {
				c := dockerClientMocks.NewMockClient(t)
				m.On("Make").Return(c, nil)

				expectedResources := &resources.Resources{}
				r.On("Make", types.Resources{}).Return(expectedResources, nil)

				return c, expectedResources
			},
			getExpectedEnv: func(
				fndMock *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
				rscs *resources.Resources,
			) *dockerEnvironment {
				return &dockerEnvironment{
					ContainerEnvironment: environment.ContainerEnvironment{
						CommonEnvironment: environment.CommonEnvironment{
							Fnd:  fndMock,
							Used: false,
							Ports: environment.Ports{
								Start: 8000,
								Used:  8000,
								End:   8500,
							},
							EnvResources: rscs,
						},
						Registry: environment.ContainerRegistry{
							Auth: environment.ContainerRegistryAuth{
								Username: "u1",
								Password: "p1",
							},
						},
					},
					cli:              cli,
					namePrefix:       "test",
					networkName:      "",
					tasks:            make(map[string]*dockerTask),
					waitTickDuration: 100 * time.Millisecond,
				}
			},
		},
		{
			name: "failed docker environment maker creation due to resource failure",
			config: &types.DockerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				NamePrefix: "test",
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Content: "echo test",
							Mode:    "invalid_mode",
						},
					},
				},
			},
			setupMocks: func(t *testing.T, m *dockerClientMocks.MockMaker, r *resourcesMocks.MockMaker) (*dockerClientMocks.MockClient, *resources.Resources) {
				c := dockerClientMocks.NewMockClient(t)
				m.On("Make").Return(c, nil)

				// Fix: The Make method will be called with the actual resources from the config
				// This matches what's actually passed to MakeContainerEnvironment
				r.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Content: "echo test",
							Path:    "",
							Mode:    "invalid_mode",
						},
					},
				}).Return(nil, errors.New("resource creation failed"))

				return c, nil
			},
			expectError:      true,
			expectedErrorMsg: "resource creation failed",
		},
		{
			name: "failed docker environment maker creation due to client failure",
			config: &types.DockerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				NamePrefix: "test",
				Resources:  types.Resources{},
			},
			setupMocks: func(t *testing.T, m *dockerClientMocks.MockMaker, r *resourcesMocks.MockMaker) (*dockerClientMocks.MockClient, *resources.Resources) {
				// Client creation fails first, so resources are never created
				m.On("Make").Return(nil, errors.New("docker error"))
				// No resource mock needed since client creation fails first

				return nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to create docker client: docker error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			resourcesMaker := resourcesMocks.NewMockMaker(t)
			clientMakerMock := dockerClientMocks.NewMockMaker(t)
			m := &dockerMaker{
				CommonMaker: &environment.CommonMaker{
					Fnd:            fndMock,
					ResourcesMaker: resourcesMaker,
				},
				clientMaker: clientMakerMock,
			}

			cli, expectedResources := tt.setupMocks(t, clientMakerMock, resourcesMaker)

			got, err := m.Make(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualEnv, ok := got.(*dockerEnvironment)
				assert.True(t, ok)
				expectedEnv := tt.getExpectedEnv(fndMock, cli, expectedResources)
				assert.Equal(t, expectedEnv, actualEnv)

				// Assert that resources are properly set
				assert.Equal(t, expectedResources, actualEnv.Resources())
			}

			// Only assert expectations that were actually set up
			if !tt.expectError || tt.expectedErrorMsg != "failed to create docker client: docker error" {
				resourcesMaker.AssertExpectations(t)
			}
			clientMakerMock.AssertExpectations(t)
		})
	}
}

func Test_dockerEnvironment_Init(t *testing.T) {
	env := &dockerEnvironment{}
	ctx := context.Background()
	assert.Nil(t, env.Init(ctx))
}

type pullReaderCloser struct {
	msg     string
	waitErr error
	closed  int
}

func (b *pullReaderCloser) Read(p []byte) (n int, err error) {
	if len(b.msg) > 0 {
		n = copy(p, b.msg)
		b.msg = b.msg[n:]
		return n, nil
	}
	return 0, io.EOF
}

func (b *pullReaderCloser) JSONMessages(ctx context.Context) iter.Seq2[jsonstream.Message, error] {
	return func(yield func(jsonstream.Message, error) bool) {}
}

func (b *pullReaderCloser) Wait(ctx context.Context) error {
	return b.waitErr
}

func (b *pullReaderCloser) Close() error {
	b.closed++
	return nil
}

// muxFrame builds one frame of the Docker multiplexed stream (stdcopy format) for the given stream type.
func muxFrame(stream byte, data string) []byte {
	frame := make([]byte, 8+len(data))
	frame[0] = stream
	binary.BigEndian.PutUint32(frame[4:], uint32(len(data)))
	copy(frame[8:], data)
	return frame
}

const (
	muxStdout byte = 1
	muxStderr byte = 2
)

// logsReaderCloser is a container logs stream with the given multiplexed stdout and stderr content.
type logsReaderCloser struct {
	*bytes.Reader
	closeErr error
	closed   int
}

func newLogsReaderCloser(stdout, stderr string) *logsReaderCloser {
	var buf []byte
	if stdout != "" {
		buf = append(buf, muxFrame(muxStdout, stdout)...)
	}
	if stderr != "" {
		buf = append(buf, muxFrame(muxStderr, stderr)...)
	}
	return &logsReaderCloser{Reader: bytes.NewReader(buf)}
}

func (l *logsReaderCloser) Close() error {
	l.closed++
	return l.closeErr
}

type fakeCloser struct {
	err    error
	closed int
}

func (c *fakeCloser) Close() error {
	c.closed++
	return c.err
}

type invalidTask struct{}

func (t *invalidTask) Pid() int {
	return 1
}

func (t *invalidTask) Id() string {
	return ""
}

func (t *invalidTask) Executable() string {
	return ""
}

func (t *invalidTask) Name() string {
	return ""
}

func (t *invalidTask) Type() providers.Type {
	return providers.DockerType
}

func (t *invalidTask) PublicUrl(scheme string) string {
	return scheme
}

func (t *invalidTask) PrivateUrl(scheme string) string {
	return scheme
}

// testEnv groups the environment under test with all its mocks.
type testEnv struct {
	env *dockerEnvironment
	fnd *appMocks.MockFoundation
	cli *dockerClientMocks.MockClient
	om  *outputMocks.MockMaker
}

func newTestEnv(t *testing.T, networkName string, tasks map[string]*dockerTask) *testEnv {
	fndMock := appMocks.NewMockFoundation(t)
	mockLogger := external.NewMockLogger()
	fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
	te := &testEnv{
		fnd: fndMock,
		cli: dockerClientMocks.NewMockClient(t),
		om:  outputMocks.NewMockMaker(t),
	}
	if tasks == nil {
		tasks = make(map[string]*dockerTask)
	}
	te.env = &dockerEnvironment{
		ContainerEnvironment: environment.ContainerEnvironment{
			CommonEnvironment: environment.CommonEnvironment{
				Fnd:         fndMock,
				OutputMaker: te.om,
				Ports: environment.Ports{
					Start: 8080,
					Used:  8080,
					End:   8180,
				},
			},
			Registry: environment.ContainerRegistry{},
		},
		cli:              te.cli,
		namePrefix:       "wst",
		networkName:      networkName,
		tasks:            tasks,
		waitTickDuration: time.Millisecond,
	}
	return te
}

func Test_dockerEnvironment_Destroy(t *testing.T) {
	stopOpts := moby.ContainerStopOptions{}
	removeOpts := moby.ContainerRemoveOptions{Force: true}
	networkOpts := moby.NetworkRemoveOptions{}
	makeTask := func(id string, collector output.Collector, logs io.Closer) *dockerTask {
		return &dockerTask{
			containerName:   "n" + id,
			containerId:     "cid" + id,
			outputCollector: collector,
			logs:            logs,
		}
	}
	tests := []struct {
		name             string
		dryRun           bool
		networkName      string
		setupMocks       func(t *testing.T, ctx context.Context, te *testEnv) map[string]*dockerTask
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:        "successful docker env destroying",
			networkName: "net1",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*dockerTask {
				collector := outputMocks.NewMockCollector(t)
				collector.On("Close").Return(nil)
				logs := &fakeCloser{}
				te.cli.On("ContainerStop", ctx, "cid1", stopOpts).Return(moby.ContainerStopResult{}, nil)
				te.cli.On("ContainerRemove", ctx, "cid1", removeOpts).Return(moby.ContainerRemoveResult{}, nil)
				te.cli.On("ContainerStop", ctx, "cid2", stopOpts).Return(moby.ContainerStopResult{}, nil)
				te.cli.On("ContainerRemove", ctx, "cid2", removeOpts).Return(moby.ContainerRemoveResult{}, nil)
				te.cli.On("NetworkRemove", ctx, "net1", networkOpts).Return(moby.NetworkRemoveResult{}, nil)
				return map[string]*dockerTask{
					"t1": makeTask("1", collector, logs),
					"t2": makeTask("2", nil, nil),
				}
			},
		},
		{
			name:        "successful docker env destroying when container is already gone",
			networkName: "net1",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*dockerTask {
				te.cli.On("ContainerStop", ctx, "cid1", stopOpts).Return(moby.ContainerStopResult{}, cerrdefs.ErrNotFound)
				te.cli.On("ContainerRemove", ctx, "cid1", removeOpts).Return(moby.ContainerRemoveResult{}, cerrdefs.ErrNotFound)
				te.cli.On("NetworkRemove", ctx, "net1", networkOpts).Return(moby.NetworkRemoveResult{}, nil)
				return map[string]*dockerTask{"t1": makeTask("1", nil, nil)}
			},
		},
		{
			name:        "successful docker env destroying without network",
			networkName: "",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*dockerTask {
				return map[string]*dockerTask{}
			},
		},
		{
			name:        "successful docker env destroying in dry run",
			dryRun:      true,
			networkName: "net1",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*dockerTask {
				return map[string]*dockerTask{"t1": makeTask("1", nil, nil)}
			},
		},
		{
			name:        "failed docker env destroying due to all errors",
			networkName: "net1",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*dockerTask {
				collector := outputMocks.NewMockCollector(t)
				collector.On("Close").Return(errors.New("close err"))
				logs := &fakeCloser{err: errors.New("logs err")}
				te.cli.On("ContainerStop", ctx, "cid1", stopOpts).Return(moby.ContainerStopResult{}, errors.New("stop err"))
				te.cli.On("ContainerRemove", ctx, "cid1", removeOpts).Return(moby.ContainerRemoveResult{}, errors.New("rm err"))
				te.cli.On("NetworkRemove", ctx, "net1", networkOpts).Return(moby.NetworkRemoveResult{}, errors.New("net err"))
				return map[string]*dockerTask{"t1": makeTask("1", collector, logs)}
			},
			expectError:      true,
			expectedErrorMsg: "Destroying docker environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			te := newTestEnv(t, tt.networkName, nil)
			te.fnd.On("DryRun").Return(tt.dryRun)
			tasks := tt.setupMocks(t, ctx, te)
			te.env.tasks = tasks

			err := te.env.Destroy(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Empty(t, te.env.tasks)
			assert.Equal(t, "", te.env.networkName)
			if !tt.dryRun {
				for _, tsk := range tasks {
					assert.Nil(t, tsk.outputCollector)
					assert.Nil(t, tsk.logs)
				}
			}
		})
	}
}

func getTestingServiceSettings(public bool) *environment.ServiceSettings {
	return &environment.ServiceSettings{
		Name:       "svc",
		FullName:   "mysvc",
		UniqueName: "i1-svc",
		Port:       8080,
		ServerPort: 1234,
		Public:     public,
		ContainerConfig: &containers.ContainerConfig{
			ImageName:        "wst",
			ImageTag:         "test",
			RegistryUsername: "u1",
			RegistryPassword: "p1",
		},
		EnvironmentConfigPaths: map[string]string{
			"main_conf": "/etc/main.conf",
		},
		EnvironmentScriptPaths: map[string]string{
			"test_php": "/www/test.php",
		},
		WorkspaceConfigPaths: map[string]string{
			"main_conf": "/tmp/wst/main.conf",
		},
		WorkspaceScriptPaths: map[string]string{
			"test_php": "/tmp/wst/my_test.php",
		},
	}
}

func getTestingCreateOptions(public bool, cmd []string, extraBinds ...string) moby.ContainerCreateOptions {
	containerPort, _ := network.ParsePort("1234/tcp")
	hostConfig := &container.HostConfig{}
	if public {
		hostConfig.PortBindings = network.PortMap{
			containerPort: []network.PortBinding{{HostIP: netip.IPv4Unspecified(), HostPort: "8080"}},
		}
	}
	hostConfig.Binds = append([]string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/my_test.php:/www/test.php"}, extraBinds...)
	return moby.ContainerCreateOptions{
		Config: &container.Config{
			Image:        "wst:test",
			Cmd:          cmd,
			ExposedPorts: network.PortSet{containerPort: struct{}{}},
		},
		HostConfig: hostConfig,
		NetworkingConfig: &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"wst": {},
			},
		},
		Name: "wst-svc",
	}
}

func inspectResult(status container.ContainerState, running bool, exitCode int) moby.ContainerInspectResult {
	return moby.ContainerInspectResult{
		Container: container.InspectResponse{
			State: &container.State{
				Status:   status,
				Running:  running,
				ExitCode: exitCode,
				Error:    "state err",
			},
		},
	}
}

func Test_dockerEnvironment_RunTask(t *testing.T) {
	phpCmd := &environment.Command{Name: "php", Args: []string{"test.php", "run"}}
	phpArgs := []string{"php", "test.php", "run"}
	networkCreateOpts := moby.NetworkCreateOptions{Driver: "bridge"}
	logsOpts := moby.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true}
	inspectOpts := moby.ContainerInspectOptions{}

	expectNetwork := func(ctx context.Context, te *testEnv) {
		te.cli.On("NetworkCreate", ctx, "wst", networkCreateOpts).Return(moby.NetworkCreateResult{ID: "nid"}, nil)
	}
	expectImage := func(ctx context.Context, te *testEnv) {
		te.cli.On("ImageInspect", ctx, "wst:test").Return(moby.ImageInspectResult{}, nil)
	}
	expectCreateAndStart := func(ctx context.Context, te *testEnv, opts moby.ContainerCreateOptions) {
		te.cli.On("ContainerCreate", ctx, opts).Return(moby.ContainerCreateResult{ID: "cid"}, nil)
		te.cli.On("ContainerStart", ctx, "cid", moby.ContainerStartOptions{}).Return(moby.ContainerStartResult{}, nil)
	}
	expectRunning := func(ctx context.Context, te *testEnv) {
		te.cli.On("ContainerInspect", ctx, "cid", inspectOpts).Return(inspectResult(container.StateRunning, true, 0), nil)
	}
	expectOutput := func(t *testing.T, te *testEnv) (*outputMocks.MockCollector, *logsReaderCloser) {
		logs := newLogsReaderCloser("out", "err")
		te.cli.On("ContainerLogs", mock.Anything, "cid", logsOpts).Return(logs, nil)
		collector := outputMocks.NewMockCollector(t)
		collector.On("Start", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			// Both pipes have to be read concurrently as the demultiplexing writes them in sequence.
			stderrCh := make(chan string, 1)
			go func() {
				stderr, _ := io.ReadAll(args.Get(1).(io.Reader))
				stderrCh <- string(stderr)
			}()
			stdout, _ := io.ReadAll(args.Get(0).(io.Reader))
			assert.Equal(t, "out", string(stdout))
			assert.Equal(t, "err", <-stderrCh)
		})
		te.om.On("MakeCollector", "cid").Return(collector)
		return collector, logs
	}
	expectedTask := func(public bool, executable string, collector output.Collector, logs io.Closer) *dockerTask {
		publicUrl := ""
		if public {
			publicUrl = "://localhost:8080"
		}
		return &dockerTask{
			containerName:       "wst-svc",
			containerId:         "cid",
			containerExecutable: executable,
			containerReady:      true,
			containerPublicUrl:  publicUrl,
			containerPrivateUrl: "://wst-svc:1234",
			outputCollector:     collector,
			logs:                logs,
		}
	}

	tests := []struct {
		name            string
		dryRun          bool
		networkName     string
		ss              *environment.ServiceSettings
		cmd             *environment.Command
		setupMocks      func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask
		expectError     bool
		expectedErrMsg  string
		expectTaskSaved bool
	}{
		{
			name: "successful public run",
			ss:   getTestingServiceSettings(true),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectNetwork(ctx, te)
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(true, phpArgs))
				expectRunning(ctx, te)
				collector, logs := expectOutput(t, te)
				return expectedTask(true, "php", collector, logs)
			},
			expectTaskSaved: true,
		},
		{
			name:        "successful private run with existing network and certificates",
			networkName: "wst",
			ss: func() *environment.ServiceSettings {
				ss := getTestingServiceSettings(false)
				ss.Certificates = map[string]*certificates.RenderedCertificate{
					"c1": {
						CertificateSourceFilePath: "/tmp/wst/c1.crt",
						CertificateFilePath:       "/etc/c1.crt",
						PrivateKeySourceFilePath:  "/tmp/wst/c1.key",
						PrivateKeyFilePath:        "/etc/c1.key",
					},
				}
				return ss
			}(),
			cmd: phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(false, phpArgs,
					"/tmp/wst/c1.crt:/etc/c1.crt", "/tmp/wst/c1.key:/etc/c1.key"))
				expectRunning(ctx, te)
				collector, logs := expectOutput(t, te)
				return expectedTask(false, "php", collector, logs)
			},
			expectTaskSaved: true,
		},
		{
			name: "successful run without command uses service name as executable",
			ss:   getTestingServiceSettings(false),
			cmd:  nil,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectNetwork(ctx, te)
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(false, nil))
				expectRunning(ctx, te)
				collector, logs := expectOutput(t, te)
				return expectedTask(false, "svc", collector, logs)
			},
			expectTaskSaved: true,
		},
		{
			name: "successful run with image pull",
			ss:   getTestingServiceSettings(true),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectNetwork(ctx, te)
				te.cli.On("ImageInspect", ctx, "wst:test").Return(moby.ImageInspectResult{}, cerrdefs.ErrNotFound)
				pull := &pullReaderCloser{msg: "pulling"}
				te.cli.On("ImagePull", ctx, "wst:test", moby.ImagePullOptions{}).Return(pull, nil)
				t.Cleanup(func() {
					assert.Equal(t, 1, pull.closed)
				})
				expectCreateAndStart(ctx, te, getTestingCreateOptions(true, phpArgs))
				expectRunning(ctx, te)
				collector, logs := expectOutput(t, te)
				return expectedTask(true, "php", collector, logs)
			},
			expectTaskSaved: true,
		},
		{
			name: "successful run replacing existing container and waiting for start",
			ss:   getTestingServiceSettings(true),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectNetwork(ctx, te)
				expectImage(ctx, te)
				opts := getTestingCreateOptions(true, phpArgs)
				te.cli.On("ContainerCreate", ctx, opts).Return(moby.ContainerCreateResult{}, cerrdefs.ErrConflict).Once()
				te.cli.On("ContainerRemove", ctx, "wst-svc", moby.ContainerRemoveOptions{Force: true}).
					Return(moby.ContainerRemoveResult{}, nil)
				te.cli.On("ContainerCreate", ctx, opts).Return(moby.ContainerCreateResult{ID: "cid"}, nil).Once()
				te.cli.On("ContainerStart", ctx, "cid", moby.ContainerStartOptions{}).Return(moby.ContainerStartResult{}, nil)
				te.cli.On("ContainerInspect", ctx, "cid", inspectOpts).
					Return(inspectResult(container.StateCreated, false, 0), nil).Once()
				te.cli.On("ContainerInspect", ctx, "cid", inspectOpts).
					Return(inspectResult(container.StateRunning, true, 0), nil).Once()
				collector, logs := expectOutput(t, te)
				return expectedTask(true, "php", collector, logs)
			},
			expectTaskSaved: true,
		},
		{
			name:   "successful public dry run",
			dryRun: true,
			ss:     getTestingServiceSettings(true),
			cmd:    phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				dt := expectedTask(true, "php", nil, nil)
				dt.containerId = "container"
				return dt
			},
			expectTaskSaved: true,
		},
		{
			name: "failed run due to missing container config",
			ss: func() *environment.ServiceSettings {
				ss := getTestingServiceSettings(true)
				ss.ContainerConfig = nil
				return ss
			}(),
			cmd: phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				return nil
			},
			expectError:    true,
			expectedErrMsg: "container config is not set",
		},
		{
			name: "failed run due to invalid server port with default network name",
			ss: func() *environment.ServiceSettings {
				ss := getTestingServiceSettings(true)
				ss.ServerPort = -1
				return ss
			}(),
			cmd: phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				te.env.namePrefix = ""
				expectNetwork(ctx, te)
				expectImage(ctx, te)
				return nil
			},
			expectError:    true,
			expectedErrMsg: "invalid server port -1 for service svc",
		},
		{
			name: "failed run due to network creation error",
			ss:   getTestingServiceSettings(true),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				te.cli.On("NetworkCreate", ctx, "wst", networkCreateOpts).Return(moby.NetworkCreateResult{}, errors.New("net err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to create network wst - net err",
		},
		{
			name:        "failed run due to image inspect error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				te.cli.On("ImageInspect", ctx, "wst:test").Return(moby.ImageInspectResult{}, errors.New("inspect err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to inspect Docker image wst:test - inspect err",
		},
		{
			name:        "failed run due to image pull error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				te.cli.On("ImageInspect", ctx, "wst:test").Return(moby.ImageInspectResult{}, cerrdefs.ErrNotFound)
				te.cli.On("ImagePull", ctx, "wst:test", moby.ImagePullOptions{}).Return(nil, errors.New("pull err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to pull Docker image wst:test - pull err",
		},
		{
			name:        "failed run due to image pull wait error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				te.cli.On("ImageInspect", ctx, "wst:test").Return(moby.ImageInspectResult{}, cerrdefs.ErrNotFound)
				pull := &pullReaderCloser{waitErr: errors.New("wait err")}
				te.cli.On("ImagePull", ctx, "wst:test", moby.ImagePullOptions{}).Return(pull, nil)
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to finish pulling Docker image wst:test - wait err",
		},
		{
			name:        "failed run due to missing config workspace path",
			networkName: "wst",
			ss: func() *environment.ServiceSettings {
				ss := getTestingServiceSettings(true)
				ss.WorkspaceConfigPaths = map[string]string{}
				return ss
			}(),
			cmd: phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to bind config main_conf for service svc",
		},
		{
			name:        "failed run due to missing script workspace path",
			networkName: "wst",
			ss: func() *environment.ServiceSettings {
				ss := getTestingServiceSettings(true)
				ss.WorkspaceScriptPaths = map[string]string{}
				return ss
			}(),
			cmd: phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to bind script test_php for service svc",
		},
		{
			name:        "failed run due to existing container removal error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				te.cli.On("ContainerCreate", ctx, getTestingCreateOptions(true, phpArgs)).
					Return(moby.ContainerCreateResult{}, cerrdefs.ErrConflict)
				te.cli.On("ContainerRemove", ctx, "wst-svc", moby.ContainerRemoveOptions{Force: true}).
					Return(moby.ContainerRemoveResult{}, errors.New("rm err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to remove existing container wst-svc: rm err",
		},
		{
			name:        "failed run due to container creation error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				te.cli.On("ContainerCreate", ctx, getTestingCreateOptions(true, phpArgs)).
					Return(moby.ContainerCreateResult{}, errors.New("create err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to create Docker container wst-svc: create err",
		},
		{
			name:        "failed run due to container start error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				te.cli.On("ContainerCreate", ctx, getTestingCreateOptions(true, phpArgs)).
					Return(moby.ContainerCreateResult{ID: "cid"}, nil)
				te.cli.On("ContainerStart", ctx, "cid", moby.ContainerStartOptions{}).
					Return(moby.ContainerStartResult{}, errors.New("start err"))
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to start Docker container wst-svc cid: start err",
			expectTaskSaved: true,
		},
		{
			name:        "failed run due to container inspect error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(true, phpArgs))
				te.cli.On("ContainerInspect", ctx, "cid", inspectOpts).
					Return(moby.ContainerInspectResult{}, errors.New("inspect err"))
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed checking of container wst-svc cid readiness: failed to inspect container: inspect err",
			expectTaskSaved: true,
		},
		{
			name:        "failed run due to missing container state",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(true, phpArgs))
				te.cli.On("ContainerInspect", ctx, "cid", inspectOpts).Return(moby.ContainerInspectResult{
					Container: container.InspectResponse{},
				}, nil)
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "container state is not available",
			expectTaskSaved: true,
		},
		{
			name:        "failed run due to exited container",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(true, phpArgs))
				te.cli.On("ContainerInspect", ctx, "cid", inspectOpts).
					Return(inspectResult(container.StateExited, false, 3), nil)
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "container wst-svc cid exited with code 3: state err",
			expectTaskSaved: true,
		},
		{
			name:        "failed run due to timeout while waiting for container",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(true, phpArgs))
				te.cli.On("ContainerInspect", ctx, "cid", inspectOpts).
					Return(inspectResult(container.StateCreated, false, 0), nil).Run(func(args mock.Arguments) {
					cancel()
				})
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "timed out waiting for container wst-svc cid to be ready",
			expectTaskSaved: true,
		},
		{
			name:        "failed run due to logs error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(true, phpArgs))
				expectRunning(ctx, te)
				te.cli.On("ContainerLogs", mock.Anything, "cid", logsOpts).Return(nil, errors.New("logs err"))
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to get container wst-svc logs: logs err",
			expectTaskSaved: true,
		},
		{
			name:        "failed run due to output collector start error",
			networkName: "wst",
			ss:          getTestingServiceSettings(true),
			cmd:         phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *dockerTask {
				expectImage(ctx, te)
				expectCreateAndStart(ctx, te, getTestingCreateOptions(true, phpArgs))
				expectRunning(ctx, te)
				logs := newLogsReaderCloser("out", "")
				te.cli.On("ContainerLogs", mock.Anything, "cid", logsOpts).Return(logs, nil)
				collector := outputMocks.NewMockCollector(t)
				collector.On("Start", mock.Anything, mock.Anything).Return(errors.New("start err"))
				te.om.On("MakeCollector", "cid").Return(collector)
				t.Cleanup(func() {
					assert.Equal(t, 1, logs.closed)
				})
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to start output collection of container wst-svc: start err",
			expectTaskSaved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEnv(t, tt.networkName, nil)
			te.fnd.On("DryRun").Maybe().Return(tt.dryRun)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			expectedTask := tt.setupMocks(t, ctx, cancel, te)

			got, err := te.env.RunTask(ctx, tt.ss, tt.cmd)

			if tt.expectError {
				require.Nil(t, got)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				actualTask, ok := got.(*dockerTask)
				require.True(t, ok)
				assert.Equal(t, expectedTask, actualTask)
			}
			if tt.expectTaskSaved {
				assert.Contains(t, te.env.tasks, "mysvc")
			} else {
				assert.Empty(t, te.env.tasks)
			}
		})
	}
}

// newExecAttachResult creates a hijacked response whose reader provides the given multiplexed content.
func newExecAttachResult(t *testing.T, stdout, stderr string, raw []byte) moby.ExecAttachResult {
	serverConn, clientConn := net.Pipe()
	go func() {
		defer serverConn.Close()
		if raw != nil {
			_, _ = serverConn.Write(raw)
			return
		}
		if stdout != "" {
			_, _ = serverConn.Write(muxFrame(muxStdout, stdout))
		}
		if stderr != "" {
			_, _ = serverConn.Write(muxFrame(muxStderr, stderr))
		}
	}()
	return moby.ExecAttachResult{
		HijackedResponse: moby.NewHijackedResponse(clientConn, "application/vnd.docker.multiplexed-stream"),
	}
}

func Test_dockerEnvironment_ExecTaskCommand(t *testing.T) {
	ss := getTestingServiceSettings(false)
	cmd := &environment.Command{Name: "php", Args: []string{"-v"}}
	execCreateOpts := moby.ExecCreateOptions{
		Cmd:          []string{"php", "-v"},
		AttachStdout: true,
		AttachStderr: true,
	}
	tsk := &dockerTask{containerName: "wst-svc", containerId: "cid"}
	tests := []struct {
		name           string
		dryRun         bool
		target         task.Task
		withCollector  bool
		setupMocks     func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func()
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:          "successful command execution with collector",
			target:        tsk,
			withCollector: true,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}
				oc.On("StdoutWriter").Return(stdout)
				oc.On("StderrWriter").Return(stderr)
				te.cli.On("ExecCreate", ctx, "cid", execCreateOpts).Return(moby.ExecCreateResult{ID: "eid"}, nil)
				te.cli.On("ExecAttach", ctx, "eid", moby.ExecAttachOptions{}).
					Return(newExecAttachResult(t, "out data", "err data", nil), nil)
				te.cli.On("ExecInspect", ctx, "eid", moby.ExecInspectOptions{}).
					Return(moby.ExecInspectResult{ExitCode: 0}, nil)
				return func() {
					assert.Equal(t, "out data", stdout.String())
					assert.Equal(t, "err data", stderr.String())
				}
			},
		},
		{
			name:   "successful command execution without collector",
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				te.cli.On("ExecCreate", ctx, "cid", execCreateOpts).Return(moby.ExecCreateResult{ID: "eid"}, nil)
				te.cli.On("ExecAttach", ctx, "eid", moby.ExecAttachOptions{}).
					Return(newExecAttachResult(t, "out", "", nil), nil)
				te.cli.On("ExecInspect", ctx, "eid", moby.ExecInspectOptions{}).
					Return(moby.ExecInspectResult{ExitCode: 0}, nil)
				return nil
			},
		},
		{
			name:   "successful command execution in dry run",
			dryRun: true,
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				return nil
			},
		},
		{
			name:   "failed command execution due to exec create error",
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				te.cli.On("ExecCreate", ctx, "cid", execCreateOpts).Return(moby.ExecCreateResult{}, errors.New("create err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to create exec for command php in container wst-svc: create err",
		},
		{
			name:   "failed command execution due to exec attach error",
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				te.cli.On("ExecCreate", ctx, "cid", execCreateOpts).Return(moby.ExecCreateResult{ID: "eid"}, nil)
				te.cli.On("ExecAttach", ctx, "eid", moby.ExecAttachOptions{}).
					Return(moby.ExecAttachResult{}, errors.New("attach err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to attach exec for command php in container wst-svc: attach err",
		},
		{
			name:   "failed command execution due to invalid output stream",
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				te.cli.On("ExecCreate", ctx, "cid", execCreateOpts).Return(moby.ExecCreateResult{ID: "eid"}, nil)
				te.cli.On("ExecAttach", ctx, "eid", moby.ExecAttachOptions{}).
					Return(newExecAttachResult(t, "", "", []byte{9, 0, 0, 0, 0, 0, 0, 1, 'x'}), nil)
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to read output of command php in container wst-svc",
		},
		{
			name:   "failed command execution due to exec inspect error",
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				te.cli.On("ExecCreate", ctx, "cid", execCreateOpts).Return(moby.ExecCreateResult{ID: "eid"}, nil)
				te.cli.On("ExecAttach", ctx, "eid", moby.ExecAttachOptions{}).
					Return(newExecAttachResult(t, "out", "", nil), nil)
				te.cli.On("ExecInspect", ctx, "eid", moby.ExecInspectOptions{}).
					Return(moby.ExecInspectResult{}, errors.New("inspect err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to inspect exec for command php in container wst-svc: inspect err",
		},
		{
			name:   "failed command execution due to non zero exit code",
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				te.cli.On("ExecCreate", ctx, "cid", execCreateOpts).Return(moby.ExecCreateResult{ID: "eid"}, nil)
				te.cli.On("ExecAttach", ctx, "eid", moby.ExecAttachOptions{}).
					Return(newExecAttachResult(t, "out", "", nil), nil)
				te.cli.On("ExecInspect", ctx, "eid", moby.ExecInspectOptions{}).
					Return(moby.ExecInspectResult{ExitCode: 2}, nil)
				return nil
			},
			expectError:    true,
			expectedErrMsg: "command php in container wst-svc exited with code 2",
		},
		{
			name:   "failed command execution due to invalid task",
			target: &invalidTask{},
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) func() {
				return nil
			},
			expectError:    true,
			expectedErrMsg: "task in not a Docker task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEnv(t, "wst", nil)
			te.fnd.On("DryRun").Maybe().Return(tt.dryRun)
			ctx := context.Background()
			var oc output.Collector
			var ocMock *outputMocks.MockCollector
			if tt.withCollector {
				ocMock = outputMocks.NewMockCollector(t)
				oc = ocMock
			}
			verify := tt.setupMocks(t, ctx, te, ocMock)

			err := te.env.ExecTaskCommand(ctx, ss, tt.target, cmd, oc)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if verify != nil {
				verify()
			}
		})
	}
}

func Test_dockerEnvironment_ExecTaskSignal(t *testing.T) {
	ss := getTestingServiceSettings(false)
	tsk := &dockerTask{containerName: "wst-svc", containerId: "cid"}
	tests := []struct {
		name           string
		dryRun         bool
		signal         os.Signal
		target         task.Task
		setupMocks     func(t *testing.T, ctx context.Context, te *testEnv)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:   "successful signal",
			signal: syscall.SIGUSR2,
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.cli.On("ContainerKill", ctx, "cid", moby.ContainerKillOptions{Signal: "USR2"}).
					Return(moby.ContainerKillResult{}, nil)
			},
		},
		{
			name:       "successful signal in dry run",
			dryRun:     true,
			signal:     syscall.SIGTERM,
			target:     tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {},
		},
		{
			name:   "failed signal due to kill error",
			signal: syscall.SIGTERM,
			target: tsk,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.cli.On("ContainerKill", ctx, "cid", moby.ContainerKillOptions{Signal: "TERM"}).
					Return(moby.ContainerKillResult{}, errors.New("kill err"))
			},
			expectError:    true,
			expectedErrMsg: "failed to send signal TERM to container wst-svc: kill err",
		},
		{
			name:           "failed signal due to unsupported signal",
			signal:         syscall.SIGALRM,
			target:         tsk,
			setupMocks:     func(t *testing.T, ctx context.Context, te *testEnv) {},
			expectError:    true,
			expectedErrMsg: "signal alarm clock is not supported in Docker environment",
		},
		{
			name:           "failed signal due to invalid task",
			signal:         syscall.SIGTERM,
			target:         &invalidTask{},
			setupMocks:     func(t *testing.T, ctx context.Context, te *testEnv) {},
			expectError:    true,
			expectedErrMsg: "task in not a Docker task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEnv(t, "wst", nil)
			te.fnd.On("DryRun").Maybe().Return(tt.dryRun)
			ctx := context.Background()
			tt.setupMocks(t, ctx, te)

			err := te.env.ExecTaskSignal(ctx, ss, tt.target, tt.signal)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_dockerEnvironment_Output(t *testing.T) {
	tests := []struct {
		name           string
		dryRun         bool
		outputType     output.Type
		target         task.Task
		setupMocks     func(t *testing.T, ctx context.Context, oc *outputMocks.MockCollector) io.Reader
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:       "successful output for stdout type",
			outputType: output.Stdout,
			target:     &dockerTask{containerName: "wst-svc"},
			setupMocks: func(t *testing.T, ctx context.Context, oc *outputMocks.MockCollector) io.Reader {
				reader := strings.NewReader("data")
				oc.On("Reader", ctx, output.Stdout).Return(reader, nil)
				return reader
			},
		},
		{
			name:       "successful output for dry run",
			dryRun:     true,
			outputType: output.Any,
			target:     &dockerTask{containerName: "wst-svc"},
			setupMocks: func(t *testing.T, ctx context.Context, oc *outputMocks.MockCollector) io.Reader {
				return &app.DummyReaderCloser{}
			},
		},
		{
			name:           "failed output due to missing collector",
			outputType:     output.Any,
			target:         &dockerTask{containerName: "wst-svc"},
			expectError:    true,
			expectedErrMsg: "output of container wst-svc is not available",
		},
		{
			name:           "failed output due to invalid task",
			outputType:     output.Any,
			target:         &invalidTask{},
			expectError:    true,
			expectedErrMsg: "task in not a Docker task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEnv(t, "wst", nil)
			te.fnd.On("DryRun").Maybe().Return(tt.dryRun)
			ctx := context.Background()
			var expectedReader io.Reader
			if tt.setupMocks != nil {
				oc := outputMocks.NewMockCollector(t)
				if dockTask, ok := tt.target.(*dockerTask); ok && !tt.dryRun {
					dockTask.outputCollector = oc
				}
				expectedReader = tt.setupMocks(t, ctx, oc)
			}

			reader, err := te.env.Output(ctx, tt.target, tt.outputType)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, reader)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, expectedReader, reader)
			}
		})
	}
}

func Test_dockerEnvironment_RootPath(t *testing.T) {
	env := &dockerEnvironment{}
	assert.Equal(t, "", env.RootPath("/www/ws"))
}

func Test_dockerEnvironment_Mkdir(t *testing.T) {
	env := &dockerEnvironment{}
	assert.Nil(t, env.Mkdir("svc", "/www/ws", 0755))
}

func Test_dockerEnvironment_ServiceLocalAddress(t *testing.T) {
	env := &dockerEnvironment{}
	assert.Equal(t, "0.0.0.0:80", env.ServiceLocalAddress("svc", 1234, 80))
}

func Test_dockerEnvironment_ServiceLocalPort(t *testing.T) {
	env := &dockerEnvironment{}
	assert.Equal(t, int32(80), env.ServiceLocalPort(1234, 80))
}

func Test_dockerEnvironment_ServicePrivateAddress(t *testing.T) {
	env := &dockerEnvironment{}
	assert.Equal(t, "svc", env.ServicePrivateAddress("svc", 1234, 80))
}

func getTestTask() *dockerTask {
	return &dockerTask{
		containerName:       "dn",
		containerId:         "did",
		containerExecutable: "epk",
		containerReady:      true,
		containerPublicUrl:  "://localhost:1234",
		containerPrivateUrl: "://dn:8080",
	}
}

func Test_dockerTask_Id(t *testing.T) {
	assert.Equal(t, "did", getTestTask().Id())
}

func Test_dockerTask_Name(t *testing.T) {
	assert.Equal(t, "dn", getTestTask().Name())
}

func Test_dockerTask_Executable(t *testing.T) {
	assert.Equal(t, "epk", getTestTask().Executable())
}

func Test_dockerTask_Pid(t *testing.T) {
	assert.Equal(t, 1, getTestTask().Pid())
}

func Test_dockerTask_PrivateUrl(t *testing.T) {
	assert.Equal(t, "http://dn:8080", getTestTask().PrivateUrl("http"))
}

func Test_dockerTask_PublicUrl(t *testing.T) {
	assert.Equal(t, "https://localhost:1234", getTestTask().PublicUrl("https"))
}

func Test_dockerTask_Type(t *testing.T) {
	assert.Equal(t, providers.DockerType, getTestTask().Type())
}
