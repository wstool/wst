package services

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	environmentMocks "github.com/bukka/wst/mocks/generated/run/environments/environment"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	scriptsMocks "github.com/bukka/wst/mocks/generated/run/resources/scripts"
	sandboxMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/sandbox"
	serversMocks "github.com/bukka/wst/mocks/generated/run/servers"
	configsMocks "github.com/bukka/wst/mocks/generated/run/servers/configs"
	templatesMocks "github.com/bukka/wst/mocks/generated/run/servers/templates"
	templateMocks "github.com/bukka/wst/mocks/generated/run/services/template"
	"github.com/bukka/wst/run/environments"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/servers/templates"
	"github.com/bukka/wst/run/services/template"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func createParamMock(t *testing.T, val string) *parameterMocks.MockParameter {
	paramMock := parameterMocks.NewMockParameter(t)
	paramMock.TestData().Set("value", val)
	return paramMock
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
	tests := []struct {
		name              string
		config            map[string]types.Service
		instanceName      string
		instanceWorkspace string
		setupMocks        func(
			t *testing.T,
			pm *parametersMocks.MockMaker,
			tm *templateMocks.MockMaker,
		) (environments.Environments, servers.Servers, scripts.Scripts, Services)
		expectedService  func() *nativeService
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
						Scripts: types.ServiceScripts{
							IncludeAll:  true,
							IncludeList: nil,
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
						Scripts: types.ServiceScripts{
							IncludeAll: true,
						},
					},
					Public: true,
				},
			},
			instanceName:      "ti",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				dockerEnv := environmentMocks.NewMockEnvironment(t)
				dockerEnv.On("MarkUsed").Return()
				kubeEnv := environmentMocks.NewMockEnvironment(t)
				envs := environments.Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubeEnv,
				}
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
				sb := sandboxMocks.NewMockSandbox(t)
				sb.On("Available").Return(true)
				fpmPhpIniConfig := createConfigMock(t, "fpm-php-ini")
				fpmConfConfig := createConfigMock(t, "fpm-conf")
				nginxConfConfig := createConfigMock(t, "nginx-conf")
				fpmTemplates := templates.Templates{
					"fpm_conf": templatesMocks.NewMockTemplate(t),
				}
				nginxTemplates := templates.Templates{
					"ngx_conf": templatesMocks.NewMockTemplate(t),
				}
				fpmDebSrv := serversMocks.NewMockServer(t)
				fpmDebSrv.On("Sandbox", providers.DockerType).Return(sb, true)
				fpmDebSrv.On("Config", "php.ini").Return(fpmPhpIniConfig, true)
				fpmDebSrv.On("Config", "fpm.conf").Return(fpmConfConfig, true)
				fpmDebSrv.On("Templates").Return(fpmTemplates)
				nginxDebSrv := serversMocks.NewMockServer(t)
				nginxDebSrv.On("Sandbox", providers.DockerType).Return(sb, true)
				nginxDebSrv.On("Config", "nginx.conf").Return(nginxConfConfig, true)
				nginxDebSrv.On("Templates").Return(nginxTemplates)
				srvs := servers.Servers{
					"fpm": {
						"debian": fpmDebSrv,
					},
					"nginx": {
						"debian": nginxDebSrv,
					},
				}
				scrs := scripts.Scripts{
					"index.php": scriptsMocks.NewMockScript(t),
				}
				fpmTemplate := createTemplateMock(t, "fpm")
				nginxTemplate := createTemplateMock(t, "nginx")
				fpmSvc := &nativeService{
					name:             "fpm",
					fullName:         "ti-fpm",
					public:           false,
					scripts:          scrs,
					server:           fpmDebSrv,
					serverParameters: fpmServerParams,
					sandbox:          sb,
					task:             nil,
					environment:      dockerEnv,
					configs: map[string]nativeServiceConfig{
						"php.ini": {
							parameters: parameters.Parameters{
								"memory_limit": createParamMock(t, "1G"),
								"tag":          createParamMock(t, "prod"),
								"type":         createParamMock(t, "php"),
							},
							config: fpmPhpIniConfig,
						},
						"fpm.conf": {
							parameters: parameters.Parameters{
								"max_children": createParamMock(t, "10"),
								"tag":          createParamMock(t, "prod"),
								"type":         createParamMock(t, "php"),
							},
							config: fpmConfConfig,
						},
					},
					environmentConfigPaths: nil,
					workspaceConfigPaths:   nil,
					environmentScriptPaths: nil,
					workspaceScriptPaths:   nil,
					workspace:              "/test/workspace/fpm",
					template:               nil,
				}
				nginxSvc := &nativeService{
					name:             "nginx",
					fullName:         "ti-nginx",
					public:           true,
					scripts:          scrs,
					server:           nginxDebSrv,
					serverParameters: nginxServerParams,
					sandbox:          sb,
					task:             nil,
					environment:      dockerEnv,
					configs: map[string]nativeServiceConfig{
						"nginx.conf": {
							parameters: parameters.Parameters{
								"worker_connections": createParamMock(t, "1024"),
								"tag":                createParamMock(t, "prod"),
								"type":               createParamMock(t, "ws"),
							},
							config: nginxConfConfig,
						},
					},
					environmentConfigPaths: nil,
					workspaceConfigPaths:   nil,
					environmentScriptPaths: nil,
					workspaceScriptPaths:   nil,
					workspace:              "/test/workspace/nginx",
					template:               nil,
				}
				tm.On("Make", fpmSvc, mock.Anything, fpmTemplates).Return(fpmTemplate)
				finalFpmSvc := *fpmSvc
				finalFpmSvc.template = fpmTemplate
				tm.On("Make", nginxSvc, mock.Anything, nginxTemplates).Return(nginxTemplate)
				finalNginxSvc := *nginxSvc
				finalNginxSvc.template = nginxTemplate
				svcs := Services{
					"fpm":   &finalFpmSvc,
					"nginx": &finalNginxSvc,
				}
				return envs, srvs, scrs, svcs
			},
			expectError: false,
		},
		{
			name: "successful single service creation",
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s1"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				localEnv.On("MarkUsed").Return()
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
				configParams := parameters.Parameters{
					"p0": parameterMocks.NewMockParameter(t),
					"p1": parameterMocks.NewMockParameter(t),
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
				tmpls := templates.Templates{
					"t": templatesMocks.NewMockTemplate(t),
				}
				debSrv := serversMocks.NewMockServer(t)
				debSrv.On("Sandbox", providers.LocalType).Return(sb, true)
				debSrv.On("Config", "c").Return(cfg, true)
				debSrv.On("Templates").Return(tmpls)
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}
				tmpl := templateMocks.NewMockTemplate(t)
				svc := &nativeService{
					name:     "svc",
					fullName: "testInstance-svc",
					public:   true,
					scripts: scripts.Scripts{
						"s1": scriptsMocks.NewMockScript(t),
					},
					server:           debSrv,
					serverParameters: serverParams,
					sandbox:          sb,
					task:             nil,
					environment:      localEnv,
					configs: map[string]nativeServiceConfig{
						"c": {
							parameters: parameters.Parameters{
								"p0": parameterMocks.NewMockParameter(t),
								"p1": parameterMocks.NewMockParameter(t),
								"p2": parameterMocks.NewMockParameter(t),
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
				return envs, srvs, scrs, svcs
			},
			expectError: false,
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s1"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				localEnv.On("MarkUsed").Return()
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
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}

				return envs, srvs, scrs, nil
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s1"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
				localEnv := environmentMocks.NewMockEnvironment(t)
				localEnv.On("MarkUsed").Return()
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
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}

				return envs, srvs, scrs, nil
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s1"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
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
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}

				return envs, srvs, scrs, nil
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s1"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
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
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}

				return envs, srvs, scrs, nil
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s1"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
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
				srvs := servers.Servers{
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}

				return envs, srvs, scrs, nil
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s1"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
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

				return envs, srvs, scrs, nil
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s1"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
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

				return envs, srvs, scrs, nil
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
						Scripts: types.ServiceScripts{
							IncludeList: []string{"s3"},
						},
					},
					Public: true,
				},
			},
			instanceName:      "testInstance",
			instanceWorkspace: "/test/workspace",
			setupMocks: func(
				t *testing.T,
				pm *parametersMocks.MockMaker,
				tm *templateMocks.MockMaker,
			) (environments.Environments, servers.Servers, scripts.Scripts, Services) {
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
					"php": {
						"debian": debSrv,
					},
				}
				scrs := scripts.Scripts{
					"s1": scriptsMocks.NewMockScript(t),
					"s2": scriptsMocks.NewMockScript(t),
				}

				return envs, srvs, scrs, nil
			},
			expectError:      true,
			expectedErrorMsg: "script s3 not found for service svc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			parametersMakerMock := parametersMocks.NewMockMaker(t)
			templateMakerMock := templateMocks.NewMockMaker(t)

			maker := &nativeMaker{
				fnd:             fndMock,
				parametersMaker: parametersMakerMock,
				templateMaker:   templateMakerMock,
			}

			envs, srvs, scrs, svcs := tt.setupMocks(t, parametersMakerMock, templateMakerMock)

			locator, err := maker.Make(tt.config, scrs, srvs, envs, tt.instanceName, tt.instanceWorkspace)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				assert.Nil(t, locator)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, svcs, locator.Services())
			}
		})
	}
}
