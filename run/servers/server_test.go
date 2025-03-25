package servers

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	expectationsMocks "github.com/wstool/wst/mocks/generated/run/expectations"
	parametersMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	sandboxesMocks "github.com/wstool/wst/mocks/generated/run/sandboxes"
	sandboxMocks "github.com/wstool/wst/mocks/generated/run/sandboxes/sandbox"
	actionsMocks "github.com/wstool/wst/mocks/generated/run/servers/actions"
	configsMocks "github.com/wstool/wst/mocks/generated/run/servers/configs"
	templatesMocks "github.com/wstool/wst/mocks/generated/run/servers/templates"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/sandboxes"
	"github.com/wstool/wst/run/servers/actions"
	"github.com/wstool/wst/run/servers/configs"
	"github.com/wstool/wst/run/servers/templates"
	"os/user"
	"testing"
)

func TestServers_GetServer(t *testing.T) {
	tests := []struct {
		name          string
		servers       Servers
		serverName    string
		serverTag     string
		expectedName  string
		expectedTag   string
		expectedFound bool
	}{
		{
			name: "Existing server with tag in name",
			servers: Servers{
				"server1": {
					"production": &nativeServer{name: "s1"},
				},
			},
			serverName:    "server1/production",
			expectedName:  "server1",
			expectedTag:   "production",
			expectedFound: true,
		},
		{
			name: "Existing server with separate tag",
			servers: Servers{
				"server1": {
					"production": &nativeServer{name: "s1"},
				},
			},
			serverName:    "server1",
			serverTag:     "production",
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
			serverName:    "server1",
			expectedName:  "server1",
			expectedTag:   "default",
			expectedFound: true,
		},
		{
			name:          "Non-existing server",
			servers:       Servers{},
			serverName:    "unknown",
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
			serverName:    "server1/production",
			expectedName:  "server1",
			expectedTag:   "production",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, found := tt.servers.GetServer(tt.serverName, tt.serverTag)
			name, tag := composeNameAndTag(tt.serverName, tt.serverTag)
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

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	expectationsMock := expectationsMocks.NewMockMaker(t)
	parametersMock := parametersMocks.NewMockMaker(t)

	maker := CreateMaker(fndMock, expectationsMock, parametersMock).(*nativeMaker)

	assert.Equal(t, fndMock, maker.fnd)
	assert.Equal(t, parametersMock, maker.parametersMaker)
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

func testSequentialActions(t *testing.T, len int) []*actionsMocks.MockSequentialAction {
	seqActions := make([]*actionsMocks.MockSequentialAction, len)
	for i := 0; i < len; i++ {
		seqAction := actionsMocks.NewMockSequentialAction(t)
		// Differentiate
		seqAction.TestData().Set("sa_id", i)
		seqActions[i] = seqAction
	}
	return seqActions
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
	params := testParams(t, 4)
	tests := []struct {
		name       string
		spec       *types.Spec
		setupMocks func(
			t *testing.T,
			nm *nativeMaker,
			fndMock *appMocks.MockFoundation,
			actionsMock *actionsMocks.MockMaker,
			configsMock *configsMocks.MockMaker,
			parametersMock *parametersMocks.MockMaker,
			templatesMock *templatesMocks.MockMaker,
			sandboxesMock *sandboxesMocks.MockMaker,
			snbs []*sandboxMocks.MockSandbox,
		)
		expectedServers  func(fndMock app.Foundation, snbs []*sandboxMocks.MockSandbox) Servers
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "Successfully create non inherited server with user and group",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
			expectedServers: func(fndMock app.Foundation, snbs []*sandboxMocks.MockSandbox) Servers {
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
			name: "Successfully create non inherited server with user and without group",
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
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				u := &user.User{
					Uid:      "1100",
					Gid:      "1200",
					Username: "u1",
					Name:     "Default User",
					HomeDir:  "/home/u1",
				}
				g := &user.Group{
					Gid:  "1200",
					Name: "grp",
				}
				fndMock.On("User", "u1").Return(u, nil)
				fndMock.On("UserGroup", u).Return(g, nil)
			},
			expectedServers: func(fndMock app.Foundation, snbs []*sandboxMocks.MockSandbox) Servers {
				return Servers{
					"server1": {
						"prod": &nativeServer{
							fnd:        fndMock,
							name:       "server1",
							tag:        "prod",
							parentName: "",
							user:       "u1",
							group:      "grp",
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
			name: "Successfully create inherited servers without users and groups sets",
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
						Port:    0,
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				u := &user.User{
					Uid:      "1000",
					Gid:      "1100",
					Username: "du",
					Name:     "Default User",
					HomeDir:  "/home/du",
				}
				g := &user.Group{
					Gid:  "1100",
					Name: "dg",
				}
				fndMock.On("CurrentUser").Return(u, nil).Once()
				fndMock.On("UserGroup", u).Return(g, nil).Once()
				snbs[1].On("Inherit", snbs[2]).Return(nil)
				snbs[2].On("Inherit", snbs[3]).Return(nil)
			},
			expectedServers: func(fndMock app.Foundation, snbs []*sandboxMocks.MockSandbox) Servers {
				defaultServer := &nativeServer{
					fnd:        fndMock,
					name:       "base",
					tag:        "default",
					parentName: "",
					user:       "du",
					group:      "dg",
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
					user:       "du",
					group:      "dg",
					port:       1234,
					extended:   true,
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
					user:       "du",
					group:      "dg",
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
			name: "Successfully create inherited servers with users and groups sets",
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
						Port:    0,
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				u := &user.User{
					Uid:      "1000",
					Gid:      "1100",
					Username: "du",
					Name:     "Default User",
					HomeDir:  "/home/du",
				}
				g := &user.Group{
					Gid:  "1100",
					Name: "dg",
				}
				fndMock.On("CurrentUser").Return(u, nil).Once()
				fndMock.On("UserGroup", u).Return(g, nil).Once()
				snbs[1].On("Inherit", snbs[2]).Return(nil)
				snbs[2].On("Inherit", snbs[3]).Return(nil)
			},
			expectedServers: func(fndMock app.Foundation, snbs []*sandboxMocks.MockSandbox) Servers {
				defaultServer := &nativeServer{
					fnd:        fndMock,
					name:       "base",
					tag:        "default",
					parentName: "",
					user:       "du",
					group:      "dg",
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
					user:       "du",
					group:      "dg",
					port:       1234,
					extended:   true,
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
					user:       "du",
					group:      "dg",
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
			name: "Failure in circular server inheriting",
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
						Extends: "server1/dev",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
			},
			expectedError:    true,
			expectedErrorMsg: "circular server inheritance identified for server",
		},
		{
			name: "Failure in inheriting sandbox",
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
						Name:    "server1/dev",
						Extends: "server1/prod",
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
						Name:    "server1/prod",
						Extends: "base/default",
						Port:    0,
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
						Name:    "base/default",
						Extends: "",
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
					{
						Name:    "base/default",
						Extends: "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				u := &user.User{
					Uid:      "1000",
					Gid:      "1100",
					Username: "du",
					Name:     "Default User",
					HomeDir:  "/home/du",
				}
				g := &user.Group{
					Gid:  "1100",
					Name: "dg",
				}
				fndMock.On("CurrentUser").Return(u, nil).Once()
				fndMock.On("UserGroup", u).Return(g, nil).Once()
				snbs[2].On("Inherit", snbs[3]).Return(errors.New("inherit fail"))
			},
			expectedError:    true,
			expectedErrorMsg: "inherit fail",
		},
		{
			name: "Failure in user group creation",
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
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				u := &user.User{
					Uid:      "1100",
					Gid:      "1200",
					Username: "u1",
					Name:     "Default User",
					HomeDir:  "/home/u1",
				}
				fndMock.On("User", "u1").Return(u, nil)
				fndMock.On("UserGroup", u).Return(nil, errors.New("user group fail"))
			},
			expectedError:    true,
			expectedErrorMsg: "user group fail",
		},
		{
			name: "Failure in user fetching",
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
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				fndMock.On("User", "u1").Return(nil, errors.New("user fail"))
			},
			expectedError:    true,
			expectedErrorMsg: "user fail",
		},
		{
			name: "Failure in current user",
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
						User:    "",
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				fndMock.On("CurrentUser").Return(nil, errors.New("current user fail"))
			},
			expectedError:    true,
			expectedErrorMsg: "current user fail",
		},
		{
			name: "Failure in current user",
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
						Extends: "base/x",
						User:    "",
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
			expectedError:    true,
			expectedErrorMsg: "parent base/x for server server1/prod not found",
		},
		{
			name: "Failure in params maker",
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
						User:    "",
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				}).Return(nil, errors.New("params fail"))
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
			expectedError:    true,
			expectedErrorMsg: "params fail",
		},
		{
			name: "Failure in sandboxes maker",
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
						User:    "",
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				).Return(nil, errors.New("sandboxes fail"))
			},
			expectedError:    true,
			expectedErrorMsg: "sandboxes fail",
		},
		{
			name: "Failure in server templates maker",
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
						User:    "",
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				}).Return(nil, errors.New("server templates fail"))
			},
			expectedError:    true,
			expectedErrorMsg: "server templates fail",
		},
		{
			name: "Failure in server configs maker",
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
						User:    "",
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
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
				}).Return(nil, errors.New("server configs fail"))
			},
			expectedError:    true,
			expectedErrorMsg: "server configs fail",
		},
		{
			name: "Failure in actions maker",
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
						User:    "",
						Group:   "",
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
				fndMock *appMocks.MockFoundation,
				actionsMock *actionsMocks.MockMaker,
				configsMock *configsMocks.MockMaker,
				parametersMock *parametersMocks.MockMaker,
				templatesMock *templatesMocks.MockMaker,
				sandboxesMock *sandboxesMocks.MockMaker,
				snbs []*sandboxMocks.MockSandbox,
			) {
				actionsMock.On("Make", &types.ServerActions{
					Expect: map[string]types.ServerExpectationAction{
						"resp": &types.ServerResponseExpectation{
							Response: types.ResponseExpectation{
								Request: "x",
							},
						},
					},
				}).Return(nil, fmt.Errorf("actions error"))
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
			snbs := testSandboxes(t, 6)

			nm := &nativeMaker{
				fnd:             fndMock,
				actionsMaker:    actionsMock,
				configsMaker:    configsMock,
				templatesMaker:  templatesMock,
				parametersMaker: parametersMock,
				sandboxesMaker:  sandboxesMock,
			}

			if tt.setupMocks != nil {
				tt.setupMocks(
					t,
					nm,
					fndMock,
					actionsMock,
					configsMock,
					parametersMock,
					templatesMock,
					sandboxesMock,
					snbs,
				)
			}

			servers, err := nm.Make(tt.spec)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				assert.Nil(t, servers)
			} else {
				assert.NoError(t, err)
				expectedServers := tt.expectedServers(fndMock, snbs)
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

func testNativeServer(t *testing.T) *nativeServer {
	return &nativeServer{
		fnd:        appMocks.NewMockFoundation(t),
		name:       "sn",
		tag:        "tt",
		user:       "u1",
		group:      "g1",
		port:       8888,
		parentName: "",
		parent:     nil,
		extended:   false,
		actions:    nil,
		configs:    nil,
		templates:  nil,
	}
}

func Test_nativeServer_Group(t *testing.T) {
	s := testNativeServer(t)
	assert.Equal(t, "g1", s.Group())
}

func Test_nativeServer_User(t *testing.T) {
	s := testNativeServer(t)
	assert.Equal(t, "u1", s.User())
}

func Test_nativeServer_Port(t *testing.T) {
	s := testNativeServer(t)
	assert.Equal(t, int32(8888), s.Port())
}

func Test_nativeServer_Template(t *testing.T) {
	tmpls := testTemplates(t, 2)
	s := testNativeServer(t)
	s.templates = templates.Templates{
		"t1": tmpls[0],
		"t2": tmpls[1],
	}
	tmpl, found := s.Template("t1")
	assert.True(t, found)
	assert.Equal(t, tmpls[0], tmpl)
}

func Test_nativeServer_Templates(t *testing.T) {
	tmpls := testTemplates(t, 2)
	s := testNativeServer(t)
	s.templates = templates.Templates{
		"t1": tmpls[0],
		"t2": tmpls[1],
	}
	assert.Equal(t, s.templates, s.Templates())
}

func Test_nativeServer_Parameters(t *testing.T) {
	params := testParams(t, 2)
	s := testNativeServer(t)
	s.parameters = parameters.Parameters{
		"key1": params[0],
		"key2": params[1],
	}
	assert.Equal(t, s.parameters, s.Parameters())
}

func Test_nativeServer_ExpectAction(t *testing.T) {
	eas := testExpectActions(t, 2)
	s := testNativeServer(t)
	s.actions = &actions.Actions{
		Expect: map[string]actions.ExpectAction{
			"e1": eas[0],
			"e2": eas[1],
		},
	}
	ea, found := s.ExpectAction("e2")
	assert.True(t, found)
	assert.Equal(t, eas[1], ea)
}

func Test_nativeServer_SequentialAction(t *testing.T) {
	// Create mock sequential actions
	sas := testSequentialActions(t, 2)

	// Create a test native server and set its sequential actions
	s := testNativeServer(t)
	s.actions = &actions.Actions{
		Sequential: map[string]actions.SequentialAction{
			"s1": sas[0],
			"s2": sas[1],
		},
	}

	// Test querying a specific sequential action
	sa, found := s.SequentialAction("s2")
	assert.True(t, found)
	assert.Equal(t, sas[1], sa)
}

func Test_nativeServer_Config(t *testing.T) {
	confs := testConfigs(t, 2)
	s := testNativeServer(t)
	s.configs = configs.Configs{
		"c1": confs[0],
		"c2": confs[1],
	}
	conf, found := s.Config("c2")
	assert.True(t, found)
	assert.Equal(t, confs[1], conf)
}

func Test_nativeServer_Configs(t *testing.T) {
	confs := testConfigs(t, 2)
	s := testNativeServer(t)
	s.configs = configs.Configs{
		"c1": confs[0],
		"c2": confs[1],
	}
	assert.Equal(t, s.configs, s.Configs())
}

func Test_nativeServer_ConfigPaths(t *testing.T) {
	confs := testConfigs(t, 2)
	confs[0].On("FilePath").Return("/tmp/c1")
	confs[1].On("FilePath").Return("/tmp/c2")
	s := testNativeServer(t)
	s.configs = configs.Configs{
		"c1": confs[0],
		"c2": confs[1],
	}
	assert.Equal(t, map[string]string{
		"c1": "/tmp/c1",
		"c2": "/tmp/c2",
	}, s.ConfigPaths())
}

func Test_nativeServer_Sandbox(t *testing.T) {
	sbs := testSandboxes(t, 2)
	s := testNativeServer(t)
	s.sandboxes = sandboxes.Sandboxes{
		"s1": sbs[0],
		"s2": sbs[1],
	}
	sb, found := s.Sandbox("s1")
	assert.True(t, found)
	assert.Equal(t, sbs[0], sb)
}
