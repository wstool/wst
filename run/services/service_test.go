package services

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	environmentMocks "github.com/wstool/wst/mocks/generated/run/environments/environment"
	outputMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/output"
	taskMocks "github.com/wstool/wst/mocks/generated/run/environments/task"
	parametersMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	certificatesMocks "github.com/wstool/wst/mocks/generated/run/resources/certificates"
	scriptsMocks "github.com/wstool/wst/mocks/generated/run/resources/scripts"
	hooksMocks "github.com/wstool/wst/mocks/generated/run/sandboxes/hooks"
	sandboxMocks "github.com/wstool/wst/mocks/generated/run/sandboxes/sandbox"
	serversMocks "github.com/wstool/wst/mocks/generated/run/servers"
	configsMocks "github.com/wstool/wst/mocks/generated/run/servers/configs"
	templatesMocks "github.com/wstool/wst/mocks/generated/run/servers/templates"
	templateMocks "github.com/wstool/wst/mocks/generated/run/services/template"
	"github.com/wstool/wst/run/environments"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/environments/task"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/resources"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/resources/scripts"
	"github.com/wstool/wst/run/sandboxes/containers"
	"github.com/wstool/wst/run/sandboxes/dir"
	"github.com/wstool/wst/run/sandboxes/hooks"
	"github.com/wstool/wst/run/servers"
	"github.com/wstool/wst/run/servers/templates"
	"github.com/wstool/wst/run/services/template"
	"github.com/wstool/wst/run/spec/defaults"
	"io"
	"os"
	"testing"
)

func TestServices_FindService(t *testing.T) {
	testSvc := &nativeService{name: "svc"}
	services := Services{
		"svc": testSvc,
	}

	tests := []struct {
		name          string
		serviceName   string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Service found",
			serviceName: "svc",
			expectError: false,
		},
		{
			name:          "Service not found",
			serviceName:   "UnknownService",
			expectError:   true,
			expectedError: "service UnknownService not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := services.FindService(tt.serviceName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, svc)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
				assert.Equal(t, testSvc, svc)
			}
		})
	}
}

func TestServices_AddService(t *testing.T) {
	s1 := &nativeService{name: "s1"}
	s2 := &nativeService{name: "s2"}
	services := Services{
		"s1": s1,
	}
	services.AddService(s2)
	assert.Equal(t, Services{"s1": s1, "s2": s2}, services)
}

func Test_nativeServiceLocator_Services(t *testing.T) {
	testSvc := &nativeService{name: "svc"}
	services := Services{
		"svc": testSvc,
	}
	locator := NewServiceLocator(services)
	assert.Equal(t, services, locator.Services())
}

func Test_nativeServiceLocator_Find(t *testing.T) {
	testSvc := &nativeService{name: "svc"}
	services := Services{
		"svc": testSvc,
	}
	locator := NewServiceLocator(services)
	svc, err := locator.Find("svc")
	assert.NoError(t, err)
	assert.Equal(t, testSvc, svc)
}

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	fndMock.TestData().Set("id", "fnd")
	parametersMakerMock := parametersMocks.NewMockMaker(t)
	parametersMakerMock.TestData().Set("id", "pm")
	m := CreateMaker(fndMock, parametersMakerMock)
	require.NotNil(t, m)
	nm, ok := m.(*nativeMaker)
	assert.True(t, ok)
	assert.Equal(t, fndMock, nm.fnd)
	assert.Equal(t, parametersMakerMock, nm.parametersMaker)
	assert.NotNil(t, nm.templateMaker)
}

var paramMocks = make(map[string]*parameterMocks.MockParameter)

func createParamMock(t *testing.T, val string) *parameterMocks.MockParameter {
	paramMock, found := paramMocks[val]
	if !found {
		paramMock = parameterMocks.NewMockParameter(t)
		paramMock.TestData().Set("value", val)
		paramMocks[val] = paramMock
	}
	return paramMock
}

func createCertificateMock(t *testing.T, name string) *certificatesMocks.MockCertificate {
	cert := certificatesMocks.NewMockCertificate(t)
	cert.TestData().Set("cid", name)
	return cert
}

func createScriptMock(t *testing.T, id string) *scriptsMocks.MockScript {
	scriptMock := scriptsMocks.NewMockScript(t)
	scriptMock.TestData().Set("sid", id)
	return scriptMock
}

func createTemplateMock(t *testing.T, id string) *templateMocks.MockTemplate {
	templateMock := templateMocks.NewMockTemplate(t)
	templateMock.TestData().Set("tid", id)
	return templateMock
}

func createConfigMock(t *testing.T, id string) *configsMocks.MockConfig {
	configMock := configsMocks.NewMockConfig(t)
	configMock.TestData().Set("cid", id)
	return configMock
}

