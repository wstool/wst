package servers

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	sandboxesMocks "github.com/bukka/wst/mocks/generated/run/sandboxes"
	sandboxMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/sandbox"
	actionsMocks "github.com/bukka/wst/mocks/generated/run/servers/actions"
	configsMocks "github.com/bukka/wst/mocks/generated/run/servers/configs"
	templatesMocks "github.com/bukka/wst/mocks/generated/run/servers/templates"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/sandboxes"
	"github.com/bukka/wst/run/servers/actions"
	"github.com/bukka/wst/run/servers/configs"
	"github.com/bukka/wst/run/servers/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServers_GetServer(t *testing.T) {
	tests := []struct {
		name          string
		servers       Servers
		fullName      string
		expectedName  string
		expectedTag   string
		expectedFound bool
	}{
		{
			name: "Existing server with tag",
			servers: Servers{
				"server1": {
					"production": &nativeServer{name: "s1"},
				},
			},
			fullName:      "server1/production",
			expectedName:  "server1",
			expectedTag:   "production",
			expectedFound: true,
		},
		{
			name: "Existing server without tag defaults to 'default'",
			servers: Servers{
				"server1": {
					"default": &nativeServer{name: "s2"},
				},
			},
			fullName:      "server1",
			expectedName:  "server1",
			expectedTag:   "default",
			expectedFound: true,
		},
		{
			name:          "Non-existing server",
			servers:       Servers{},
			fullName:      "unknown",
			expectedName:  "unknown",
			expectedTag:   "default",
			expectedFound: false,
		},
		{
			name: "Non-existing tag",
			servers: Servers{
				"server1": {
					"staging": &nativeServer{name: "s3"},
				},
			},
			fullName:      "server1/production",
			expectedName:  "server1",
			expectedTag:   "production",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, found := tt.servers.GetServer(tt.fullName)
			name, tag := splitFullName(tt.fullName)
			assert.Equal(t, tt.expectedName, name, "server name should match")
			assert.Equal(t, tt.expectedTag, tag, "server tag should match")
			assert.Equal(t, tt.expectedFound, found, "server found state should match")
			if found {
				assert.NotNil(t, server, "server should not be nil when found")
			} else {
				assert.Nil(t, server, "server should be nil when not found")
			}
		})
	}
}

func testParams(t *testing.T, len int) []*parameterMocks.MockParameter {
	params := make([]*parameterMocks.MockParameter, len)
	for i := 0; i < len; i++ {
		param := parameterMocks.NewMockParameter(t)
		// Differentiate
		param.TestData().Set("param_id", i)
		params[i] = param
	}
	return params
}

func testExpectActions(t *testing.T, len int) []*actionsMocks.MockExpectAction {
	params := make([]*actionsMocks.MockExpectAction, len)
	for i := 0; i < len; i++ {
		param := actionsMocks.NewMockExpectAction(t)
		// Differentiate
		param.TestData().Set("ea_id", i)
		params[i] = param
	}
	return params
}

func testConfigs(t *testing.T, len int) []*configsMocks.MockConfig {
	params := make([]*configsMocks.MockConfig, len)
	for i := 0; i < len; i++ {
		param := configsMocks.NewMockConfig(t)
		// Differentiate
		param.TestData().Set("conf_id", i)
		params[i] = param
	}
	return params
}

func testSandboxes(t *testing.T, len int) []*sandboxMocks.MockSandbox {
	params := make([]*sandboxMocks.MockSandbox, len)
	for i := 0; i < len; i++ {
		param := sandboxMocks.NewMockSandbox(t)
		// Differentiate
		param.TestData().Set("sandbox_id", i)
		params[i] = param
	}
	return params
}

func testTemplates(t *testing.T, len int) []*templatesMocks.MockTemplate {
	params := make([]*templatesMocks.MockTemplate, len)
	for i := 0; i < len; i++ {
		param := templatesMocks.NewMockTemplate(t)
		// Differentiate
		param.TestData().Set("template_id", i)
		params[i] = param
	}
	return params
}

