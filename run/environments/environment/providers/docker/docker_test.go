package docker

import (
	"context"
	apitypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	dockerClientMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/providers/docker/client"
	resourcesMocks "github.com/wstool/wst/mocks/generated/run/resources"
	certificatesMocks "github.com/wstool/wst/mocks/generated/run/resources/certificates"
	scriptsMocks "github.com/wstool/wst/mocks/generated/run/resources/scripts"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/resources"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/resources/scripts"
	"github.com/wstool/wst/run/sandboxes/containers"
	"io"
	"os"
	"testing"
	"time"
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
					waitTickDuration: 1 * time.Second,
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
					waitTickDuration: 1 * time.Second,
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

func Test_dockerEnvironment_Destroy(t *testing.T) {
	tests := []struct {
		name             string
		networkName      string
		tasks            map[string]*dockerTask
		setupMocks       func(*testing.T, context.Context, *appMocks.MockFoundation, *dockerClientMocks.MockClient)
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:        "successful docker destroying",
			networkName: "net1",
			tasks: map[string]*dockerTask{
				"t1": {
					containerName: "cn1",
					containerId:   "cid1",
				},
				"t2": {
					containerName: "cn2",
					containerId:   "cid2",
				},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				cli.On("ContainerStop", ctx, "cid1", container.StopOptions{}).Return(nil)
				cli.On("ContainerRemove", ctx, "cid1", container.RemoveOptions{}).Return(nil)
				cli.On("ContainerStop", ctx, "cid2", container.StopOptions{}).Return(nil)
				cli.On("ContainerRemove", ctx, "cid2", container.RemoveOptions{}).Return(nil)
				cli.On("NetworkRemove", ctx, "net1").Return(nil)
			},
		},
		{
			name:        "successful docker destroying with dry run",
			networkName: "net1",
			tasks: map[string]*dockerTask{
				"t1": {
					containerName: "cn1",
					containerId:   "cid1",
				},
				"t2": {
					containerName: "cn2",
					containerId:   "cid2",
				},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(true)
			},
		},
		{
			name:        "failed docker destroying with single error",
			networkName: "net1",
			tasks: map[string]*dockerTask{
				"t1": {
					containerName: "cn1",
					containerId:   "cid1",
				},
				"t2": {
					containerName: "cn2",
					containerId:   "cid2",
				},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				cli.On("ContainerStop", ctx, "cid1", container.StopOptions{}).Return(nil)
				cli.On("ContainerRemove", ctx, "cid1", container.RemoveOptions{}).Return(nil)
				cli.On("ContainerStop", ctx, "cid2", container.StopOptions{}).Return(nil)
				cli.On("ContainerRemove", ctx, "cid2", container.RemoveOptions{}).Return(nil)
				cli.On("NetworkRemove", ctx, "net1").Return(errors.New("failed net"))
			},
			expectError:      true,
			expectedErrorMsg: "Destroying docker environment failed",
		},
		{
			name:        "failed docker destroying with more errors",
			networkName: "net1",
			tasks: map[string]*dockerTask{
				"t1": {
					containerName: "cn1",
					containerId:   "cid1",
				},
				"t2": {
					containerName: "cn2",
					containerId:   "cid2",
				},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				cli.On("ContainerStop", ctx, "cid1", container.StopOptions{}).Return(errors.New("cs1"))
				cli.On("ContainerStop", ctx, "cid2", container.StopOptions{}).Return(nil)
				cli.On("ContainerRemove", ctx, "cid2", container.RemoveOptions{}).Return(errors.New("cr2"))
				cli.On("NetworkRemove", ctx, "net1").Return(errors.New("failed net"))
			},
			expectError:      true,
			expectedErrorMsg: "Destroying docker environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			clientMock := dockerClientMocks.NewMockClient(t)
			ctx := context.Background()
			e := &dockerEnvironment{
				ContainerEnvironment: environment.ContainerEnvironment{
					CommonEnvironment: environment.CommonEnvironment{
						Fnd: fndMock,
					},
					Registry: environment.ContainerRegistry{},
				},
				cli:         clientMock,
				networkName: tt.networkName,
				tasks:       tt.tasks,
			}

			tt.setupMocks(t, ctx, fndMock, clientMock)

			err := e.Destroy(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type pullReaderCloser struct {
	msg string
	err string
}

func (b *pullReaderCloser) Read(p []byte) (n int, err error) {
	if len(b.err) > 0 {
		return 0, errors.New(b.err)
	}
	if len(b.msg) > 0 {
		n = copy(p, b.msg)
		b.msg = b.msg[n:]
		return n, nil
	}
	return 0, io.EOF
}

func (b *pullReaderCloser) Close() error {
	return nil
}

func Test_dockerEnvironment_RunTask(t *testing.T) {
	tests := []struct {
		name          string
		envNamePrefix string
		envStartPort  int32
		networkName   string
		ss            *environment.ServiceSettings
		cmd           *environment.Command
		setupMocks    func(
			*testing.T,
			context.Context,
			context.CancelFunc,
			*appMocks.MockFoundation,
			*dockerClientMocks.MockClient,
		)
		contextSetup     func() (context.Context, context.CancelFunc)
		expectedTask     *dockerTask
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:          "successful docker public run without network and certs",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				cli.On("NetworkCreate", ctx, "wt", network.CreateOptions{
					Driver: "bridge",
				}).Return(network.CreateResponse{}, nil)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
						PortBindings: nat.PortMap{
							"1234/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
						},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				waitResp := container.WaitResponse{}
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				go func() {
					defer close(statusCh)
					defer close(errCh)
					statusCh <- waitResp
				}()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{
					ContainerJSONBase: &apitypes.ContainerJSONBase{
						State: &apitypes.ContainerState{
							Running: true,
						},
					},
				}, nil)
			},
			expectedTask: &dockerTask{
				containerName:       "wt-svc",
				containerId:         "dcid",
				containerExecutable: "php",
				containerReady:      true,
				containerPublicUrl:  "://localhost:8080",
				containerPrivateUrl: "://wt-svc:1234",
			},
		},
		{
			name:          "successful docker run with certificates",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
				Certificates: map[string]*certificates.RenderedCertificate{
					"server_cert": {
						Certificate:               certificatesMocks.NewMockCertificate(t),
						CertificateSourceFilePath: "/tmp/wst/ssl/server.crt",
						CertificateFilePath:       "/etc/ssl/server.crt",
						PrivateKeySourceFilePath:  "/tmp/wst/ssl/server.key",
						PrivateKeyFilePath:        "/etc/ssl/server.key",
					},
					"ca_cert": {
						// Test certificate with only certificate file (no private key)
						Certificate:               certificatesMocks.NewMockCertificate(t),
						CertificateSourceFilePath: "/tmp/wst/ssl/ca.crt",
						CertificateFilePath:       "/etc/ssl/ca.crt",
						PrivateKeySourceFilePath:  "", // Empty to test the conditional
						PrivateKeyFilePath:        "",
					},
					"client_key": {
						// Test certificate with only private key file (no certificate)
						Certificate:               certificatesMocks.NewMockCertificate(t),
						CertificateSourceFilePath: "",
						CertificateFilePath:       "",
						PrivateKeySourceFilePath:  "/tmp/wst/ssl/client.key",
						PrivateKeyFilePath:        "/etc/ssl/client.key",
					},
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					mock.MatchedBy(func(hostConfig *container.HostConfig) bool {
						expectedBinds := []string{
							"/tmp/wst/main.conf:/etc/main.conf",
							"/tmp/wst/test.php:/www/test.php",
							"/tmp/wst/ssl/server.crt:/etc/ssl/server.crt",
							"/tmp/wst/ssl/server.key:/etc/ssl/server.key",
							"/tmp/wst/ssl/ca.crt:/etc/ssl/ca.crt",
							"/tmp/wst/ssl/client.key:/etc/ssl/client.key",
						}

						// Check if all expected binds are present
						if len(hostConfig.Binds) != len(expectedBinds) {
							return false
						}

						// Convert to maps for easy comparison
						actualBindsMap := make(map[string]bool)
						for _, bind := range hostConfig.Binds {
							actualBindsMap[bind] = true
						}

						for _, expectedBind := range expectedBinds {
							if !actualBindsMap[expectedBind] {
								return false
							}
						}

						return true
					}),
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				waitResp := container.WaitResponse{}
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				go func() {
					defer close(statusCh)
					defer close(errCh)
					statusCh <- waitResp
				}()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{
					ContainerJSONBase: &apitypes.ContainerJSONBase{
						State: &apitypes.ContainerState{
							Running: true,
						},
					},
				}, nil)
			},
			expectedTask: &dockerTask{
				containerName:       "wt-svc",
				containerId:         "dcid",
				containerExecutable: "php",
				containerReady:      true,
				containerPublicUrl:  "",
				containerPrivateUrl: "://wt-svc:1234",
			},
		},

		// Also add a test case with certificates that have empty paths to ensure full coverage
		{
			name:          "successful docker run with empty certificate paths",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{},
				EnvironmentScriptPaths: map[string]string{},
				WorkspaceConfigPaths:   map[string]string{},
				WorkspaceScriptPaths:   map[string]string{},
				Certificates: map[string]*certificates.RenderedCertificate{
					"empty_cert": {
						Certificate: certificatesMocks.NewMockCertificate(t),
						// Test certificate with empty paths - should not add any binds
						CertificateSourceFilePath: "",
						CertificateFilePath:       "",
						PrivateKeySourceFilePath:  "",
						PrivateKeyFilePath:        "",
					},
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{}, // Empty binds since all certificate paths are empty
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				waitResp := container.WaitResponse{}
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				go func() {
					defer close(statusCh)
					defer close(errCh)
					statusCh <- waitResp
				}()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{
					ContainerJSONBase: &apitypes.ContainerJSONBase{
						State: &apitypes.ContainerState{
							Running: true,
						},
					},
				}, nil)
			},
			expectedTask: &dockerTask{
				containerName:       "wt-svc",
				containerId:         "dcid",
				containerExecutable: "php",
				containerReady:      true,
				containerPublicUrl:  "",
				containerPrivateUrl: "://wt-svc:1234",
			},
		},
		{
			name:          "successful docker private run with network",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				waitResp := container.WaitResponse{}
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				go func() {
					defer close(statusCh)
					defer close(errCh)
					statusCh <- waitResp
				}()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{
					ContainerJSONBase: &apitypes.ContainerJSONBase{
						State: &apitypes.ContainerState{
							Running: false,
						},
					},
				}, nil).Once()
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{
					ContainerJSONBase: &apitypes.ContainerJSONBase{
						State: &apitypes.ContainerState{
							Running: true,
						},
					},
				}, nil)
			},
			expectedTask: &dockerTask{
				containerName:       "wt-svc",
				containerId:         "dcid",
				containerExecutable: "php",
				containerReady:      true,
				containerPublicUrl:  "",
				containerPrivateUrl: "://wt-svc:1234",
			},
		},
		{
			name:          "successful docker private dry run",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(true)
			},
			expectedTask: &dockerTask{
				containerName:       "wt-svc",
				containerId:         "container",
				containerExecutable: "php",
				containerReady:      true,
				containerPublicUrl:  "",
				containerPrivateUrl: "://wt-svc:1234",
			},
		},
		{
			name:          "failed docker run on the second inspect",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				waitResp := container.WaitResponse{}
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				go func() {
					defer close(statusCh)
					defer close(errCh)
					statusCh <- waitResp
				}()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{
					ContainerJSONBase: &apitypes.ContainerJSONBase{
						State: &apitypes.ContainerState{
							Running: false,
						},
					},
				}, nil).Once()
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{}, errors.New("ci2"))
			},
			expectError:      true,
			expectedErrorMsg: "failed checking of container wt-svc dcid readiness: failed to inspect container: ci2",
		},
		{
			name:          "failed docker run on the second wait due to context being done",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				waitResp := container.WaitResponse{}
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				go func() {
					defer close(statusCh)
					defer close(errCh)
					statusCh <- waitResp
				}()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{
					ContainerJSONBase: &apitypes.ContainerJSONBase{
						State: &apitypes.ContainerState{
							Running: false,
						},
					},
				}, nil).Once()
				cli.On("ContainerInspect", ctx, "dcid").Return(
					apitypes.ContainerJSON{
						ContainerJSONBase: &apitypes.ContainerJSONBase{
							State: &apitypes.ContainerState{
								Running: false,
							},
						},
					},
					nil,
				).Run(func(args mock.Arguments) {
					cancel()
				})
			},
			expectError:      true,
			expectedErrorMsg: "timed out waiting for container wt-svc dcid to be ready",
		},
		{
			name:          "failed docker run on the first inspect",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				waitResp := container.WaitResponse{}
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				go func() {
					defer close(statusCh)
					defer close(errCh)
					statusCh <- waitResp
				}()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
				cli.On("ContainerInspect", ctx, "dcid").Return(apitypes.ContainerJSON{}, errors.New("ci2"))
			},
			expectError:      true,
			expectedErrorMsg: "failed checking of container wt-svc dcid readiness: failed to inspect container: ci2",
		},
		{
			name:          "failed docker run on the first wait due to context being done",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				cancel()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
			},
			expectError:      true,
			expectedErrorMsg: "timed out waiting for container wt-svc dcid to be ready",
		},
		{
			name:          "failed docker run on container wait",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(nil)
				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error)
				go func() {
					defer close(statusCh)
					defer close(errCh)
					errCh <- errors.New("wait err")
				}()
				cli.On("ContainerWait", ctx, "dcid", container.WaitConditionNotRunning).Return(
					(<-chan container.WaitResponse)(statusCh),
					(<-chan error)(errCh),
				)
			},
			expectError:      true,
			expectedErrorMsg: "failed waiting on container wt-svc dcid to run: wait err",
		},
		{
			name:          "failed docker run on container start with success container remove",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(errors.New("start err"))
				cli.On("ContainerRemove", ctx, "dcid", container.RemoveOptions{}).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "failed to start Docker container wt-svc dcid: start err",
		},
		{
			name:          "failed docker run on container start with failed container remove",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{ID: "dcid"}, nil)
				cli.On("ContainerStart", ctx, "dcid", container.StartOptions{}).Return(errors.New("start err"))
				cli.On("ContainerRemove", ctx, "dcid", container.RemoveOptions{}).Return(errors.New("failed rem"))
			},
			expectError:      true,
			expectedErrorMsg: "failed to start Docker container wt-svc dcid: start err",
		},
		{
			name:          "failed docker run on container create",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
				var platform *v1.Platform = nil
				cli.On(
					"ContainerCreate",
					ctx,
					&container.Config{
						Image: "wst:test",
						Cmd:   []string{"php", "test.php", "run"},
					},
					&container.HostConfig{
						Binds: []string{"/tmp/wst/main.conf:/etc/main.conf", "/tmp/wst/test.php:/www/test.php"},
					},
					&network.NetworkingConfig{
						EndpointsConfig: map[string]*network.EndpointSettings{
							"wt": {},
						},
					},
					platform,
					"wt-svc",
				).Return(container.CreateResponse{}, errors.New("create err"))
			},
			expectError:      true,
			expectedErrorMsg: "failed to create Docker container for service svc: create err",
		},
		{
			name:          "failed docker run on unmatched script",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_x": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
			},
			expectError:      true,
			expectedErrorMsg: "failed to bind script test_php for service svc",
		},
		{
			name:          "failed docker run on unmatched script",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"main_cx": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				pullOut := &pullReaderCloser{}
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(pullOut, nil)
			},
			expectError:      true,
			expectedErrorMsg: "failed to bind config main_conf for service svc",
		},
		{
			name:          "failed docker run on failed pull",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "wt",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				cli.On("ImagePull", ctx, "wst:test", image.PullOptions{}).Return(nil, errors.New("pull err"))
			},
			expectError:      true,
			expectedErrorMsg: "failed to pull Docker image wst:test - pull err",
		},
		{
			name:          "failed docker run on network create",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "",
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
				fnd.On("DryRun").Return(false)
				cli.On("NetworkCreate", ctx, "wt", network.CreateOptions{
					Driver: "bridge",
				}).Return(network.CreateResponse{}, errors.New("net create err"))
			},
			expectError:      true,
			expectedErrorMsg: "failed to create network wt - net create err",
		},

		{
			name:          "failed docker run on not set container config",
			envNamePrefix: "wt",
			envStartPort:  8080,
			networkName:   "",
			ss: &environment.ServiceSettings{
				Name:            "svc",
				FullName:        "mysvc",
				Port:            8080,
				ServerPort:      1234,
				Public:          false,
				ContainerConfig: nil,
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
					"test_php": "/tmp/wst/test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) {
			},
			expectError:      true,
			expectedErrorMsg: "container config is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			clientMock := dockerClientMocks.NewMockClient(t)
			var ctx context.Context
			var cancel context.CancelFunc
			if tt.contextSetup == nil {
				ctx, cancel = context.WithCancel(context.Background())
			} else {
				ctx, cancel = tt.contextSetup()
			}
			e := &dockerEnvironment{
				ContainerEnvironment: environment.ContainerEnvironment{
					CommonEnvironment: environment.CommonEnvironment{
						Fnd: fndMock,
						Ports: environment.Ports{
							Start: tt.envStartPort,
							Used:  tt.envStartPort,
							End:   tt.envStartPort + 100,
						},
					},
					Registry: environment.ContainerRegistry{},
				},
				cli:              clientMock,
				networkName:      tt.networkName,
				namePrefix:       tt.envNamePrefix,
				tasks:            make(map[string]*dockerTask),
				waitTickDuration: 10 * time.Millisecond,
			}

			tt.setupMocks(t, ctx, cancel, fndMock, clientMock)

			got, err := e.RunTask(ctx, tt.ss, tt.cmd)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				actualTask, ok := got.(*dockerTask)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedTask, actualTask)
			}
		})
	}
}