func Test_nativeMaker_Make(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	tests := []struct {
		name           string
		config         map[string]types.Service
		defaults       defaults.Defaults
		instanceName   string
		instanceIdx    int
		instanceWs     string
		instanceParams parameters.Parameters
		setupMocks     func(
			t *testing.T,
			pm *parametersMocks.MockMaker,
			tm *templateMocks.MockMaker,
		) (environments.Environments, servers.Servers, *resources.Resources, Services)
		expectedService  func() *nativeService
		verifyResult     func(t *testing.T, locator ServiceLocator, err error)
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful multi services creation",
			config: map[string]types.Service{
				"fpm": {
					Server: types.ServiceServer{
						Name:    "fpm/debian",
						Sandbox: "docker",
						Configs: map[string]types.ServiceConfig{
							"test.conf": {
								Include: false,
							},
							"php.ini": {
								Parameters: types.Parameters{
									"memory_limit": "1G",
								},
								Include: true,
							},
							"fpm.conf": {
								Parameters: types.Parameters{
									"max_children": "10",
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"tag":  "prod",
							"type": "php",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeAll:  true,
							IncludeList: nil,
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{"ssl-cert"},
						},
					},
					Public: false,
				},
				"nginx": {
					Server: types.ServiceServer{
						Name:    "nginx/debian",
						Sandbox: "docker",
						Configs: map[string]types.ServiceConfig{
							"nginx.conf": {
								Parameters: types.Parameters{
									"worker_connections": 1024,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"tag":  "prod",
							"type": "ws",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeAll: true,
						},
						Certificates: types.ServiceResource{
							IncludeAll: true,
						},
					},
					Public: true,
				},
			},
			instanceName: "ti",
			instanceWs:   "/test/workspace",
			instanceParams: parameters.Parameters{
				"ip": createParamMock(t, "ip"),
			},
			instanceIdx: 1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				dockerEnv.On("MarkUsed").Return()
				dockerEnv.On("Resources").Return(&resources.Resources{})
				dockerEnv.On("ReservePort").Return(int32(8500))
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}

				// Parameter mocks
				fpmServerParams := parameters.Parameters{
					"tag":  createParamMock(t, "prod"),
					"type": createParamMock(t, "php"),
				}
				fpmConfConfigParams := parameters.Parameters{
					"max_children": createParamMock(t, "10"),
				}
				fpmPhpIniConfigParams := parameters.Parameters{
					"memory_limit": createParamMock(t, "1G"),
				}
				nginxServerParams := parameters.Parameters{
					"tag":  createParamMock(t, "prod"),
					"type": createParamMock(t, "ws"),
				}
				nginxConfConfigParams := parameters.Parameters{
					"worker_connections": createParamMock(t, "1024"),
				}

				pm.On("Make", types.Parameters{
					"tag":  "prod",
					"type": "php",
				}).Return(fpmServerParams, nil)
				pm.On("Make", types.Parameters{
					"max_children": "10",
				}).Return(fpmConfConfigParams, nil)
				pm.On("Make", types.Parameters{
					"memory_limit": "1G",
				}).Return(fpmPhpIniConfigParams, nil)
				pm.On("Make", types.Parameters{
					"tag":  "prod",
					"type": "ws",
				}).Return(nginxServerParams, nil)
				pm.On("Make", types.Parameters{
					"worker_connections": 1024,
				}).Return(nginxConfConfigParams, nil)

				// Sandbox mock
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(true)

				// Config mocks
				fpmPhpIniConfig := createConfigMock(t, "fpm-php-ini")
				fpmPhpIniConfig.On("Parameters").Return(parameters.Parameters{})
				fpmConfConfig := createConfigMock(t, "fpm-conf")
				fpmConfConfig.On("Parameters").Return(parameters.Parameters{
					"pm": createParamMock(t, "static"),
				})
				nginxConfConfig := createConfigMock(t, "nginx-conf")
				nginxConfConfig.On("Parameters").Return(parameters.Parameters{
					"error_log_level":    createParamMock(t, "debug"),
					"worker_connections": createParamMock(t, "100"),
				})

				// Template mocks
				fpmTemplates := templates.Templates{
					"fpm_conf": templatesMocks.NewMockTemplate(t),
				}
				nginxTemplates := templates.Templates{
					"ngx_conf": templatesMocks.NewMockTemplate(t),
				}

				// Server mocks
				fpmDebSrv := serversMocks.NewMockServer(t)
				fpmDebSrv.On("Sandbox", providers.DockerType).Return(sb, true)
				fpmDebSrv.On("Config", "php.ini").Return(fpmPhpIniConfig, true)
				fpmDebSrv.On("Config", "fpm.conf").Return(fpmConfConfig, true)
				fpmDebSrv.On("Templates").Return(fpmTemplates)
				fpmDebSrv.On("Parameters").Return(parameters.Parameters{
					"fpm_binary": createParamMock(t, "php-fpm"),
				})

				nginxDebSrv := serversMocks.NewMockServer(t)
				nginxDebSrv.On("Sandbox", providers.DockerType).Return(sb, true)
				nginxDebSrv.On("Config", "nginx.conf").Return(nginxConfConfig, true)
				nginxDebSrv.On("Templates").Return(nginxTemplates)
				nginxDebSrv.On("Parameters").Return(parameters.Parameters{
					"nginx_binary": createParamMock(t, "nginx"),
				})

				srvs := servers.Servers{
					"fpm": {
						"debian": fpmDebSrv,
					},
					"nginx": {
						"debian": nginxDebSrv,
					},
				}

				// Scripts and certificates - create once and reuse
				scrs := scripts.Scripts{
					"index.php": scriptsMocks.NewMockScript(t),
				}

				// Create shared certificate instances
				sslCert := createCertificateMock(t, "ssl")
				authCert := createCertificateMock(t, "auth")
				certs := certificates.Certificates{
					"ssl-cert":  sslCert,
					"auth-cert": authCert,
				}

				// Template maker mocks
				fpmTemplate := createTemplateMock(t, "fpm")
				nginxTemplate := createTemplateMock(t, "nginx")

				// Setup template maker expectations with proper types
				tm.On("Make",
					mock.MatchedBy(func(svc *nativeService) bool { return svc.name == "fpm" }),
					mock.MatchedBy(func(services template.Services) bool {
						return len(services) == 2 && services["fpm"] != nil && services["nginx"] != nil
					}),
					fpmTemplates,
				).Return(fpmTemplate)

				tm.On("Make",
					mock.MatchedBy(func(svc *nativeService) bool { return svc.name == "nginx" }),
					mock.MatchedBy(func(services template.Services) bool {
						return len(services) == 2 && services["fpm"] != nil && services["nginx"] != nil
					}),
					nginxTemplates,
				).Return(nginxTemplate)

				// Return nil for expected services since we'll verify differently
				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			// Custom verification function instead of expectedService
			verifyResult: func(t *testing.T, locator ServiceLocator, err error) {
				require.NoError(t, err)
				require.NotNil(t, locator)

				services := locator.Services()
				require.Equal(t, 2, len(services))

				// Verify FPM service (note: service name is "fpm", not "fpm")
				fpmSvc, exists := services["fpm"]
				require.True(t, exists)
				assert.Equal(t, "fpm", fpmSvc.Name())
				assert.Equal(t, "ti-fpm", fpmSvc.FullName())
				assert.Equal(t, false, fpmSvc.IsPublic())
				assert.Equal(t, "/test/workspace/fpm", fpmSvc.Workspace())

				// Verify certificates for fpm (should only have ssl-cert)
				fpmCerts := fpmSvc.(*nativeService).certificates
				assert.Equal(t, 1, len(fpmCerts))
				assert.Contains(t, fpmCerts, "ssl-cert")

				// Verify Nginx service
				nginxSvc, exists := services["nginx"]
				require.True(t, exists)
				assert.Equal(t, "nginx", nginxSvc.Name())
				assert.Equal(t, "ti-nginx", nginxSvc.FullName())
				assert.Equal(t, true, nginxSvc.IsPublic())
				assert.Equal(t, "/test/workspace/nginx", nginxSvc.Workspace())

				// Verify certificates for nginx (should have all certificates)
				nginxCerts := nginxSvc.(*nativeService).certificates
				assert.Equal(t, 2, len(nginxCerts))
				assert.Contains(t, nginxCerts, "ssl-cert")
				assert.Contains(t, nginxCerts, "auth-cert")
			},
			expectError: false,
		},
		{
			name: "successful single service creation with defaults used",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name: "php",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			defaults: defaults.Defaults{
				Service: defaults.ServiceDefaults{
					Sandbox: "local",
					Server: defaults.ServiceServerDefaults{
						Tag: "debian",
					},
				},
				Parameters: parameters.Parameters{
					"dp": createParamMock(t, "dp"),
					"p1": createParamMock(t, "dp1"),
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  2,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				localEnv.On("MarkUsed").Return()
				localEnv.On("Resources").Return(&resources.Resources{})
				localEnv.On("ReservePort").Return(int32(8500))
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				serverParams := parameters.Parameters{
					"p1": createParamMock(t, "p1"),
					"p2": createParamMock(t, "p2"),
				}
				configParams := parameters.Parameters{
					"p0": createParamMock(t, "p0"),
					"p1": createParamMock(t, "p1"),
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(serverParams, nil)
				pm.On("Make", types.Parameters{
					"p0": 10,
					"p1": 2,
				}).Return(configParams, nil)
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(true)
				cfg := configsMocks.NewMockConfig(t)
				cfg.On("Parameters").Return(parameters.Parameters{
					"cp": createParamMock(t, "cp"),
				})
				tmpls := templates.Templates{
					"t": templatesMocks.NewMockTemplate(t),
				}
				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(sb, true)
				debSrv.On("Config", "c").Return(cfg, true)
				debSrv.On("Templates").Return(tmpls)
				debSrv.On("Parameters").Return(parameters.Parameters{
					"p1": createParamMock(t, "p1s"),
					"ps": createParamMock(t, "ps"),
				})
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}
				tmpl := templateMocks.NewMockTemplate(t)
				svc := &nativeService{
					fnd:          fndMock,
					name:         "svc",
					fullName:     "testInstance-svc",
					uniqueName:   "i2-svc",
					public:       true,
					port:         int32(8500),
					certificates: make(certificates.Certificates), // Empty certificates
					scripts: scripts.Scripts{
						"s1": scriptsMocks.NewMockScript(t),
					},
					server: debSrv,
					serverParameters: parameters.Parameters{
						"dp": createParamMock(t, "dp"),
						"p1": createParamMock(t, "p1"),
						"p2": createParamMock(t, "p2"),
						"ps": createParamMock(t, "ps"),
					},
					sandbox:     sb,
					task:        nil,
					environment: localEnv,
					configs: map[string]nativeServiceConfig{
						"c": {
							parameters: parameters.Parameters{
								"cp": createParamMock(t, "cp"),
								"dp": createParamMock(t, "dp"),
								"p0": createParamMock(t, "p0"),
								"p1": createParamMock(t, "p1"),
								"p2": createParamMock(t, "p2"),
								"ps": createParamMock(t, "ps"),
							},
							config: cfg,
						},
					},
					environmentConfigPaths: nil,
					workspaceConfigPaths:   nil,
					environmentScriptPaths: nil,
					workspaceScriptPaths:   nil,
					workspace:              "/test/workspace/svc",
					template:               nil,
				}
				tm.On("Make", svc, template.Services{"svc": svc}, tmpls).Return(tmpl)
				finalSvc := *svc
				finalSvc.template = tmpl
				svcs := Services{
					"svc": &finalSvc,
				}
				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, svcs
			},
			expectError: false,
		},
		{
			name: "errors on certificate not found",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeList: []string{"missing-cert"},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				localEnv.On("MarkUsed").Return()
				localEnv.On("Resources").Return(&resources.Resources{})
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				serverParams := parameters.Parameters{
					"p1": createParamMock(t, "p1"),
					"p2": createParamMock(t, "p2"),
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(serverParams, nil)
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(true)
				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(sb, true)
				debSrv.On("Parameters").Return(parameters.Parameters{})
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"existing-cert": createCertificateMock(t, "existing"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "certificates missing-cert not found for service svc",
		},
		{
			name: "errors on config parameters make",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				localEnv.On("MarkUsed").Return()
				localEnv.On("Resources").Return(&resources.Resources{}) // Add missing Resources() expectation
				// Don't add ReservePort() expectation - error should occur before this
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				serverParams := parameters.Parameters{
					"p1": parameterMocks.NewMockParameter(t),
					"p2": parameterMocks.NewMockParameter(t),
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(serverParams, nil)
				pm.On("Make", types.Parameters{
					"p0": 10,
					"p1": 2,
				}).Return(nil, errors.New("config params make error"))
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(true)
				cfg := configsMocks.NewMockConfig(t)
				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(sb, true)
				debSrv.On("Config", "c").Return(cfg, true)
				debSrv.On("Parameters").Return(parameters.Parameters{})
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "config params make error",
		},
		{
			name: "errors on server config not found",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				localEnv.On("MarkUsed").Return()
				localEnv.On("Resources").Return(&resources.Resources{}) // Add missing Resources() expectation
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				serverParams := parameters.Parameters{
					"p1": parameterMocks.NewMockParameter(t),
					"p2": parameterMocks.NewMockParameter(t),
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(serverParams, nil)
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(true)
				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(sb, true)
				debSrv.On("Config", "c").Return(nil, false)
				debSrv.On("Parameters").Return(parameters.Parameters{})
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "server config c not found for service svc",
		},
		{
			name: "errors on environment not found",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				serverParams := parameters.Parameters{
					"p1": parameterMocks.NewMockParameter(t),
					"p2": parameterMocks.NewMockParameter(t),
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(serverParams, nil)
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(true)
				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(sb, true)
				debSrv.On("Parameters").Return(parameters.Parameters{})
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "environment local not found for service svc",
		},
		{
			name: "errors on sandbox not available",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				serverParams := parameters.Parameters{
					"p1": parameterMocks.NewMockParameter(t),
					"p2": parameterMocks.NewMockParameter(t),
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(serverParams, nil)
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(false)
				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(sb, true)
				debSrv.On("Parameters").Return(parameters.Parameters{})
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "sandbox local is not available for service svc",
		},
		{
			name: "errors on sandbox not found",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				serverParams := parameters.Parameters{
					"p1": parameterMocks.NewMockParameter(t),
					"p2": parameterMocks.NewMockParameter(t),
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(serverParams, nil)
				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(nil, false)
				debSrv.On("Parameters").Return(parameters.Parameters{})
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "sandbox local not found for service svc",
		},
		{
			name: "errors on server parameters make fail",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(nil, errors.New("server params make fail"))
				debSrv := serversMocks.NewMockServer(t)
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "server params make fail",
		},
		{
			name: "errors on server not found",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s1"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
				debSrv := serversMocks.NewMockServer(t)
				srvs := servers.Servers{
					"fpm": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "server php/debian not found for service svc",
		},
		{
			name: "errors on script not found",
			config: map[string]types.Service{
				"svc": {
					Server: types.ServiceServer{
						Name:    "php/debian",
						Sandbox: "local",
						Configs: map[string]types.ServiceConfig{
							"c": {
								Parameters: types.Parameters{
									"p0": 10,
									"p1": 2,
								},
								Include: true,
							},
						},
						Parameters: types.Parameters{
							"p1": 1,
							"p2": "data",
						},
					},
					Resources: types.ServiceResources{
						Scripts: types.ServiceResource{
							IncludeList: []string{"s3"},
						},
						Certificates: types.ServiceResource{
							IncludeAll:  false,
							IncludeList: []string{},
						},
					},
					Public: true,
				},
			},
			instanceName: "testInstance",
			instanceWs:   "/test/workspace",
			instanceIdx:  1,
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, *resources.Resources, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				localEnv.On("MarkUsed").Return()
				localEnv.On("Resources").Return(&resources.Resources{})
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}

				// Add server parameter mock expectations
				serverParams := parameters.Parameters{
					"p1": parameterMocks.NewMockParameter(t),
					"p2": parameterMocks.NewMockParameter(t),
				}
				pm.On("Make", types.Parameters{
					"p1": 1,
					"p2": "data",
				}).Return(serverParams, nil)

				// Add sandbox and server mocks
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(true)

				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(sb, true)
				debSrv.On("Parameters").Return(parameters.Parameters{})

				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}

				// Scripts collection missing "s3" - this will cause the error
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				certs := certificates.Certificates{
					"ssl-cert": createCertificateMock(t, "ssl"),
				}

				return envs, srvs, &resources.Resources{
					Scripts:      scrs,
					Certificates: certs,
				}, nil
			},
			expectError:      true,
			expectedErrorMsg: "script s3 not found for service svc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parametersMakerMock := parametersMocks.NewMockMaker(t)
			templateMakerMock := templateMocks.NewMockMaker(t)

			maker := &nativeMaker{
				fnd:             fndMock,
				parametersMaker: parametersMakerMock,
				templateMaker:   templateMakerMock,
			}

			envs, srvs, rscrs, svcs := tt.setupMocks(t, parametersMakerMock, templateMakerMock)

			locator, err := maker.Make(
				tt.config, &tt.defaults, rscrs, srvs, envs, tt.instanceName, tt.instanceIdx, tt.instanceWs,
				tt.instanceParams)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				assert.Nil(t, locator)
			} else if tt.verifyResult != nil {
				// Use custom verification
				tt.verifyResult(t, locator, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, svcs, locator.Services())
			}
		})
	}
}

func testingNativeService(t *testing.T) *nativeService {
	return &nativeService{
		fnd:                    appMocks.NewMockFoundation(t),
		name:                   "svc",
		fullName:               "instance-name-svc",
		public:                 true,
		port:                   8500,
		workspace:              "/tmp/ws/svc",
		server:                 serversMocks.NewMockServer(t),
		serverParameters:       parameters.Parameters{"p1": parameterMocks.NewMockParameter(t)},
		sandbox:                sandboxMocks.NewMockSandbox(t),
		environment:            environmentMocks.NewMockEnvironment(t),
		template:               templateMocks.NewMockTemplate(t),
		environmentConfigPaths: map[string]string{"env": "/path/to/env"},
		workspaceConfigPaths:   map[string]string{"ws": "/path/to/ws"},
		environmentScriptPaths: map[string]string{"script_env": "/path/to/script_env"},
		workspaceScriptPaths:   map[string]string{"script_ws": "/path/to/script_ws"},
		task:                   taskMocks.NewMockTask(t),
	}
}

func testingServiceSettings(s *nativeService) *environment.ServiceSettings {
	s.server.(*serversMocks.MockServer).On("Port").Return(int32(12345))
	cc := &containers.ContainerConfig{
		ImageName:        "test",
		ImageTag:         "1.0",
		RegistryUsername: "usr",
		RegistryPassword: "grp",
	}
	s.sandbox.(*sandboxMocks.MockSandbox).On("ContainerConfig").Return(cc)
	return &environment.ServiceSettings{
		Name:                   s.name,
		FullName:               s.fullName,
		Port:                   int32(8500),
		ServerPort:             int32(12345),
		Public:                 s.public,
		ContainerConfig:        cc,
		ServerParameters:       s.serverParameters,
		EnvironmentConfigPaths: s.environmentConfigPaths,
		EnvironmentScriptPaths: s.environmentScriptPaths,
		WorkspaceConfigPaths:   s.workspaceConfigPaths,
		WorkspaceScriptPaths:   s.workspaceScriptPaths,
	}
}

func Test_nativeService_PrivateAddress(t *testing.T) {
	svc := testingNativeService(t)
	svc.environment.(*environmentMocks.MockEnvironment).On(
		"ServicePrivateAddress",
		"svc",
		int32(8500),
		int32(80),
	).Return("127.0.0.1:8500")
	svc.server.(*serversMocks.MockServer).On("Port").Return(int32(80))
	assert.Equal(t, "127.0.0.1:8500", svc.PrivateAddress())
}

func Test_nativeService_LocalAddress(t *testing.T) {
	svc := testingNativeService(t)
	svc.environment.(*environmentMocks.MockEnvironment).On(
		"ServiceLocalAddress",
		"svc",
		int32(8500),
		int32(80),
	).Return("127.0.0.1:80")
	svc.server.(*serversMocks.MockServer).On("Port").Return(int32(80))
	assert.Equal(t, "127.0.0.1:80", svc.LocalAddress())
}

func Test_nativeService_LocalPort(t *testing.T) {
	svc := testingNativeService(t)
	svc.environment.(*environmentMocks.MockEnvironment).On(
		"ServiceLocalPort",
		int32(8500),
		int32(80),
	).Return(int32(80))
	svc.server.(*serversMocks.MockServer).On("Port").Return(int32(80))
	assert.Equal(t, int32(80), svc.LocalPort())
}

func Test_nativeService_UdsPath(t *testing.T) {
	tests := []struct {
		name         string
		sockNameArg  string
		expectedPath string
		err          error
	}{
		{
			name:         "successful default result for empty sockName",
			expectedPath: "/ws/run/dir/svc/svc.sock",
			sockNameArg:  "__empty__",
			err:          nil,
		},
		{
			name:         "successful default result if sockName is empty string",
			expectedPath: "/ws/run/dir/svc/svc.sock",
			sockNameArg:  "",
			err:          nil,
		},
		{
			name:         "successful custom result",
			sockNameArg:  "custom",
			expectedPath: "/ws/run/dir/svc/custom.sock",
			err:          nil,
		},
		{
			name: "error during mkdir",
			err:  errors.New("mkdir fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			svc.sandbox.(*sandboxMocks.MockSandbox).On("Dirs").Return(testSandboxDirs())
			envMock := svc.environment.(*environmentMocks.MockEnvironment)
			envMock.On("RootPath", "/tmp/ws/svc").Return("/ws")
			envMock.On("Mkdir", "svc", "/ws/run/dir/svc", os.FileMode(0755)).Return(tt.err)
			var result string
			var err error
			if tt.sockNameArg == "__empty__" {
				result, err = svc.UdsPath()
			} else {
				result, err = svc.UdsPath(tt.sockNameArg)
			}
			assert.Equal(t, tt.expectedPath, result)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_nativeService_Port(t *testing.T) {
	svc := testingNativeService(t)
	svc.server.(*serversMocks.MockServer).On("Port").Return(int32(1234))
	assert.Equal(t, int32(1234), svc.Port())
}

func Test_nativeService_IsPublic(t *testing.T) {
	svc := testingNativeService(t)
	assert.True(t, svc.IsPublic())
}

func Test_nativeService_EnvironmentConfigPaths(t *testing.T) {
	svc := testingNativeService(t)
	expected := map[string]string{"env": "/path/to/env"}
	assert.Equal(t, expected, svc.EnvironmentConfigPaths())
}

func Test_nativeService_WorkspaceConfigPaths(t *testing.T) {
	svc := testingNativeService(t)
	expected := map[string]string{"ws": "/path/to/ws"}
	assert.Equal(t, expected, svc.WorkspaceConfigPaths())
}

func Test_nativeService_EnvironmentScriptPaths(t *testing.T) {
	svc := testingNativeService(t)
	expected := map[string]string{"script_env": "/path/to/script_env"}
	assert.Equal(t, expected, svc.EnvironmentScriptPaths())
}

func Test_nativeService_WorkspaceScriptPaths(t *testing.T) {
	svc := testingNativeService(t)
	expected := map[string]string{"script_ws": "/path/to/script_ws"}
	assert.Equal(t, expected, svc.WorkspaceScriptPaths())
}

func Test_nativeService_User(t *testing.T) {
	svc := testingNativeService(t)
	svc.server.(*serversMocks.MockServer).On("User").Return("username")
	assert.Equal(t, "username", svc.User())
}

func Test_nativeService_Group(t *testing.T) {
	svc := testingNativeService(t)
	svc.server.(*serversMocks.MockServer).On("Group").Return("usergroup")
	assert.Equal(t, "usergroup", svc.Group())
}

func testSandboxDirs() map[dir.DirType]string {
	return map[dir.DirType]string{
		dir.ConfDirType:   "/conf/dir",
		dir.RunDirType:    "/run/dir",
		dir.ScriptDirType: "/script/dir",
	}
}

func Test_nativeService_Dirs(t *testing.T) {
	svc := testingNativeService(t)
	sandboxDirs := testSandboxDirs()
	svc.sandbox.(*sandboxMocks.MockSandbox).On("Dirs").Return(sandboxDirs)
	assert.Equal(t, sandboxDirs, svc.Dirs())
}

func testDirsGetters(t *testing.T, expectedDir string, cb func(svc *nativeService) (string, error)) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "successful result",
			err:  nil,
		},
		{
			name: "error during mkdir",
			err:  errors.New("mkdir fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			svc.sandbox.(*sandboxMocks.MockSandbox).On("Dirs").Return(testSandboxDirs())
			envMock := svc.environment.(*environmentMocks.MockEnvironment)
			envMock.On("RootPath", "/tmp/ws/svc").Return("/ws")
			envMock.On("Mkdir", "svc", expectedDir, os.FileMode(0755)).Return(tt.err)
			result, err := cb(svc)
			assert.Equal(t, expectedDir, result)
			assert.Equal(t, err, tt.err)
		})
	}
}

func Test_nativeService_ConfDir(t *testing.T) {
	testDirsGetters(t, "/ws/conf/dir/svc", func(svc *nativeService) (string, error) {
		return svc.ConfDir()
	})
}

func Test_nativeService_RunDir(t *testing.T) {
	testDirsGetters(t, "/ws/run/dir/svc", func(svc *nativeService) (string, error) {
		return svc.RunDir()
	})
}

func Test_nativeService_ScriptDir(t *testing.T) {
	testDirsGetters(t, "/ws/script/dir/svc", func(svc *nativeService) (string, error) {
		return svc.ScriptDir()
	})
}

func Test_nativeService_Server(t *testing.T) {
	svc := testingNativeService(t)
	assert.Equal(t, svc.server, svc.Server())
}

func Test_nativeService_ServerParameters(t *testing.T) {
	svc := testingNativeService(t)
	assert.Equal(t, parameters.Parameters{"p1": parameterMocks.NewMockParameter(t)}, svc.ServerParameters())
}

func Test_nativeService_SetTemplate(t *testing.T) {
	svc := testingNativeService(t)
	mockTemplate := templateMocks.NewMockTemplate(t)
	mockTemplate.TestData().Set("new", true)
	svc.SetTemplate(mockTemplate)
	assert.Equal(t, mockTemplate, svc.template)
}

func Test_nativeService_Workspace(t *testing.T) {
	svc := testingNativeService(t)
	assert.Equal(t, "/tmp/ws/svc", svc.Workspace())
}

func Test_nativeService_OutputReader(t *testing.T) {
	ctx := context.Background()
	outputType := output.Stdout

	tests := []struct {
		name           string
		setupMocks     func(*environmentMocks.MockEnvironment, task.Task) io.Reader
		taskNotSet     bool
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful output scanner creation",
			setupMocks: func(env *environmentMocks.MockEnvironment, tsk task.Task) io.Reader {
				reader := bytes.NewReader([]byte("test output"))
				env.On("Output", ctx, tsk, outputType).Return(reader, nil)
				return reader
			},
			expectError: false,
		},
		{
			name: "error during output fetching",
			setupMocks: func(env *environmentMocks.MockEnvironment, tsk task.Task) io.Reader {
				env.On("Output", ctx, tsk, outputType).Return(nil, errors.New("out err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "out err",
		},
		{
			name:           "error when task not set",
			taskNotSet:     true,
			expectError:    true,
			expectedErrMsg: "service has not started yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			var expectedReader io.Reader
			if tt.taskNotSet {
				svc.task = nil
			} else {
				expectedReader = tt.setupMocks(svc.environment.(*environmentMocks.MockEnvironment), svc.task)
			}

			reader, err := svc.OutputReader(ctx, outputType)

			if tt.expectError {
				assert.Nil(t, reader)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.Equal(t, expectedReader, reader)
				assert.NoError(t, err)
			}
		})
	}
}

func Test_nativeService_ExecCommand(t *testing.T) {
	ctx := context.Background()
	cmd := &environment.Command{
		Name: "test",
		Args: []string{"arg1", "arg2"},
	}

	expectedServerPort := int32(8080)
	expectedContainerConfig := &containers.ContainerConfig{
		ImageName: "test-image",
	}

	tests := []struct {
		name           string
		setupMocks     func(*environmentMocks.MockEnvironment, *serversMocks.MockServer, *sandboxMocks.MockSandbox, task.Task, output.Collector)
		taskNotSet     bool
		withCollector  bool
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful command execution with collector",
			setupMocks: func(env *environmentMocks.MockEnvironment, srv *serversMocks.MockServer, sb *sandboxMocks.MockSandbox, tsk task.Task, oc output.Collector) {
				srv.On("Port").Return(expectedServerPort)
				sb.On("ContainerConfig").Return(expectedContainerConfig)
				env.On("ExecTaskCommand", ctx, mock.MatchedBy(func(s *environment.ServiceSettings) bool {
					return s.ServerPort == expectedServerPort && s.ContainerConfig == expectedContainerConfig
				}), tsk, cmd, oc).Return(nil)

			},
			withCollector: true,
			expectError:   false,
		},
		{
			name: "successful command execution without collector",
			setupMocks: func(env *environmentMocks.MockEnvironment, srv *serversMocks.MockServer, sb *sandboxMocks.MockSandbox, tsk task.Task, oc output.Collector) {
				srv.On("Port").Return(expectedServerPort)
				sb.On("ContainerConfig").Return(expectedContainerConfig)
				env.On("ExecTaskCommand", ctx, mock.MatchedBy(func(s *environment.ServiceSettings) bool {
					return s.ServerPort == expectedServerPort && s.ContainerConfig == expectedContainerConfig
				}), tsk, cmd, nil).Return(nil)

			},
			withCollector: false,
			expectError:   false,
		},
		{
			name: "error during command execution",
			setupMocks: func(env *environmentMocks.MockEnvironment, srv *serversMocks.MockServer, sb *sandboxMocks.MockSandbox, tsk task.Task, oc output.Collector) {
				srv.On("Port").Return(expectedServerPort)
				sb.On("ContainerConfig").Return(expectedContainerConfig)
				env.On("ExecTaskCommand", ctx, mock.Anything, tsk, cmd, oc).Return(errors.New("execution error"))
			},
			withCollector:  true,
			expectError:    true,
			expectedErrMsg: "execution error",
		},
		{
			name:           "error when task not set",
			taskNotSet:     true,
			expectError:    true,
			expectedErrMsg: "service has not started yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)

			// Create server and sandbox mocks
			serverMock := serversMocks.NewMockServer(t)
			sandboxMock := sandboxMocks.NewMockSandbox(t)

			// Set mocks on service
			svc.server = serverMock
			svc.sandbox = sandboxMock

			if tt.taskNotSet {
				svc.task = nil
			}

			var oc output.Collector
			if tt.withCollector {
				oc = outputMocks.NewMockCollector(t)
			}

			if tt.setupMocks != nil {
				tt.setupMocks(svc.environment.(*environmentMocks.MockEnvironment), serverMock, sandboxMock, svc.task, oc)
			}

			err := svc.ExecCommand(ctx, cmd, oc)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all mock expectations
			serverMock.AssertExpectations(t)
			sandboxMock.AssertExpectations(t)
		})
	}
}

func Test_nativeService_Reload(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setupMocks     func(*nativeService, *sandboxMocks.MockSandbox, *hooksMocks.MockHook)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful reload",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.ReloadHookType).Return(hook, nil)
				hook.On(
					"Execute",
					ctx,
					testingServiceSettings(svc),
					svc.template,
					svc.environment,
					svc.task,
				).Return(nil, nil)
			},
			expectError: false,
		},
		{
			name: "hook execution error",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.ReloadHookType).Return(hook, nil)
				hook.On(
					"Execute",
					ctx,
					testingServiceSettings(svc),
					svc.template,
					svc.environment,
					svc.task,
				).Return(nil, errors.New("execute err"))
			},
			expectError:    true,
			expectedErrMsg: "execute err",
		},
		{
			name: "hook retrieval error",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.ReloadHookType).Return(nil, errors.New("hook err"))
			},
			expectError:    true,
			expectedErrMsg: "hook err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			mockSandbox := svc.sandbox.(*sandboxMocks.MockSandbox)
			mockHook := hooksMocks.NewMockHook(t)
			if tt.setupMocks != nil {
				tt.setupMocks(svc, mockSandbox, mockHook)
			}

			err := svc.Reload(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_nativeService_Restart(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setupMocks     func(*nativeService, *sandboxMocks.MockSandbox, *hooksMocks.MockHook)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful restart",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.RestartHookType).Return(hook, nil)
				hook.On(
					"Execute",
					ctx,
					testingServiceSettings(svc),
					svc.template,
					svc.environment,
					svc.task,
				).Return(nil, nil)
			},
			expectError: false,
		},
		{
			name: "hook execution error",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.RestartHookType).Return(hook, nil)
				hook.On(
					"Execute",
					ctx,
					testingServiceSettings(svc),
					svc.template,
					svc.environment,
					svc.task,
				).Return(nil, errors.New("execute err"))
			},
			expectError:    true,
			expectedErrMsg: "execute err",
		},
		{
			name: "hook retrieval error",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.RestartHookType).Return(nil, errors.New("hook err"))
			},
			expectError:    true,
			expectedErrMsg: "hook err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			mockSandbox := svc.sandbox.(*sandboxMocks.MockSandbox)
			mockHook := hooksMocks.NewMockHook(t)
			if tt.setupMocks != nil {
				tt.setupMocks(svc, mockSandbox, mockHook)
			}

			err := svc.Restart(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_nativeService_Start(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(
			*testing.T,
			*nativeService,
			*appMocks.MockFoundation,
			*environmentMocks.MockEnvironment,
			*sandboxMocks.MockSandbox,
			*serversMocks.MockServer,
			*hooksMocks.MockHook,
			*templateMocks.MockTemplate,
		) (*environment.ServiceSettings, *taskMocks.MockTask)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful start",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				memMapFs := afero.NewMemMapFs()
				_ = afero.WriteFile(memMapFs, "/app/fpm.conf", []byte("[global]"), 0644)
				_ = afero.WriteFile(memMapFs, "/app/php.ini", []byte("[php]"), 0644)
				fnd.On("Fs").Return(memMapFs)

				// Setup configs
				fpmConfConfig := configsMocks.NewMockConfig(t)
				fpmConfConfig.On("FilePath").Return("/app/fpm.conf")
				fpmConfConfigParams := parameters.Parameters{
					"max_children": parameterMocks.NewMockParameter(t),
				}
				phpIniConfig := configsMocks.NewMockConfig(t)
				phpIniConfig.On("FilePath").Return("/app/php.ini")
				phpIniConfigParams := parameters.Parameters{
					"memory_limit": parameterMocks.NewMockParameter(t),
				}
				svc.configs = map[string]nativeServiceConfig{
					"fpm_conf": {
						parameters: fpmConfConfigParams,
						config:     fpmConfConfig,
					},
					"php_ini": {
						parameters: phpIniConfigParams,
						config:     phpIniConfig,
					},
				}

				// Setup environment and sandbox for configs
				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.ConfDirType).Return("conf", nil)
				tmpl.On(
					"RenderToFile",
					"[global]",
					fpmConfConfigParams,
					"/tmp/ws/svc/conf/fpm.conf",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"[php]",
					phpIniConfigParams,
					"/tmp/ws/svc/conf/php.ini",
					os.FileMode(0644),
				).Return(nil)

				// Setup certificates
				testCert := certificatesMocks.NewMockCertificate(t)
				testCert.On("CertificateName").Return("test.crt")
				testCert.On("CertificateData").Return("-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----")
				testCert.On("PrivateKeyName").Return("test.key")
				testCert.On("PrivateKeyData").Return("-----BEGIN PRIVATE KEY-----\ntest key data\n-----END PRIVATE KEY-----")
				svc.certificates = certificates.Certificates{
					"test_cert": testCert,
				}

				// Setup sandbox for certificates
				sb.On("Dir", dir.CertDirType).Return("certs", nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/test.crt",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN PRIVATE KEY-----\ntest key data\n-----END PRIVATE KEY-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/test.key",
					os.FileMode(0600),
				).Return(nil)

				// Setup scripts
				indexScript := scriptsMocks.NewMockScript(t)
				indexScript.On("Path").Return("")
				indexScript.On("Mode").Return(os.FileMode(0664))
				indexScript.On("Content").Return("<?php echo 1;")
				indexScriptParams := parameters.Parameters{
					"num": parameterMocks.NewMockParameter(t),
				}
				indexScript.On("Parameters").Return(indexScriptParams)
				svc.scripts = scripts.Scripts{
					"index.php": indexScript,
				}
				sb.On("Dir", dir.ScriptDirType).Return("scr", nil)
				tmpl.On(
					"RenderToFile",
					"<?php echo 1;",
					indexScriptParams,
					"/tmp/ws/svc/scr/index.php",
					os.FileMode(0664),
				).Return(nil)

				// Setup hook execution
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				ss := testingServiceSettings(svc)
				ss.EnvironmentConfigPaths = map[string]string{
					"fpm_conf": "/tmp/svc/conf/svc/fpm.conf",
					"php_ini":  "/tmp/svc/conf/svc/php.ini",
				}
				ss.EnvironmentScriptPaths = map[string]string{
					"index.php": "/tmp/svc/scr/svc/index.php",
				}
				ss.WorkspaceConfigPaths = map[string]string{
					"fpm_conf": "/tmp/ws/svc/conf/fpm.conf",
					"php_ini":  "/tmp/ws/svc/conf/php.ini",
				}
				ss.WorkspaceScriptPaths = map[string]string{
					"index.php": "/tmp/ws/svc/scr/index.php",
				}
				ss.Certificates = map[string]*certificates.RenderedCertificate{
					"test_cert": {
						Certificate:               testCert,
						CertificateFilePath:       "/tmp/svc/certs/svc/test.crt",
						CertificateSourceFilePath: "/tmp/ws/svc/certs/test.crt",
						PrivateKeyFilePath:        "/tmp/svc/certs/svc/test.key",
						PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/test.key",
					},
				}
				tsk := taskMocks.NewMockTask(t)
				tsk.TestData().Set("id", "task")
				hook.On(
					"Execute",
					ctx,
					ss,
					svc.template,
					svc.environment,
					svc.task,
				).Return(tsk, nil)
				return ss, tsk
			},
			expectError: false,
		},
		{
			name: "certificate rendering error - cert file",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				// Setup minimal configs first
				svc.configs = map[string]nativeServiceConfig{}

				// Setup certificates that will fail
				testCert := certificatesMocks.NewMockCertificate(t)
				testCert.On("CertificateName").Return("test.crt")
				testCert.On("CertificateData").Return("-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----")
				svc.certificates = certificates.Certificates{
					"test_cert": testCert,
				}

				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.CertDirType).Return("certs", nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/test.crt",
					os.FileMode(0644),
				).Return(errors.New("cert render fail"))

				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "cert render fail",
		},
		{
			name: "certificate rendering error - private key file",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				// Setup minimal configs first
				svc.configs = map[string]nativeServiceConfig{}

				// Setup certificates that will fail on private key
				testCert := certificatesMocks.NewMockCertificate(t)
				testCert.On("CertificateName").Return("test.crt")
				testCert.On("CertificateData").Return("-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----")
				testCert.On("PrivateKeyName").Return("test.key")
				testCert.On("PrivateKeyData").Return("-----BEGIN PRIVATE KEY-----\ntest key data\n-----END PRIVATE KEY-----")
				svc.certificates = certificates.Certificates{
					"test_cert": testCert,
				}

				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.CertDirType).Return("certs", nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/test.crt",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN PRIVATE KEY-----\ntest key data\n-----END PRIVATE KEY-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/test.key",
					os.FileMode(0600),
				).Return(errors.New("private key render fail"))

				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "private key render fail",
		},
		{
			name: "certificate sandbox dir error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				// Setup minimal configs first
				svc.configs = map[string]nativeServiceConfig{}

				// Setup certificates - need to mock the methods that get called before the error
				testCert := certificatesMocks.NewMockCertificate(t)
				testCert.On("CertificateName").Return("test.crt")
				testCert.On("CertificateData").Return("-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----")
				svc.certificates = certificates.Certificates{
					"test_cert": testCert,
				}

				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.CertDirType).Return("", errors.New("cert dir fail"))
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "cert dir fail",
		},
		{
			name: "hook execution error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				memMapFs := afero.NewMemMapFs()
				_ = afero.WriteFile(memMapFs, "/app/fpm.conf", []byte("[global]"), 0644)
				_ = afero.WriteFile(memMapFs, "/app/php.ini", []byte("[php]"), 0644)
				fnd.On("Fs").Return(memMapFs)

				// Setup configs
				fpmConfConfig := configsMocks.NewMockConfig(t)
				fpmConfConfig.On("FilePath").Return("/app/fpm.conf")
				fpmConfConfigParams := parameters.Parameters{
					"max_children": parameterMocks.NewMockParameter(t),
				}
				phpIniConfig := configsMocks.NewMockConfig(t)
				phpIniConfig.On("FilePath").Return("/app/php.ini")
				phpIniConfigParams := parameters.Parameters{
					"memory_limit": parameterMocks.NewMockParameter(t),
				}
				svc.configs = map[string]nativeServiceConfig{
					"fpm_conf": {
						parameters: fpmConfConfigParams,
						config:     fpmConfConfig,
					},
					"php_ini": {
						parameters: phpIniConfigParams,
						config:     phpIniConfig,
					},
				}
				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.ConfDirType).Return("conf", nil)
				tmpl.On(
					"RenderToFile",
					"[global]",
					fpmConfConfigParams,
					"/tmp/ws/svc/conf/fpm.conf",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"[php]",
					phpIniConfigParams,
					"/tmp/ws/svc/conf/php.ini",
					os.FileMode(0644),
				).Return(nil)

				// Setup empty certificates (no certificates to render)
				svc.certificates = certificates.Certificates{}

				// Setup scripts
				indexScript := scriptsMocks.NewMockScript(t)
				indexScript.On("Path").Return("/app/index.php")
				indexScript.On("Mode").Return(os.FileMode(0664))
				indexScript.On("Content").Return("<?php echo 1;")
				indexScriptParams := parameters.Parameters{
					"num": parameterMocks.NewMockParameter(t),
				}
				indexScript.On("Parameters").Return(indexScriptParams)
				svc.scripts = scripts.Scripts{
					"index": indexScript,
				}
				sb.On("Dir", dir.ScriptDirType).Return("scr", nil)
				tmpl.On(
					"RenderToFile",
					"<?php echo 1;",
					indexScriptParams,
					"/tmp/ws/svc/scr/index.php",
					os.FileMode(0664),
				).Return(nil)

				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				ss := testingServiceSettings(svc)
				ss.EnvironmentConfigPaths = map[string]string{
					"fpm_conf": "/tmp/svc/conf/svc/fpm.conf", // Fixed: was "fpm.conf", should be "fpm.conf"
					"php_ini":  "/tmp/svc/conf/svc/php.ini",
				}
				ss.EnvironmentScriptPaths = map[string]string{
					"index": "/tmp/svc/scr/svc/index.php",
				}
				ss.WorkspaceConfigPaths = map[string]string{
					"fpm_conf": "/tmp/ws/svc/conf/fpm.conf",
					"php_ini":  "/tmp/ws/svc/conf/php.ini",
				}
				ss.WorkspaceScriptPaths = map[string]string{
					"index": "/tmp/ws/svc/scr/index.php",
				}
				ss.Certificates = map[string]*certificates.RenderedCertificate{}
				hook.On(
					"Execute",
					ctx,
					ss,
					svc.template,
					svc.environment,
					svc.task,
				).Return(nil, errors.New("execute fail"))
				return ss, nil
			},
			expectError:    true,
			expectedErrMsg: "execute fail",
		},
		{
			name: "script rendering error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				memMapFs := afero.NewMemMapFs()
				_ = afero.WriteFile(memMapFs, "/app/fpm.conf", []byte("[global]"), 0644)
				_ = afero.WriteFile(memMapFs, "/app/php.ini", []byte("[php]"), 0644)
				fnd.On("Fs").Return(memMapFs)

				// Setup configs
				fpmConfConfig := configsMocks.NewMockConfig(t)
				fpmConfConfig.On("FilePath").Return("/app/fpm.conf")
				fpmConfConfigParams := parameters.Parameters{
					"max_children": parameterMocks.NewMockParameter(t),
				}
				phpIniConfig := configsMocks.NewMockConfig(t)
				phpIniConfig.On("FilePath").Return("/app/php.ini")
				phpIniConfigParams := parameters.Parameters{
					"memory_limit": parameterMocks.NewMockParameter(t),
				}
				svc.configs = map[string]nativeServiceConfig{
					"fpm_conf": {
						parameters: fpmConfConfigParams,
						config:     fpmConfConfig,
					},
					"php_ini": {
						parameters: phpIniConfigParams,
						config:     phpIniConfig,
					},
				}
				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.ConfDirType).Return("conf", nil)
				tmpl.On(
					"RenderToFile",
					"[global]",
					fpmConfConfigParams,
					"/tmp/ws/svc/conf/fpm.conf",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"[php]",
					phpIniConfigParams,
					"/tmp/ws/svc/conf/php.ini",
					os.FileMode(0644),
				).Return(nil)

				// Setup empty certificates (no certificates to render)
				svc.certificates = certificates.Certificates{}

				// Setup scripts that will fail
				indexScript := scriptsMocks.NewMockScript(t)
				indexScript.On("Path").Return("/app/index.php")
				indexScript.On("Mode").Return(os.FileMode(0664))
				indexScript.On("Content").Return("<?php echo 1;")
				indexScriptParams := parameters.Parameters{
					"num": parameterMocks.NewMockParameter(t),
				}
				indexScript.On("Parameters").Return(indexScriptParams)
				svc.scripts = scripts.Scripts{
					"index": indexScript,
				}
				sb.On("Dir", dir.ScriptDirType).Return("scr", nil)
				tmpl.On(
					"RenderToFile",
					"<?php echo 1;",
					indexScriptParams,
					"/tmp/ws/svc/scr/index.php",
					os.FileMode(0664),
				).Return(errors.New("script render fail"))

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "script render fail",
		},
		{
			name: "sandbox script dir error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				memMapFs := afero.NewMemMapFs()
				_ = afero.WriteFile(memMapFs, "/app/fpm.conf", []byte("[global]"), 0644)
				_ = afero.WriteFile(memMapFs, "/app/php.ini", []byte("[php]"), 0644)
				fnd.On("Fs").Return(memMapFs)

				// Setup configs
				fpmConfConfig := configsMocks.NewMockConfig(t)
				fpmConfConfig.On("FilePath").Return("/app/fpm.conf")
				fpmConfConfigParams := parameters.Parameters{
					"max_children": parameterMocks.NewMockParameter(t),
				}
				phpIniConfig := configsMocks.NewMockConfig(t)
				phpIniConfig.On("FilePath").Return("/app/php.ini")
				phpIniConfigParams := parameters.Parameters{
					"memory_limit": parameterMocks.NewMockParameter(t),
				}
				svc.configs = map[string]nativeServiceConfig{
					"fpm_conf": {
						parameters: fpmConfConfigParams,
						config:     fpmConfConfig,
					},
					"php_ini": {
						parameters: phpIniConfigParams,
						config:     phpIniConfig,
					},
				}
				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.ConfDirType).Return("conf", nil)
				tmpl.On(
					"RenderToFile",
					"[global]",
					fpmConfConfigParams,
					"/tmp/ws/svc/conf/fpm.conf",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"[php]",
					phpIniConfigParams,
					"/tmp/ws/svc/conf/php.ini",
					os.FileMode(0644),
				).Return(nil)

				// Setup empty certificates
				svc.certificates = certificates.Certificates{}

				// Setup scripts
				indexScript := scriptsMocks.NewMockScript(t)
				indexScript.On("Path").Return("/app/index.php")
				svc.scripts = scripts.Scripts{
					"index": indexScript,
				}
				sb.On("Dir", dir.ScriptDirType).Return("", errors.New("sandbox script dir fail"))

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "sandbox script dir fail",
		},
		{
			name: "config rendering error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				memMapFs := afero.NewMemMapFs()
				_ = afero.WriteFile(memMapFs, "/app/fpm.conf", []byte("[global]"), 0644)
				_ = afero.WriteFile(memMapFs, "/app/php.ini", []byte("[php]"), 0644)
				fnd.On("Fs").Return(memMapFs)

				// Setup configs
				fpmConfConfig := configsMocks.NewMockConfig(t)
				fpmConfConfig.On("FilePath").Return("/app/fpm.conf")
				fpmConfConfigParams := parameters.Parameters{
					"max_children": parameterMocks.NewMockParameter(t),
				}
				phpIniConfig := configsMocks.NewMockConfig(t)
				phpIniConfig.On("FilePath").Return("/app/php.ini")
				phpIniConfigParams := parameters.Parameters{
					"memory_limit": parameterMocks.NewMockParameter(t),
				}
				svc.configs = map[string]nativeServiceConfig{
					"fpm_conf": {
						parameters: fpmConfConfigParams,
						config:     fpmConfConfig,
					},
					"php_ini": {
						parameters: phpIniConfigParams,
						config:     phpIniConfig,
					},
				}
				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.ConfDirType).Return("conf", nil)
				tmpl.On(
					"RenderToFile",
					"[global]",
					fpmConfConfigParams,
					"/tmp/ws/svc/conf/fpm.conf",
					os.FileMode(0644),
				).Return(nil).Maybe()
				tmpl.On(
					"RenderToFile",
					"[php]",
					phpIniConfigParams,
					"/tmp/ws/svc/conf/php.ini",
					os.FileMode(0644),
				).Return(errors.New("config render fail"))

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "config render fail",
		},
		{
			name: "sandbox config dir error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)

				// Setup configs
				phpIniConfig := configsMocks.NewMockConfig(t)
				phpIniConfig.On("FilePath").Return("/app/php.ini")
				phpIniConfigParams := parameters.Parameters{
					"memory_limit": parameterMocks.NewMockParameter(t),
				}
				svc.configs = map[string]nativeServiceConfig{
					"php_ini": {
						parameters: phpIniConfigParams,
						config:     phpIniConfig,
					},
				}

				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.ConfDirType).Return("conf", errors.New("sandbox config dir fail"))
				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "sandbox config dir fail",
		},
		{
			name: "config file read error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)

				// Setup file system mocks that will fail on read
				mockFile := appMocks.NewMockFile(t)
				mockFile.On("Close").Return(nil)
				mockFile.On("Read", mock.Anything).Return(0, errors.New("config read fail"))
				mockFs := appMocks.NewMockFs(t)
				mockFs.On("Open", "/app/php.ini").Return(mockFile, nil)
				fnd.On("Fs").Return(mockFs)

				// Setup configs
				phpIniConfig := configsMocks.NewMockConfig(t)
				phpIniConfig.On("FilePath").Return("/app/php.ini")
				phpIniConfigParams := parameters.Parameters{
					"memory_limit": parameterMocks.NewMockParameter(t),
				}
				svc.configs = map[string]nativeServiceConfig{
					"php_ini": {
						parameters: phpIniConfigParams,
						config:     phpIniConfig,
					},
				}
				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.ConfDirType).Return("conf", nil)

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "config read fail",
		},
		{
			name: "config file open error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)

				// Setup file system mocks that will fail on open
				mockFs := appMocks.NewMockFs(t)
				mockFs.On("Open", "/app/php.ini").Return(nil, errors.New("config open fail"))
				fnd.On("Fs").Return(mockFs)

				// Setup configs
				phpIniConfig := configsMocks.NewMockConfig(t)
				phpIniConfig.On("FilePath").Return("/app/php.ini")
				phpIniConfigParams := parameters.Parameters{
					"memory_limit": parameterMocks.NewMockParameter(t),
				}
				svc.configs = map[string]nativeServiceConfig{
					"php_ini": {
						parameters: phpIniConfigParams,
						config:     phpIniConfig,
					},
				}
				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.ConfDirType).Return("conf", nil)

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "config open fail",
		},
		{
			name: "hook retrieval error",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				sb.On("Hook", hooks.StartHookType).Return(nil, assert.AnError)
				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: assert.AnError.Error(),
		},
		{
			name: "certificate rendering error - cert file",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				// Setup minimal configs first
				svc.configs = map[string]nativeServiceConfig{}

				// Setup certificates that will fail
				testCert := certificatesMocks.NewMockCertificate(t)
				testCert.On("CertificateName").Return("test.crt")
				testCert.On("CertificateData").Return("-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----")
				svc.certificates = certificates.Certificates{
					"test_cert": testCert,
				}

				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.CertDirType).Return("certs", nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/test.crt",
					os.FileMode(0644),
				).Return(errors.New("cert render fail"))

				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "cert render fail",
		},
		{
			name: "certificate rendering error - private key file",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				// Setup minimal configs first
				svc.configs = map[string]nativeServiceConfig{}

				// Setup certificates that will fail on private key
				testCert := certificatesMocks.NewMockCertificate(t)
				testCert.On("CertificateName").Return("test.crt")
				testCert.On("CertificateData").Return("-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----")
				testCert.On("PrivateKeyName").Return("test.key")
				testCert.On("PrivateKeyData").Return("-----BEGIN PRIVATE KEY-----\ntest key data\n-----END PRIVATE KEY-----")
				svc.certificates = certificates.Certificates{
					"test_cert": testCert,
				}

				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.CertDirType).Return("certs", nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN CERTIFICATE-----\ntest cert data\n-----END CERTIFICATE-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/test.crt",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN PRIVATE KEY-----\ntest key data\n-----END PRIVATE KEY-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/test.key",
					os.FileMode(0600),
				).Return(errors.New("private key render fail"))

				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "private key render fail",
		},
		{
			name: "successful start with empty resources",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				// Setup with no configs, certificates, or scripts
				svc.configs = map[string]nativeServiceConfig{}
				svc.certificates = certificates.Certificates{}
				svc.scripts = scripts.Scripts{}

				// Setup hook execution
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				ss := testingServiceSettings(svc)
				ss.EnvironmentConfigPaths = map[string]string{}
				ss.EnvironmentScriptPaths = map[string]string{}
				ss.WorkspaceConfigPaths = map[string]string{}
				ss.WorkspaceScriptPaths = map[string]string{}
				ss.Certificates = map[string]*certificates.RenderedCertificate{}

				tsk := taskMocks.NewMockTask(t)
				tsk.TestData().Set("id", "task")
				hook.On(
					"Execute",
					ctx,
					ss,
					svc.template,
					svc.environment,
					svc.task,
				).Return(tsk, nil)
				return ss, tsk
			},
			expectError: false,
		},
		{
			name: "successful start with multiple certificates",
			setupMocks: func(
				t *testing.T,
				svc *nativeService,
				fnd *appMocks.MockFoundation,
				env *environmentMocks.MockEnvironment,
				sb *sandboxMocks.MockSandbox,
				sv *serversMocks.MockServer,
				hook *hooksMocks.MockHook,
				tmpl *templateMocks.MockTemplate,
			) (*environment.ServiceSettings, *taskMocks.MockTask) {
				// Setup with no configs or scripts but multiple certificates
				svc.configs = map[string]nativeServiceConfig{}
				svc.scripts = scripts.Scripts{}

				// Setup multiple certificates
				cert1 := certificatesMocks.NewMockCertificate(t)
				cert1.On("CertificateName").Return("server.crt")
				cert1.On("CertificateData").Return("-----BEGIN CERTIFICATE-----\nserver cert\n-----END CERTIFICATE-----")
				cert1.On("PrivateKeyName").Return("server.key")
				cert1.On("PrivateKeyData").Return("-----BEGIN PRIVATE KEY-----\nserver key\n-----END PRIVATE KEY-----")

				cert2 := certificatesMocks.NewMockCertificate(t)
				cert2.On("CertificateName").Return("ca.crt")
				cert2.On("CertificateData").Return("-----BEGIN CERTIFICATE-----\nca cert\n-----END CERTIFICATE-----")
				cert2.On("PrivateKeyName").Return("ca.key")
				cert2.On("PrivateKeyData").Return("-----BEGIN PRIVATE KEY-----\nca key\n-----END PRIVATE KEY-----")

				svc.certificates = certificates.Certificates{
					"server_cert": cert1,
					"ca_cert":     cert2,
				}

				// Setup sandbox for certificates
				env.On("RootPath", "/tmp/ws/svc").Return("/tmp/svc")
				sb.On("Dir", dir.CertDirType).Return("certs", nil)

				// Mock certificate rendering
				tmpl.On(
					"RenderToFile",
					"-----BEGIN CERTIFICATE-----\nserver cert\n-----END CERTIFICATE-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/server.crt",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN PRIVATE KEY-----\nserver key\n-----END PRIVATE KEY-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/server.key",
					os.FileMode(0600),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN CERTIFICATE-----\nca cert\n-----END CERTIFICATE-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/ca.crt",
					os.FileMode(0644),
				).Return(nil)
				tmpl.On(
					"RenderToFile",
					"-----BEGIN PRIVATE KEY-----\nca key\n-----END PRIVATE KEY-----",
					parameters.Parameters{},
					"/tmp/ws/svc/certs/ca.key",
					os.FileMode(0600),
				).Return(nil)

				// Setup hook execution
				sb.On("Hook", hooks.StartHookType).Return(hook, nil)
				ss := testingServiceSettings(svc)
				ss.EnvironmentConfigPaths = map[string]string{}
				ss.EnvironmentScriptPaths = map[string]string{}
				ss.WorkspaceConfigPaths = map[string]string{}
				ss.WorkspaceScriptPaths = map[string]string{}
				ss.Certificates = map[string]*certificates.RenderedCertificate{
					"server_cert": {
						Certificate:               cert1,
						CertificateFilePath:       "/tmp/svc/certs/svc/server.crt",
						CertificateSourceFilePath: "/tmp/ws/svc/certs/server.crt",
						PrivateKeyFilePath:        "/tmp/svc/certs/svc/server.key",
						PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/server.key",
					},
					"ca_cert": {
						Certificate:               cert2,
						CertificateFilePath:       "/tmp/svc/certs/svc/ca.crt",
						CertificateSourceFilePath: "/tmp/ws/svc/certs/ca.crt",
						PrivateKeyFilePath:        "/tmp/svc/certs/svc/ca.key",
						PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/ca.key",
					},
				}

				tsk := taskMocks.NewMockTask(t)
				tsk.TestData().Set("id", "task")
				hook.On(
					"Execute",
					ctx,
					ss,
					svc.template,
					svc.environment,
					svc.task,
				).Return(tsk, nil)
				return ss, tsk
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			svc.environmentConfigPaths = nil
			svc.environmentScriptPaths = nil
			svc.workspaceConfigPaths = nil
			svc.workspaceScriptPaths = nil
			svc.renderedCertificates = nil

			mockFoundation := svc.fnd.(*appMocks.MockFoundation)
			mockEnv := svc.environment.(*environmentMocks.MockEnvironment)
			mockSandbox := svc.sandbox.(*sandboxMocks.MockSandbox)
			mockServer := svc.server.(*serversMocks.MockServer)
			mockTemplate := svc.template.(*templateMocks.MockTemplate)
			mockHook := hooksMocks.NewMockHook(t)
			ss, tsk := tt.setupMocks(t, svc, mockFoundation, mockEnv, mockSandbox, mockServer, mockHook, mockTemplate)

			err := svc.Start(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tsk, svc.task)
				assert.Equal(t, ss.EnvironmentConfigPaths, svc.environmentConfigPaths)
				assert.Equal(t, ss.EnvironmentScriptPaths, svc.environmentScriptPaths)
				assert.Equal(t, ss.WorkspaceConfigPaths, svc.workspaceConfigPaths)
				assert.Equal(t, ss.WorkspaceScriptPaths, svc.workspaceScriptPaths)
				assert.Equal(t, ss.Certificates, svc.renderedCertificates)
			}
		})
	}
}