func Test_nativeMaker_Make(t *testing.T) {
	eas := testExpectActions(t, 4)
	confs := testConfigs(t, 4)
	tmpls := testTemplates(t, 4)
	snbs := testSandboxes(t, 6)
	params := testParams(t, 4)
	tests := []struct {
		name       string
		spec       *types.Spec
		setupMocks func(
			t *testing.T,
			nm *nativeMaker,
			actionsMock *actionsMocks.MockMaker,
			configsMock *configsMocks.MockMaker,
			parametersMock *parametersMocks.MockMaker,
			templatesMock *templatesMocks.MockMaker,
			sandboxesMock *sandboxesMocks.MockMaker,
		)
		expectedServers  func(fndMock app.Foundation) Servers
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "Successfully create non inherited servers",
			spec: &types.Spec{
				Sandboxes: map[string]types.Sandbox{
					"docker": &types.DockerSandbox{
						Available: true,
						Image: types.ContainerImage{
							Name: "img",
						},
					},
				},
				Servers: []types.Server{
					{
						Name:    "server1/prod",
						Extends: "",
						User:    "u1",
						Group:   "g1",
						Port:    1234,
						Configs: map[string]types.ServerConfig{
							"ct": {File: "test.php"},
						},
						Templates: map[string]types.ServerTemplate{
							"tt": {File: "t.tmpl"},
						},
						Sandboxes: map[string]types.Sandbox{
							"local": &types.LocalSandbox{Available: true},
						},
						Parameters: types.Parameters{
							"key": "value",
						},
						Actions: types.ServerActions{
							Expect: map[string]types.ServerExpectationAction{
								"resp": &types.ServerResponseExpectation{
									Response: types.ResponseExpectation{
										Request: "x",
									},
								},
							},
						},
					},
				},
			},
			setupMocks: func(t *testing.T,
				nm *nativeMaker,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
			) {
				actionsMock.On("Make", &types.ServerActions{
					Expect: map[string]types.ServerExpectationAction{
						"resp": &types.ServerResponseExpectation{
							Response: types.ResponseExpectation{
								Request: "x",
							},
						},
					},
				}).Return(&actions.Actions{
					Expect: map[string]actions.ExpectAction{
						"resp": eas[0],
					},
				}, nil)
				configsMock.On("Make", map[string]types.ServerConfig{
					"ct": {File: "test.php"},
				}).Return(configs.Configs{
					"ct": confs[0],
				}, nil)
				templatesMock.On("Make", map[string]types.ServerTemplate{
					"tt": {File: "t.tmpl"},
				}).Return(templates.Templates{
					"tt": tmpls[0],
				}, nil)
				parametersMock.On("Make", types.Parameters{
					"key": "value",
				}).Return(parameters.Parameters{
					"key": params[0],
				}, nil)
				sandboxesMock.On(
					"MakeSandboxes",
					map[string]types.Sandbox{
						"docker": &types.DockerSandbox{
							Available: true,
							Image: types.ContainerImage{
								Name: "img",
							},
						},
					},
					map[string]types.Sandbox{
						"local": &types.LocalSandbox{Available: true},
					},
				).Return(sandboxes.Sandboxes{
					providers.DockerType: snbs[0],
					providers.LocalType:  snbs[1],
				}, nil)
			},
			expectedServers: func(fndMock app.Foundation) Servers {
				return Servers{
					"server1": {
						"prod": &nativeServer{
							fnd:        fndMock,
							name:       "server1",
							tag:        "prod",
							parentName: "",
							user:       "u1",
							group:      "g1",
							port:       1234,
							extended:   true,
							actions: &actions.Actions{
								Expect: map[string]actions.ExpectAction{
									"resp": eas[0],
								},
							},
							configs: configs.Configs{
								"ct": confs[0],
							},
							templates: templates.Templates{
								"tt": tmpls[0],
							},
							parameters: parameters.Parameters{
								"key": params[0],
							},
							sandboxes: sandboxes.Sandboxes{
								providers.DockerType: snbs[0],
								providers.LocalType:  snbs[1],
							},
						},
					},
				}
			},
		},
		{
			name: "Successfully create inherited servers",
			spec: &types.Spec{
				Sandboxes: map[string]types.Sandbox{
					"docker": &types.DockerSandbox{
						Available: true,
						Image: types.ContainerImage{
							Name: "img",
						},
					},
				},
				Servers: []types.Server{
					{
						Name:    "server1/prod",
						Extends: "base/default",
						User:    "u1",
						Group:   "g1",
						Port:    1234,
						Configs: map[string]types.ServerConfig{
							"ct": {File: "test-prod.php"},
						},
						Templates: map[string]types.ServerTemplate{
							"tt": {File: "t-prod.tmpl"},
						},
						Sandboxes: map[string]types.Sandbox{
							"local": &types.LocalSandbox{Available: true, Dirs: map[string]string{
								"run": "test",
							}},
						},
						Parameters: types.Parameters{
							"prod_key": "prod_value",
						},
						Actions: types.ServerActions{
							Expect: map[string]types.ServerExpectationAction{
								"resp": &types.ServerResponseExpectation{
									Response: types.ResponseExpectation{
										Request: "y",
									},
								},
							},
						},
					},
					{
						Name:    "server1/dev",
						Extends: "server1/prod",
						User:    "u1",
						Group:   "g1",
						Port:    1234,
						Configs: map[string]types.ServerConfig{
							"ct": {File: "test.php"},
						},
						Templates: map[string]types.ServerTemplate{
							"tt": {File: "t.tmpl"},
						},
						Sandboxes: map[string]types.Sandbox{
							"local": &types.LocalSandbox{Available: true},
						},
						Parameters: types.Parameters{
							"key": "value",
						},
						Actions: types.ServerActions{
							Expect: map[string]types.ServerExpectationAction{
								"resp": &types.ServerResponseExpectation{
									Response: types.ResponseExpectation{
										Request: "x",
									},
								},
							},
						},
					},
					{
						Name:    "base/default",
						Extends: "",
						User:    "user",
						Group:   "grp",
						Port:    1234,
						Configs: map[string]types.ServerConfig{
							"ct": {File: "test-base.php"},
						},
						Templates: map[string]types.ServerTemplate{
							"tt": {File: "t-base.tmpl"},
						},
						Sandboxes: map[string]types.Sandbox{
							"local": &types.LocalSandbox{Available: true, Dirs: map[string]string{
								"run": "base",
							}},
						},
						Parameters: types.Parameters{
							"key_base": "value_base",
							"key":      "value_base_base",
						},
						Actions: types.ServerActions{
							Expect: map[string]types.ServerExpectationAction{
								"resp": &types.ServerResponseExpectation{
									Response: types.ResponseExpectation{
										Request: "z",
									},
								},
							},
						},
					},
				},
			},
			setupMocks: func(t *testing.T,
				nm *nativeMaker,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
			) {
				actionsMock.On("Make", &types.ServerActions{
					Expect: map[string]types.ServerExpectationAction{
						"resp": &types.ServerResponseExpectation{
							Response: types.ResponseExpectation{
								Request: "z",
							},
						},
					},
				}).Return(&actions.Actions{
					Expect: map[string]actions.ExpectAction{
						"resp": eas[2],
					},
				}, nil)
				actionsMock.On("Make", &types.ServerActions{
					Expect: map[string]types.ServerExpectationAction{
						"resp": &types.ServerResponseExpectation{
							Response: types.ResponseExpectation{
								Request: "y",
							},
						},
					},
				}).Return(&actions.Actions{
					Expect: map[string]actions.ExpectAction{
						"resp": eas[1],
					},
				}, nil)
				actionsMock.On("Make", &types.ServerActions{
					Expect: map[string]types.ServerExpectationAction{
						"resp": &types.ServerResponseExpectation{
							Response: types.ResponseExpectation{
								Request: "x",
							},
						},
					},
				}).Return(&actions.Actions{
					Expect: map[string]actions.ExpectAction{
						"resp": eas[0],
					},
				}, nil)
				configsMock.On("Make", map[string]types.ServerConfig{
					"ct": {File: "test-base.php"},
				}).Return(configs.Configs{
					"ct": confs[2],
				}, nil)
				configsMock.On("Make", map[string]types.ServerConfig{
					"ct": {File: "test-prod.php"},
				}).Return(configs.Configs{
					"ct": confs[1],
				}, nil)
				configsMock.On("Make", map[string]types.ServerConfig{
					"ct": {File: "test.php"},
				}).Return(configs.Configs{
					"ct": confs[0],
				}, nil)
				templatesMock.On("Make", map[string]types.ServerTemplate{
					"tt": {File: "t-base.tmpl"},
				}).Return(templates.Templates{
					"tt": tmpls[2],
				}, nil)

				templatesMock.On("Make", map[string]types.ServerTemplate{
					"tt": {File: "t-prod.tmpl"},
				}).Return(templates.Templates{
					"tt": tmpls[1],
				}, nil)
				templatesMock.On("Make", map[string]types.ServerTemplate{
					"tt": {File: "t.tmpl"},
				}).Return(templates.Templates{
					"tt": tmpls[0],
				}, nil)
				parametersMock.On("Make", types.Parameters{
					"key_base": "value_base",
					"key":      "value_base_base",
				}).Return(parameters.Parameters{
					"key":      params[2],
					"key_base": params[3],
				}, nil)
				parametersMock.On("Make", types.Parameters{
					"prod_key": "prod_value",
				}).Return(parameters.Parameters{
					"prod_key": params[1],
				}, nil)
				parametersMock.On("Make", types.Parameters{
					"key": "value",
				}).Return(parameters.Parameters{
					"key": params[0],
				}, nil)
				sandboxesMock.On(
					"MakeSandboxes",
					map[string]types.Sandbox{
						"docker": &types.DockerSandbox{
							Available: true,
							Image: types.ContainerImage{
								Name: "img",
							},
						},
					},
					map[string]types.Sandbox{
						"local": &types.LocalSandbox{Available: true, Dirs: map[string]string{
							"run": "base",
						}},
					},
				).Return(sandboxes.Sandboxes{
					providers.DockerType: snbs[0],
					providers.LocalType:  snbs[3],
				}, nil)
				sandboxesMock.On(
					"MakeSandboxes",
					map[string]types.Sandbox{
						"docker": &types.DockerSandbox{
							Available: true,
							Image: types.ContainerImage{
								Name: "img",
							},
						},
					},
					map[string]types.Sandbox{
						"local": &types.LocalSandbox{Available: true, Dirs: map[string]string{
							"run": "test",
						}},
					},
				).Return(sandboxes.Sandboxes{
					providers.DockerType: snbs[0],
					providers.LocalType:  snbs[2],
				}, nil)
				sandboxesMock.On(
					"MakeSandboxes",
					map[string]types.Sandbox{
						"docker": &types.DockerSandbox{
							Available: true,
							Image: types.ContainerImage{
								Name: "img",
							},
						},
					},
					map[string]types.Sandbox{
						"local": &types.LocalSandbox{Available: true},
					},
				).Return(sandboxes.Sandboxes{
					providers.DockerType: snbs[0],
					providers.LocalType:  snbs[1],
				}, nil)
				snbs[1].On("Inherit", snbs[2]).Return(nil)
				snbs[2].On("Inherit", snbs[3]).Return(nil)
			},
			expectedServers: func(fndMock app.Foundation) Servers {
				defaultServer := &nativeServer{
					fnd:        fndMock,
					name:       "base",
					tag:        "default",
					parentName: "",
					user:       "user",
					group:      "grp",
					port:       1234,
					extended:   true,
					actions: &actions.Actions{
						Expect: map[string]actions.ExpectAction{
							"resp": eas[2],
						},
					},
					configs: configs.Configs{
						"ct": confs[2],
					},
					templates: templates.Templates{
						"tt": tmpls[2],
					},
					parameters: parameters.Parameters{
						"key":      params[2],
						"key_base": params[3],
					},
					sandboxes: sandboxes.Sandboxes{
						providers.DockerType: snbs[0],
						providers.LocalType:  snbs[3],
					},
				}
				prodServer := &nativeServer{
					fnd:        fndMock,
					name:       "server1",
					tag:        "prod",
					parentName: "base/default",
					parent:     defaultServer,
					user:       "u1",
					group:      "g1",
					port:       1234,
					extended:   false,
					actions: &actions.Actions{
						Expect: map[string]actions.ExpectAction{
							"resp": eas[1],
						},
					},
					configs: configs.Configs{
						"ct": confs[1],
					},
					templates: templates.Templates{
						"tt": tmpls[1],
					},
					parameters: parameters.Parameters{
						"prod_key": params[1],
						"key":      params[2],
						"key_base": params[3],
					},
					sandboxes: sandboxes.Sandboxes{
						providers.DockerType: snbs[0],
						providers.LocalType:  snbs[2],
					},
				}
				devServer := &nativeServer{
					fnd:        fndMock,
					name:       "server1",
					tag:        "dev",
					parentName: "server1/prod",
					parent:     prodServer,
					user:       "u1",
					group:      "g1",
					port:       1234,
					extended:   false,
					actions: &actions.Actions{
						Expect: map[string]actions.ExpectAction{
							"resp": eas[0],
						},
					},
					configs: configs.Configs{
						"ct": confs[0],
					},
					templates: templates.Templates{
						"tt": tmpls[0],
					},
					parameters: parameters.Parameters{
						"prod_key": params[1],
						"key":      params[0],
						"key_base": params[3],
					},
					sandboxes: sandboxes.Sandboxes{
						providers.DockerType: snbs[0],
						providers.LocalType:  snbs[1],
					},
				}

				return Servers{
					"base": {
						"default": defaultServer,
					},
					"server1": {
						"dev":  devServer,
						"prod": prodServer,
					},
				}
			},
		},
		{
			name: "Failure in actions maker",
			spec: &types.Spec{
				Servers: []types.Server{
					{
						Name:       "server1/prod",
						Actions:    types.ServerActions{},
						Configs:    map[string]types.ServerConfig{},
						Templates:  map[string]types.ServerTemplate{},
						Sandboxes:  map[string]types.Sandbox{},
						Parameters: types.Parameters{},
					},
				},
			},
			setupMocks: func(t *testing.T,
				nm *nativeMaker,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
			) {
				actionsMock.On("Make", mock.Anything).Return(nil, fmt.Errorf("actions error"))
			},
			expectedError:    true,
			expectedErrorMsg: "actions error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			actionsMock := actionsMocks.NewMockMaker(t)
			configsMock := configsMocks.NewMockMaker(t)
			parametersMock := parametersMocks.NewMockMaker(t)
			templatesMock := templatesMocks.NewMockMaker(t)
			sandboxesMock := sandboxesMocks.NewMockMaker(t)

			nm := &nativeMaker{
				fnd:             fndMock,
				actionsMaker:    actionsMock,
				configsMaker:    configsMock,
				templatesMaker:  templatesMock,
				parametersMaker: parametersMock,
				sandboxesMaker:  sandboxesMock,
			}

			if tt.setupMocks != nil {
				tt.setupMocks(t, nm, actionsMock, configsMock, parametersMock, templatesMock, sandboxesMock)
			}

			servers, err := nm.Make(tt.spec)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				assert.Nil(t, servers)
			} else {
				assert.NoError(t, err)
				expectedServers := tt.expectedServers(fndMock)
				// do a deeper check manually to get better idea what failed (otherwise it reports just pointer diffs)
				assert.Equal(t, len(expectedServers), len(servers))
				for name, nameServers := range servers {
					expectedNameServers, ok := expectedServers[name]
					require.True(t, ok, fmt.Sprintf("server %s not found", name))
					assert.Equal(t, len(expectedNameServers), len(nameServers))
					for tag, server := range nameServers {
						expectedServer, ok := expectedNameServers[tag]
						require.True(t, ok, fmt.Sprintf("server %s/%s not found", name, tag))
						assert.Equal(t, expectedServer, server)
					}
				}
			}

			actionsMock.AssertExpectations(t)
			configsMock.AssertExpectations(t)
			parametersMock.AssertExpectations(t)
			templatesMock.AssertExpectations(t)
			sandboxesMock.AssertExpectations(t)
		})
	}
}