func Test_dockerEnvironment_ExecTaskCommand(t *testing.T) {
	env := &dockerEnvironment{}
	ctx := context.Background()
	ss := &environment.ServiceSettings{
		Name:       "svc",
		FullName:   "mysvc",
		Port:       8080,
		ServerPort: 1234,
	}
	cmd := &environment.Command{Name: "test"}
	tsk := &dockerTask{}
	oc := &output.BufferedCollector{}
	err := env.ExecTaskCommand(ctx, ss, tsk, cmd, oc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executing command is not currently supported in Docker environment")
}

func Test_dockerEnvironment_ExecTaskSignal(t *testing.T) {
	env := &dockerEnvironment{}
	ctx := context.Background()
	ss := &environment.ServiceSettings{
		Name:       "svc",
		FullName:   "mysvc",
		Port:       8080,
		ServerPort: 1234,
	}
	tsk := &dockerTask{}
	err := env.ExecTaskSignal(ctx, ss, tsk, os.Interrupt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executing signal is not currently supported in Kubernetes environment")
}

func Test_dockerEnvironment_Output(t *testing.T) {
	tests := []struct {
		name       string
		outputType output.Type
		target     *dockerTask
		setupMocks func(
			*testing.T,
			context.Context,
			*appMocks.MockFoundation,
			*dockerClientMocks.MockClient,
		) io.Reader
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:       "successful output for any type",
			outputType: output.Any,
			target: &dockerTask{
				containerName: "cn1",
				containerId:   "cid1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) io.Reader {
				reader := &pullReaderCloser{
					msg: "data",
				}
				fnd.On("DryRun").Return(false)
				cli.On("ContainerLogs", ctx, "cid1", container.LogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Follow:     true,
				}).Return(reader, nil)
				return reader
			},
		},
		{
			name:       "successful output for stdout type",
			outputType: output.Stdout,
			target: &dockerTask{
				containerName: "cn1",
				containerId:   "cid1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) io.Reader {
				reader := &pullReaderCloser{
					msg: "data",
				}
				fnd.On("DryRun").Return(false)
				cli.On("ContainerLogs", ctx, "cid1", container.LogsOptions{
					ShowStdout: true,
					ShowStderr: false,
					Follow:     true,
				}).Return(reader, nil)
				return reader
			},
		},
		{
			name:       "successful output for stderr type",
			outputType: output.Stderr,
			target: &dockerTask{
				containerName: "cn1",
				containerId:   "cid1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) io.Reader {
				reader := &pullReaderCloser{
					msg: "data",
				}
				fnd.On("DryRun").Return(false)
				cli.On("ContainerLogs", ctx, "cid1", container.LogsOptions{
					ShowStdout: false,
					ShowStderr: true,
					Follow:     true,
				}).Return(reader, nil)
				return reader
			},
		},

		{
			name:       "successful output with dry run",
			outputType: output.Stderr,
			target: &dockerTask{
				containerName: "cn1",
				containerId:   "cid1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) io.Reader {
				fnd.On("DryRun").Return(true)
				return &app.DummyReaderCloser{}
			},
		},
		{
			name:       "failed output on container logs",
			outputType: output.Stderr,
			target: &dockerTask{
				containerName: "cn1",
				containerId:   "cid1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cli *dockerClientMocks.MockClient,
			) io.Reader {
				fnd.On("DryRun").Return(false)
				cli.On("ContainerLogs", ctx, "cid1", container.LogsOptions{
					ShowStdout: false,
					ShowStderr: true,
					Follow:     true,
				}).Return(nil, errors.New("log err"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to get container logs: log err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			clientMock := dockerClientMocks.NewMockClient(t)
			ctx := context.Background()
			e := &dockerEnvironment{
				ContainerEnvironment: environment.ContainerEnvironment{
					CommonEnvironment: environment.CommonEnvironment{
						Fnd: fndMock,
					},
					Registry: environment.ContainerRegistry{},
				},
				cli: clientMock,
			}

			expectedReader := tt.setupMocks(t, ctx, fndMock, clientMock)
			actualReader, err := e.Output(ctx, tt.target, tt.outputType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedReader, actualReader)
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
	assert.Equal(t, "127.0.0.1:80", env.ServiceLocalAddress("svc", 1234, 80))
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
		containerName:       "cn",
		containerId:         "cid",
		containerExecutable: "epd",
		containerReady:      true,
		containerPublicUrl:  "://localhost:8080",
		containerPrivateUrl: "://cn:1234",
	}
}

func Test_dockerTask_Id(t *testing.T) {
	assert.Equal(t, "cid", getTestTask().Id())
}

func Test_dockerTask_Name(t *testing.T) {
	assert.Equal(t, "cn", getTestTask().Name())
}

func Test_dockerTask_Executable(t *testing.T) {
	assert.Equal(t, "epd", getTestTask().Executable())
}

func Test_dockerTask_Pid(t *testing.T) {
	assert.Equal(t, 1, getTestTask().Pid())

}

func Test_dockerTask_PrivateUrl(t *testing.T) {
	assert.Equal(t, "http://cn:1234", getTestTask().PrivateUrl("http"))
}

func Test_dockerTask_PublicUrl(t *testing.T) {
	assert.Equal(t, "https://localhost:8080", getTestTask().PublicUrl("https"))
}

func Test_dockerTask_Type(t *testing.T) {
	assert.Equal(t, providers.DockerType, getTestTask().Type())
}