func Test_nativeService_Stop(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setupMocks     func(*nativeService, *sandboxMocks.MockSandbox, *hooksMocks.MockHook)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful stop",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.StopHookType).Return(hook, nil)
				hook.On(
					"Execute",
					ctx,
					testingServiceSettings(svc),
					svc.template,
					svc.environment,
					svc.task,
				).Return(nil, nil)
			},
			expectError: false,
		},
		{
			name: "hook execution error",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.StopHookType).Return(hook, nil)
				hook.On(
					"Execute",
					ctx,
					testingServiceSettings(svc),
					svc.template,
					svc.environment,
					svc.task,
				).Return(nil, errors.New("execute err"))
			},
			expectError:    true,
			expectedErrMsg: "execute err",
		},
		{
			name: "hook retrieval error",
			setupMocks: func(svc *nativeService, sb *sandboxMocks.MockSandbox, hook *hooksMocks.MockHook) {
				sb.On("Hook", hooks.StopHookType).Return(nil, errors.New("hook err"))
			},
			expectError:    true,
			expectedErrMsg: "hook err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			mockSandbox := svc.sandbox.(*sandboxMocks.MockSandbox)
			mockHook := hooksMocks.NewMockHook(t)
			if tt.setupMocks != nil {
				tt.setupMocks(svc, mockSandbox, mockHook)
			}

			err := svc.Stop(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc.task)
			}
		})
	}
}

func Test_nativeService_Name(t *testing.T) {
	svc := testingNativeService(t)
	assert.Equal(t, "svc", svc.Name())
}

func Test_nativeService_FullName(t *testing.T) {
	svc := testingNativeService(t)
	assert.Equal(t, "instance-name-svc", svc.FullName())
}

func Test_nativeService_Executable(t *testing.T) {
	tests := []struct {
		name               string
		hasTask            bool
		taskExecutable     string
		expectedExecutable string
		expectError        bool
		expectedErrMsg     string
	}{
		{
			name:               "successful executable",
			hasTask:            true,
			taskExecutable:     "ep",
			expectedExecutable: "ep",
			expectError:        false,
		},
		{
			name:           "error on missing task",
			hasTask:        false,
			expectError:    true,
			expectedErrMsg: "service has not started yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			if !tt.hasTask {
				svc.task = nil
			}
			if tt.taskExecutable != "" {
				svc.task.(*taskMocks.MockTask).On("Executable").Return(tt.taskExecutable)
			}

			actualUrl, err := svc.Executable()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedExecutable, actualUrl)
			}
		})
	}
}

func Test_nativeService_PublicUrl(t *testing.T) {
	tests := []struct {
		name           string
		hasTask        bool
		isPublic       bool
		taskUrl        string
		scheme         string
		path           string
		expectedUrl    string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "successful public URL",
			hasTask:     true,
			isPublic:    true,
			taskUrl:     "http://svc",
			scheme:      "http",
			path:        "test",
			expectedUrl: "http://svc/test",
			expectError: false,
		},
		{
			name:           "error on private",
			hasTask:        true,
			isPublic:       false,
			expectError:    true,
			expectedErrMsg: "only public service has public URL",
		},
		{
			name:           "error on missing task",
			hasTask:        false,
			isPublic:       true,
			expectError:    true,
			expectedErrMsg: "service has not started yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			svc.public = tt.isPublic
			if !tt.hasTask {
				svc.task = nil
			}
			if tt.taskUrl != "" {
				svc.task.(*taskMocks.MockTask).On("PublicUrl", tt.scheme).Return(tt.taskUrl)
			}

			actualUrl, err := svc.PublicUrl(tt.scheme, tt.path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUrl, actualUrl)
			}
		})
	}
}

func Test_nativeService_PrivateUrl(t *testing.T) {
	tests := []struct {
		name           string
		hasTask        bool
		scheme         string
		taskUrl        string
		expectedUrl    string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "successful private URL",
			hasTask:     true,
			scheme:      "http",
			taskUrl:     "http://svc",
			expectedUrl: "http://svc",
			expectError: false,
		},
		{
			name:           "error on missing task",
			hasTask:        false,
			expectError:    true,
			expectedErrMsg: "service has not started yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			if !tt.hasTask {
				svc.task = nil
			}
			if tt.taskUrl != "" {
				svc.task.(*taskMocks.MockTask).On("PrivateUrl", tt.scheme).Return(tt.taskUrl)
			}

			actualUrl, err := svc.PrivateUrl(tt.scheme)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUrl, actualUrl)
			}
		})
	}
}

func Test_nativeService_Pid(t *testing.T) {
	tests := []struct {
		name           string
		hasTask        bool
		taskPid        int
		expectedPid    int
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "successful PID",
			hasTask:     true,
			taskPid:     1234,
			expectedPid: 1234,
			expectError: false,
		},
		{
			name:           "error on missing task",
			hasTask:        false,
			expectError:    true,
			expectedErrMsg: "service has not started yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			if !tt.hasTask {
				svc.task = nil
			}
			if tt.taskPid != 0 {
				svc.task.(*taskMocks.MockTask).On("Pid").Return(tt.taskPid)
			}

			actualPid, err := svc.Pid()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPid, actualPid)
			}
		})
	}
}

func Test_nativeService_RenderTemplate(t *testing.T) {
	svc := testingNativeService(t)
	params := parameters.Parameters{
		"p1": parameterMocks.NewMockParameter(t),
	}
	text := "hey {{ .Parameters.GetString \"name\" }}"
	svc.template.(*templateMocks.MockTemplate).On("RenderToString", text, params).Return("hey you", nil)
	result, err := svc.RenderTemplate(text, params)
	assert.NoError(t, err)
	assert.Equal(t, "hey you", result)
}

func Test_nativeService_Sandbox(t *testing.T) {
	svc := testingNativeService(t)
	assert.Equal(t, svc.sandbox, svc.Sandbox())
}

func Test_nativeService_Environment(t *testing.T) {
	svc := testingNativeService(t)
	assert.Equal(t, svc.environment, svc.Environment())
}

func Test_nativeService_Task(t *testing.T) {
	svc := testingNativeService(t)
	assert.Equal(t, svc.task, svc.Task())
}

func Test_nativeService_FindCertificate(t *testing.T) {
	tests := []struct {
		name               string
		certificateName    string
		setupRenderedCerts func(*nativeService)
		expectedCert       *certificates.RenderedCertificate
		expectError        bool
		expectedErrMsg     string
	}{
		{
			name:            "certificate found",
			certificateName: "test_cert",
			setupRenderedCerts: func(svc *nativeService) {
				mockCert := certificatesMocks.NewMockCertificate(t)
				renderedCert := &certificates.RenderedCertificate{
					Certificate:               mockCert,
					CertificateFilePath:       "/tmp/svc/certs/test.crt",
					CertificateSourceFilePath: "/tmp/ws/svc/certs/test.crt",
					PrivateKeyFilePath:        "/tmp/svc/certs/test.key",
					PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/test.key",
				}
				svc.renderedCertificates = map[string]*certificates.RenderedCertificate{
					"test_cert": renderedCert,
				}
			},
			expectedCert: &certificates.RenderedCertificate{
				Certificate:               nil, // Will be set in setup
				CertificateFilePath:       "/tmp/svc/certs/test.crt",
				CertificateSourceFilePath: "/tmp/ws/svc/certs/test.crt",
				PrivateKeyFilePath:        "/tmp/svc/certs/test.key",
				PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/test.key",
			},
			expectError: false,
		},
		{
			name:            "certificate not found",
			certificateName: "nonexistent_cert",
			setupRenderedCerts: func(svc *nativeService) {
				mockCert := certificatesMocks.NewMockCertificate(t)
				renderedCert := &certificates.RenderedCertificate{
					Certificate:               mockCert,
					CertificateFilePath:       "/tmp/svc/certs/test.crt",
					CertificateSourceFilePath: "/tmp/ws/svc/certs/test.crt",
					PrivateKeyFilePath:        "/tmp/svc/certs/test.key",
					PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/test.key",
				}
				svc.renderedCertificates = map[string]*certificates.RenderedCertificate{
					"test_cert": renderedCert,
				}
			},
			expectedCert:   nil,
			expectError:    true,
			expectedErrMsg: "certificate nonexistent_cert not found",
		},
		{
			name:            "empty rendered certificates",
			certificateName: "any_cert",
			setupRenderedCerts: func(svc *nativeService) {
				svc.renderedCertificates = map[string]*certificates.RenderedCertificate{}
			},
			expectedCert:   nil,
			expectError:    true,
			expectedErrMsg: "certificate any_cert not found",
		},
		{
			name:            "nil rendered certificates map",
			certificateName: "any_cert",
			setupRenderedCerts: func(svc *nativeService) {
				svc.renderedCertificates = nil
			},
			expectedCert:   nil,
			expectError:    true,
			expectedErrMsg: "certificate any_cert not found",
		},
		{
			name:            "multiple certificates - find specific one",
			certificateName: "server_cert",
			setupRenderedCerts: func(svc *nativeService) {
				mockCert1 := certificatesMocks.NewMockCertificate(t)
				mockCert2 := certificatesMocks.NewMockCertificate(t)

				svc.renderedCertificates = map[string]*certificates.RenderedCertificate{
					"ca_cert": {
						Certificate:               mockCert1,
						CertificateFilePath:       "/tmp/svc/certs/ca.crt",
						CertificateSourceFilePath: "/tmp/ws/svc/certs/ca.crt",
						PrivateKeyFilePath:        "/tmp/svc/certs/ca.key",
						PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/ca.key",
					},
					"server_cert": {
						Certificate:               mockCert2,
						CertificateFilePath:       "/tmp/svc/certs/server.crt",
						CertificateSourceFilePath: "/tmp/ws/svc/certs/server.crt",
						PrivateKeyFilePath:        "/tmp/svc/certs/server.key",
						PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/server.key",
					},
				}
			},
			expectedCert: &certificates.RenderedCertificate{
				Certificate:               nil, // Will be set in setup
				CertificateFilePath:       "/tmp/svc/certs/server.crt",
				CertificateSourceFilePath: "/tmp/ws/svc/certs/server.crt",
				PrivateKeyFilePath:        "/tmp/svc/certs/server.key",
				PrivateKeySourceFilePath:  "/tmp/ws/svc/certs/server.key",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testingNativeService(t)
			tt.setupRenderedCerts(svc)

			result, err := svc.FindCertificate(tt.certificateName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Compare all fields except Certificate (since it's a mock)
				assert.Equal(t, tt.expectedCert.CertificateFilePath, result.CertificateFilePath)
				assert.Equal(t, tt.expectedCert.CertificateSourceFilePath, result.CertificateSourceFilePath)
				assert.Equal(t, tt.expectedCert.PrivateKeyFilePath, result.PrivateKeyFilePath)
				assert.Equal(t, tt.expectedCert.PrivateKeySourceFilePath, result.PrivateKeySourceFilePath)
				assert.NotNil(t, result.Certificate) // Just verify the Certificate field is not nil
			}
		})
	}
}
