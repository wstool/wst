package instances

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	externalMocks "github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	actionsMocks "github.com/wstool/wst/mocks/generated/run/actions"
	actionMocks "github.com/wstool/wst/mocks/generated/run/actions/action"
	environmentsMocks "github.com/wstool/wst/mocks/generated/run/environments"
	environmentMocks "github.com/wstool/wst/mocks/generated/run/environments/environment"
	expectationsMocks "github.com/wstool/wst/mocks/generated/run/expectations"
	runtimeMocks "github.com/wstool/wst/mocks/generated/run/instances/runtime"
	parametersMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	scriptsMocks "github.com/wstool/wst/mocks/generated/run/resources/scripts"
	serversMocks "github.com/wstool/wst/mocks/generated/run/servers"
	servicesMocks "github.com/wstool/wst/mocks/generated/run/services"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/environments"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/resources/scripts"
	"github.com/wstool/wst/run/servers"
	"github.com/wstool/wst/run/services"
	"github.com/wstool/wst/run/spec/defaults"
	"testing"
	"time"
)

func TestCreateInstanceMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	expectationsMakerMock := expectationsMocks.NewMockMaker(t)
	parametersMakerMock := parametersMocks.NewMockMaker(t)

	instanceMaker := CreateInstanceMaker(fndMock, expectationsMakerMock, parametersMakerMock)

	actualMaker, ok := instanceMaker.(*nativeInstanceMaker)
	assert.True(t, ok, "CreateInstanceMaker should return an instance of nativeInstanceMaker")
	assert.Same(t, actualMaker.fnd, fndMock)
	assert.NotNil(t, actualMaker.environmentMaker)
	assert.NotNil(t, actualMaker.actionMaker)
	assert.Same(t, actualMaker.parametersMaker, parametersMakerMock)
	assert.NotNil(t, actualMaker.scriptsMaker)
	assert.NotNil(t, actualMaker.servicesMaker)
	assert.NotNil(t, actualMaker.runtimeMaker)
}

func TestNativeInstanceMaker_Make(t *testing.T) {
	testScripts := map[string]types.Script{
		"test-script": {
			Content:    "test content",
			Path:       "/path/to/script",
			Mode:       "0644",
			Parameters: nil,
		},
	}
	testInstanceEnvironments := map[string]types.Environment{
		"local": types.LocalEnvironment{
			Ports: types.EnvironmentPorts{Start: 9000},
		},
	}
	testSpecEnvironments := map[string]types.Environment{
		"local": types.LocalEnvironment{
			Ports: types.EnvironmentPorts{Start: 8000},
		},
	}
	testServices := map[string]types.Service{
		"svc": {
			Server: types.ServiceServer{
				Name:    "test",
				Sandbox: "local",
			},
		},
	}
	testServers := servers.Servers{
		"t1": map[string]servers.Server{
			"t2": serversMocks.NewMockServer(t),
		},
	}
	testActions := []types.Action{
		types.StartAction{
			Service: "s1",
		},
		types.BenchAction{
			Service: "btest",
			Id:      "bt",
		},
		types.StopAction{
			Service: "s1",
		},
	}
	testData := runtimeMocks.NewMockData(t)
	testDefaultsParams := parameters.Parameters{
		"default_key": parameterMocks.NewMockParameter(t),
	}
	testParams := types.Parameters{
		"test_key": "test_value",
	}
	testResultParams := parameters.Parameters{
		"test_key": parameterMocks.NewMockParameter(t),
	}
	testExtendsParams := types.Parameters{
		"etest_key": "test_value",
	}
	testExtendsResultParams := parameters.Parameters{
		"etest_key": parameterMocks.NewMockParameter(t),
	}
	instanceIdx := 1
	tests := []struct {
		name       string
		setupMocks func(
			*testing.T,
			*parametersMocks.MockMaker,
			*runtimeMocks.MockMaker,
			*defaults.Defaults,
		)
		instanceConfig                 types.Instance
		defaults                       defaults.Defaults
		expectError                    bool
		expectedErrorMsg               string
		expectedExtendedParams         parameters.Parameters
		expectedInstanceTimeout        time.Duration
		expectedInstanceTimeoutDefault bool
		expectedActionTimeout          int
		expectedActionTimeoutDefault   bool
	}{
		{
			name: "successful creation with instance timeouts and no extends",
			setupMocks: func(
				t *testing.T,
				paramsMaker *parametersMocks.MockMaker,
				runtimeMaker *runtimeMocks.MockMaker,
				dflts *defaults.Defaults,
			) {
				paramsMaker.On("Make", testParams).Return(testResultParams, nil)
				runtimeMaker.On("MakeData").Return(testData)
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Timeouts: types.InstanceTimeouts{
					Action:  5000,
					Actions: 10000,
				},
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			defaults: defaults.Defaults{
				Service: defaults.ServiceDefaults{
					Sandbox: "local",
					Server: defaults.ServiceServerDefaults{
						Tag: "latest",
					},
				},
				Timeouts: defaults.TimeoutsDefaults{
					Actions: 15000,
					Action:  8000,
				},
				Parameters: testDefaultsParams,
			},
			expectedInstanceTimeout:        10000 * time.Millisecond,
			expectedInstanceTimeoutDefault: false,
			expectedActionTimeout:          5000,
			expectedActionTimeoutDefault:   false,
		},
		{
			name: "successful creation with default instance timeouts and extend",
			setupMocks: func(
				t *testing.T,
				paramsMaker *parametersMocks.MockMaker,
				runtimeMaker *runtimeMocks.MockMaker,
				dflts *defaults.Defaults,
			) {
				paramsMaker.On("Make", testParams).Return(testResultParams, nil)
				paramsMaker.On("Make", testExtendsParams).Return(testExtendsResultParams, nil)
				runtimeMaker.On("MakeData").Return(testData)
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Timeouts: types.InstanceTimeouts{
					Action:  0,
					Actions: 0,
				},
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Extends: types.InstanceExtends{
					Name:       "test-extends",
					Parameters: testExtendsParams,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			defaults: defaults.Defaults{
				Service: defaults.ServiceDefaults{
					Sandbox: "local",
					Server: defaults.ServiceServerDefaults{
						Tag: "latest",
					},
				},
				Timeouts: defaults.TimeoutsDefaults{
					Actions: 15000,
					Action:  8000,
				},
				Parameters: testDefaultsParams,
			},
			expectedExtendedParams:         testExtendsResultParams,
			expectedInstanceTimeout:        15000 * time.Millisecond,
			expectedInstanceTimeoutDefault: true,
			expectedActionTimeout:          8000,
			expectedActionTimeoutDefault:   true,
		},
		{
			name: "failed creation due to params error",
			setupMocks: func(
				t *testing.T,
				paramsMaker *parametersMocks.MockMaker,
				runtimeMaker *runtimeMocks.MockMaker,
				dflts *defaults.Defaults,
			) {
				paramsMaker.On("Make", testExtendsParams).Return(testExtendsResultParams, nil)
				paramsMaker.On("Make", testParams).Return(
					nil, errors.New("main param error"))
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Timeouts: types.InstanceTimeouts{
					Action:  0,
					Actions: 0,
				},
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Extends: types.InstanceExtends{
					Name:       "test-extends",
					Parameters: testExtendsParams,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			defaults: defaults.Defaults{
				Service: defaults.ServiceDefaults{
					Sandbox: "local",
					Server: defaults.ServiceServerDefaults{
						Tag: "latest",
					},
				},
				Timeouts: defaults.TimeoutsDefaults{
					Actions: 15000,
					Action:  8000,
				},
				Parameters: testDefaultsParams,
			},
			expectError:      true,
			expectedErrorMsg: "main param error",
		},
		{
			name: "failed creation due to extend param error",
			setupMocks: func(
				t *testing.T,
				paramsMaker *parametersMocks.MockMaker,
				runtimeMaker *runtimeMocks.MockMaker,
				dflts *defaults.Defaults,
			) {
				paramsMaker.On("Make", testExtendsParams).Return(
					nil, errors.New("extend param error"))
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Timeouts: types.InstanceTimeouts{
					Action:  0,
					Actions: 0,
				},
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Extends: types.InstanceExtends{
					Name:       "test-extends",
					Parameters: testExtendsParams,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			defaults: defaults.Defaults{
				Service: defaults.ServiceDefaults{
					Sandbox: "local",
					Server: defaults.ServiceServerDefaults{
						Tag: "latest",
					},
				},
				Timeouts: defaults.TimeoutsDefaults{
					Actions: 15000,
					Action:  8000,
				},
				Parameters: testDefaultsParams,
			},
			expectError:      true,
			expectedErrorMsg: "extend param error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			actionMaker := actionsMocks.NewMockActionMaker(t)
			serviceMaker := servicesMocks.NewMockMaker(t)
			scriptsMaker := scriptsMocks.NewMockMaker(t)
			paramsMaker := parametersMocks.NewMockMaker(t)
			envMaker := environmentsMocks.NewMockMaker(t)
			runtimeMaker := runtimeMocks.NewMockMaker(t)

			maker := &nativeInstanceMaker{
				fnd:              fndMock,
				actionMaker:      actionMaker,
				servicesMaker:    serviceMaker,
				scriptsMaker:     scriptsMaker,
				parametersMaker:  paramsMaker,
				environmentMaker: envMaker,
				runtimeMaker:     runtimeMaker,
			}

			tt.setupMocks(t, paramsMaker, runtimeMaker, &tt.defaults)

			specWorkspace := "/fake/workspace"
			got, err := maker.Make(
				tt.instanceConfig, instanceIdx, testSpecEnvironments, &tt.defaults, testServers, specWorkspace)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				actualInstance, ok := got.(*nativeInstance)
				assert.True(t, ok)
				expectedInstance := &nativeInstance{
					fnd:                    fndMock,
					runtimeMaker:           runtimeMaker,
					scriptsMaker:           scriptsMaker,
					environmentMaker:       envMaker,
					servicesMaker:          serviceMaker,
					actionMaker:            actionMaker,
					configActions:          tt.instanceConfig.Actions,
					configServices:         tt.instanceConfig.Services,
					configEnvs:             testSpecEnvironments,
					configInstanceEnvs:     tt.instanceConfig.Environments,
					configResources:        tt.instanceConfig.Resources,
					name:                   tt.instanceConfig.Name,
					index:                  instanceIdx,
					specWorkspace:          specWorkspace,
					initialized:            false,
					abstract:               tt.instanceConfig.Abstract,
					extendingStarted:       false,
					extendName:             tt.instanceConfig.Extends.Name,
					extendParams:           tt.expectedExtendedParams,
					params:                 testResultParams,
					defaults:               &tt.defaults,
					servers:                testServers,
					actions:                nil,
					services:               nil,
					envs:                   nil,
					runData:                testData,
					instanceTimeout:        tt.expectedInstanceTimeout,
					instanceTimeoutDefault: tt.expectedInstanceTimeoutDefault,
					actionTimeout:          tt.expectedActionTimeout,
					actionTimeoutDefault:   tt.expectedActionTimeoutDefault,
					workspace:              "",
				}
				assert.Equal(t, expectedInstance, actualInstance)
			}

			actionMaker.AssertExpectations(t)
			serviceMaker.AssertExpectations(t)
			scriptsMaker.AssertExpectations(t)
			envMaker.AssertExpectations(t)
		})
	}
}

func Test_nativeInstance_InstanceTimeout(t *testing.T) {
	instance := &nativeInstance{
		name:            "testInstance",
		instanceTimeout: 10 * time.Second,
		workspace:       "/fake/workspace",
	}
	assert.Equal(t, 10*time.Second, instance.InstanceTimeout())
}

func Test_nativeInstance_ActionTimeout(t *testing.T) {
	instance := &nativeInstance{
		name:          "testInstance",
		actionTimeout: 5000,
		workspace:     "/fake/workspace",
	}
	assert.Equal(t, 5000, instance.ActionTimeout())
}

func Test_nativeInstance_ConfigActions(t *testing.T) {
	instance := &nativeInstance{
		name:      "testInstance",
		workspace: "/fake/workspace",
		configActions: []types.Action{
			types.StartAction{},
			types.StopAction{},
		},
	}
	assert.Equal(t, instance.configActions, instance.ConfigActions())
}

func Test_nativeInstance_ConfigServices(t *testing.T) {
	instance := &nativeInstance{
		name:      "testInstance",
		workspace: "/fake/workspace",
		configServices: map[string]types.Service{
			"svc": {
				Server: types.ServiceServer{
					Name: "svr",
				},
			},
		},
	}
	assert.Equal(t, instance.configServices, instance.ConfigServices())
}

func Test_nativeInstance_ConfigInstanceEnvs(t *testing.T) {
	instance := &nativeInstance{
		name:      "testInstance",
		workspace: "/fake/workspace",
		configInstanceEnvs: map[string]types.Environment{
			"local": types.LocalEnvironment{Ports: types.EnvironmentPorts{
				Start: 10000,
			}},
		},
	}
	assert.Equal(t, instance.configInstanceEnvs, instance.ConfigInstanceEnvs())
}

func Test_nativeInstance_ConfigResources(t *testing.T) {
	instance := &nativeInstance{
		name:      "testInstance",
		workspace: "/fake/workspace",
		configResources: types.Resources{
			Scripts: map[string]types.Script{
				"s1": {
					Content: "abc",
				},
			},
		},
	}
	assert.Equal(t, instance.configResources, instance.ConfigResources())
}

func Test_nativeInstance_Parameters(t *testing.T) {
	instance := &nativeInstance{
		name:      "testInstance",
		workspace: "/fake/workspace",
		params: parameters.Parameters{
			"p": parameterMocks.NewMockParameter(t),
		},
	}
	assert.Equal(t, instance.params, instance.Parameters())
}

func Test_nativeInstance_IsChild(t *testing.T) {
	instance := &nativeInstance{
		name:       "testInstance",
		workspace:  "/fake/workspace",
		extendName: "anotherInstance",
	}
	assert.True(t, instance.IsChild())
	instance = &nativeInstance{
		name:      "testInstance",
		workspace: "/fake/workspace",
	}
	assert.False(t, instance.IsChild())
}

func Test_nativeInstance_Abstract(t *testing.T) {
	instance := &nativeInstance{
		name:      "testInstance",
		workspace: "/fake/workspace",
		abstract:  true,
	}
	assert.True(t, instance.IsAbstract())
	instance = &nativeInstance{
		name: "testInstance",
	}
	assert.False(t, instance.IsAbstract())
}

func Test_nativeInstance_Extend(t *testing.T) {
	action1 := &types.StartAction{Service: "test"}
	action2 := &types.StopAction{Service: "test"}
	svc1 := types.Service{
		Server: types.ServiceServer{Name: "fpm"},
	}
	svc2 := types.Service{
		Server: types.ServiceServer{Name: "nginx"},
	}
	env1 := types.LocalEnvironment{Ports: types.EnvironmentPorts{Start: 10000}}
	env2 := types.DockerEnvironment{Ports: types.EnvironmentPorts{Start: 20000}}
	script1 := types.Script{Content: "test content 1"}
	script2 := types.Script{Content: "test content 2"}

	paramParent := parameterMocks.NewMockParameter(t)
	paramChild := parameterMocks.NewMockParameter(t)
	paramExtended := parameterMocks.NewMockParameter(t)

	tests := []struct {
		name                       string
		parent                     *nativeInstance
		child                      *nativeInstance
		instsMap                   map[string]Instance
		expectedConfigActions      []types.Action
		expectedConfigServices     map[string]types.Service
		expectedConfigEnvs         map[string]types.Environment
		expectedConfigInstanceEnvs map[string]types.Environment
		expectedConfigResources    types.Resources
		expectedParams             parameters.Parameters
		expectedInstanceTimeout    time.Duration
		expectedActionTimeout      int
		expectedError              string
	}{
		{
			name: "successful extend with parameter merging",
			parent: &nativeInstance{
				name: "parentInstance",
				configActions: []types.Action{
					action1,
					action2,
				},
				configServices: map[string]types.Service{
					"s1": svc1,
					"s2": svc2,
				},
				configInstanceEnvs: map[string]types.Environment{
					"local":  env1,
					"docker": env2,
				},
				configResources: types.Resources{
					Scripts: map[string]types.Script{
						"s1": script1,
						"s2": script2,
					},
				},
				params: parameters.Parameters{
					"parent_key": paramParent,
				},
				instanceTimeout: 15 * time.Second,
				actionTimeout:   10000,
			},
			child: &nativeInstance{
				name:       "childInstance",
				extendName: "parentInstance",
				extendParams: parameters.Parameters{
					"extended_key": paramExtended,
				},
				configActions: make([]types.Action, 0),
				params: parameters.Parameters{
					"child_key": paramChild,
				},
				instanceTimeoutDefault: true,
				actionTimeoutDefault:   true,
			},
			expectedConfigActions: []types.Action{
				action1,
				action2,
			},
			expectedConfigServices: map[string]types.Service{
				"s1": svc1,
				"s2": svc2,
			},
			expectedConfigInstanceEnvs: map[string]types.Environment{
				"local":  env1,
				"docker": env2,
			},
			expectedConfigResources: types.Resources{
				Scripts: map[string]types.Script{
					"s1": script1,
					"s2": script2,
				},
			},
			expectedParams: parameters.Parameters{
				"parent_key":   paramParent,
				"child_key":    paramChild,
				"extended_key": paramExtended,
			},
			expectedInstanceTimeout: 15 * time.Second,
			expectedActionTimeout:   10000,
		},
		{
			name: "successful skip extend if all defined",
			parent: &nativeInstance{
				name: "parentInstance",
				configActions: []types.Action{
					action1,
					action2,
				},
				params: parameters.Parameters{
					"parent_key": paramParent,
				},
				instanceTimeout: 15 * time.Second,
			},
			child: &nativeInstance{
				name: "childInstance",
				configActions: []types.Action{
					action1,
				},
				configServices: map[string]types.Service{
					"s1": svc1,
				},
				configInstanceEnvs: map[string]types.Environment{
					"local": env1,
				},
				configResources: types.Resources{
					Scripts: map[string]types.Script{
						"s1": script1,
					},
				},
				extendName: "parentInstance",
				extendParams: parameters.Parameters{
					"extended_key": paramExtended,
				},
				params: parameters.Parameters{
					"child_key": paramChild,
				},
				instanceTimeoutDefault: false,
				instanceTimeout:        5 * time.Second,
				actionTimeoutDefault:   false,
				actionTimeout:          2000,
			},
			expectedConfigActions: []types.Action{
				action1,
			},
			expectedConfigServices: map[string]types.Service{
				"s1": svc1,
			},
			expectedConfigInstanceEnvs: map[string]types.Environment{
				"local": env1,
			},
			expectedConfigResources: types.Resources{
				Scripts: map[string]types.Script{
					"s1": script1,
				},
			},
			expectedParams: parameters.Parameters{
				"child_key": paramChild,
			},
			expectedInstanceTimeout: 5 * time.Second,
			expectedActionTimeout:   2000,
		},
		{
			name:   "missing parent instance",
			parent: nil,
			child: &nativeInstance{
				name:       "childInstance",
				extendName: "missingParentInstance",
			},
			expectedError: "instance missingParentInstance not found",
		},
		{
			name: "circular inheritance",
			parent: &nativeInstance{
				name:       "parentInstance",
				extendName: "childInstance",
			},
			child: &nativeInstance{
				name:       "childInstance",
				extendName: "parentInstance",
			},
			expectedError: "instance childInstance already extending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instsMap := make(map[string]Instance)
			if tt.parent != nil {
				instsMap["parentInstance"] = tt.parent
			}
			if tt.child != nil {
				instsMap["childInstance"] = tt.child
			}

			err := tt.child.Extend(instsMap)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedConfigActions, tt.child.ConfigActions())
				assert.Equal(t, tt.expectedConfigInstanceEnvs, tt.child.ConfigInstanceEnvs())
				assert.Equal(t, tt.expectedConfigServices, tt.child.ConfigServices())
				assert.Equal(t, tt.expectedConfigResources, tt.child.ConfigResources())
				assert.Equal(t, tt.expectedParams, tt.child.Parameters())
				assert.Equal(t, tt.expectedInstanceTimeout, tt.child.InstanceTimeout())
				assert.Equal(t, tt.expectedActionTimeout, tt.child.ActionTimeout())
			}
		})
	}
}

func Test_nativeInstance_Init(t *testing.T) {
	testScripts := map[string]types.Script{
		"test-script": {
			Content:    "test content",
			Path:       "/path/to/script",
			Mode:       "0644",
			Parameters: nil,
		},
	}
	testInstanceEnvironments := map[string]types.Environment{
		"local": types.LocalEnvironment{
			Ports: types.EnvironmentPorts{Start: 9000},
		},
	}
	testSpecEnvironments := map[string]types.Environment{
		"local": types.LocalEnvironment{
			Ports: types.EnvironmentPorts{Start: 8000},
		},
	}
	testEnvironments := environments.Environments{
		providers.LocalType: environmentMocks.NewMockEnvironment(t),
	}
	testServices := map[string]types.Service{
		"svc": {
			Server: types.ServiceServer{
				Name:    "test",
				Sandbox: "local",
			},
		},
	}
	testRunnableService := services.Services{
		"svc": servicesMocks.NewMockService(t),
	}
	testServers := servers.Servers{
		"t1": map[string]servers.Server{
			"t2": serversMocks.NewMockServer(t),
		},
	}
	testScriptResources := scripts.Scripts{
		"test-script": scriptsMocks.NewMockScript(t),
	}
	testActions := []types.Action{
		types.StartAction{
			Service: "s1",
		},
		types.BenchAction{
			Service: "btest",
			Id:      "bt",
		},
		types.StopAction{
			Service: "s1",
		},
	}
	testDefaultsParams := parameters.Parameters{
		"default_key": parameterMocks.NewMockParameter(t),
	}
	testParams := types.Parameters{
		"test_key": "test_value",
	}
	testInstanceParams := parameters.Parameters{
		"ikey": parameterMocks.NewMockParameter(t),
	}
	instanceIdx := 1
	tests := []struct {
		name       string
		setupMocks func(
			*testing.T,
			*actionsMocks.MockActionMaker,
			*servicesMocks.MockMaker,
			*scriptsMocks.MockMaker,
			*environmentsMocks.MockMaker,
			*defaults.Defaults,
		) []action.Action
		expectError          bool
		expectedErrorMsg     string
		instanceConfig       types.Instance
		actionTimeout        int
		envsConfig           map[string]types.Environment
		srvs                 servers.Servers
		instanceWorkspace    string
		expectedServices     services.Services
		expectedEnvironments environments.Environments
		expectedWorkspace    string
	}{
		{
			name: "successful initialization",
			setupMocks: func(
				t *testing.T,
				actionMaker *actionsMocks.MockActionMaker,
				serviceMaker *servicesMocks.MockMaker,
				scriptMaker *scriptsMocks.MockMaker,
				envMaker *environmentsMocks.MockMaker,
				dflts *defaults.Defaults,
			) []action.Action {
				scriptMaker.On("Make", testScripts).Return(testScriptResources, nil)
				envMaker.On(
					"Make",
					testSpecEnvironments,
					testInstanceEnvironments,
					"/workspace/test-instance",
				).Return(testEnvironments, nil)
				sl := servicesMocks.NewMockServiceLocator(t)
				serviceMaker.On(
					"Make",
					testServices,
					dflts,
					testScriptResources,
					testServers,
					testEnvironments,
					"test-instance",
					instanceIdx,
					"/workspace/test-instance",
					testInstanceParams,
				).Return(sl, nil)
				sl.On("Services").Return(testRunnableService)
				acts := []action.Action{
					actionMocks.NewMockAction(t),
					actionMocks.NewMockAction(t),
					actionMocks.NewMockAction(t),
				}
				actionMaker.On("MakeAction", testActions[0], sl, 5000).Return(acts[0], nil)
				actionMaker.On("MakeAction", testActions[1], sl, 5000).Return(acts[1], nil)
				actionMaker.On("MakeAction", testActions[2], sl, 5000).Return(acts[2], nil)
				return acts
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			actionTimeout:        5000,
			envsConfig:           testSpecEnvironments,
			srvs:                 testServers,
			instanceWorkspace:    "/workspace",
			expectedServices:     testRunnableService,
			expectedEnvironments: testEnvironments,
			expectedWorkspace:    "/workspace/test-instance",
		},
		{
			name: "error initialization due to action fail",
			setupMocks: func(
				t *testing.T,
				actionMaker *actionsMocks.MockActionMaker,
				serviceMaker *servicesMocks.MockMaker,
				scriptMaker *scriptsMocks.MockMaker,
				envMaker *environmentsMocks.MockMaker,
				dflts *defaults.Defaults,
			) []action.Action {
				scriptMaker.On("Make", testScripts).Return(testScriptResources, nil)
				envMaker.On(
					"Make",
					testSpecEnvironments,
					testInstanceEnvironments,
					"/workspace/test-instance",
				).Return(testEnvironments, nil)
				sl := servicesMocks.NewMockServiceLocator(t)
				serviceMaker.On(
					"Make",
					testServices,
					dflts,
					testScriptResources,
					testServers,
					testEnvironments,
					"test-instance",
					instanceIdx,
					"/workspace/test-instance",
					testInstanceParams,
				).Return(sl, nil)
				sl.On("Services").Return(testRunnableService)
				acts := []action.Action{
					actionMocks.NewMockAction(t),
					actionMocks.NewMockAction(t),
					actionMocks.NewMockAction(t),
				}
				actionMaker.On("MakeAction", testActions[0], sl, 5000).Return(acts[0], nil)
				actionMaker.On("MakeAction", testActions[1], sl, 5000).Return(acts[1], errors.New("action fail"))
				return nil
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			actionTimeout:     5000,
			envsConfig:        testSpecEnvironments,
			srvs:              testServers,
			instanceWorkspace: "/workspace",
			expectError:       true,
			expectedErrorMsg:  "action fail",
		},
		{
			name: "error initialization due to service fail",
			setupMocks: func(
				t *testing.T,
				actionMaker *actionsMocks.MockActionMaker,
				serviceMaker *servicesMocks.MockMaker,
				scriptMaker *scriptsMocks.MockMaker,
				envMaker *environmentsMocks.MockMaker,
				dflts *defaults.Defaults,
			) []action.Action {
				scriptMaker.On("Make", testScripts).Return(testScriptResources, nil)
				envMaker.On(
					"Make",
					testSpecEnvironments,
					testInstanceEnvironments,
					"/workspace/test-instance",
				).Return(testEnvironments, nil)
				serviceMaker.On(
					"Make",
					testServices,
					dflts,
					testScriptResources,
					testServers,
					testEnvironments,
					"test-instance",
					instanceIdx,
					"/workspace/test-instance",
					testInstanceParams,
				).Return(nil, errors.New("svc fail"))
				return nil
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Timeouts: types.InstanceTimeouts{
					Action:  5000,
					Actions: 10000,
				},
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			actionTimeout:     5000,
			envsConfig:        testSpecEnvironments,
			srvs:              testServers,
			instanceWorkspace: "/workspace",
			expectError:       true,
			expectedErrorMsg:  "svc fail",
		},
		{
			name: "error initialization due to env fail",
			setupMocks: func(
				t *testing.T,
				actionMaker *actionsMocks.MockActionMaker,
				serviceMaker *servicesMocks.MockMaker,
				scriptMaker *scriptsMocks.MockMaker,
				envMaker *environmentsMocks.MockMaker,
				dflts *defaults.Defaults,
			) []action.Action {
				scriptMaker.On("Make", testScripts).Return(testScriptResources, nil)
				envMaker.On(
					"Make",
					testSpecEnvironments,
					testInstanceEnvironments,
					"/workspace/test-instance",
				).Return(testEnvironments, errors.New("env fail"))
				return nil
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Timeouts: types.InstanceTimeouts{
					Action:  5000,
					Actions: 10000,
				},
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			actionTimeout:     5000,
			envsConfig:        testSpecEnvironments,
			srvs:              testServers,
			instanceWorkspace: "/workspace",
			expectError:       true,
			expectedErrorMsg:  "env fail",
		},
		{
			name: "error initialization due to script fail",
			setupMocks: func(
				t *testing.T,
				actionMaker *actionsMocks.MockActionMaker,
				serviceMaker *servicesMocks.MockMaker,
				scriptMaker *scriptsMocks.MockMaker,
				envMaker *environmentsMocks.MockMaker,
				dflts *defaults.Defaults,
			) []action.Action {
				scriptMaker.On("Make", testScripts).Return(nil, errors.New("script fail"))
				return nil
			},
			instanceConfig: types.Instance{
				Name:    "test-instance",
				Actions: testActions,
				Timeouts: types.InstanceTimeouts{
					Action:  5000,
					Actions: 10000,
				},
				Resources: types.Resources{
					Scripts: testScripts,
				},
				Parameters:   testParams,
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			envsConfig:        testSpecEnvironments,
			srvs:              testServers,
			instanceWorkspace: "/workspace",
			expectError:       true,
			expectedErrorMsg:  "script fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			actionMaker := actionsMocks.NewMockActionMaker(t)
			serviceMaker := servicesMocks.NewMockMaker(t)
			scriptsMaker := scriptsMocks.NewMockMaker(t)
			envMaker := environmentsMocks.NewMockMaker(t)

			dflts := &defaults.Defaults{
				Service: defaults.ServiceDefaults{
					Sandbox: "local",
					Server: defaults.ServiceServerDefaults{
						Tag: "latest",
					},
				},
				Timeouts: defaults.TimeoutsDefaults{
					Actions: 15000,
					Action:  8000,
				},
				Parameters: testDefaultsParams,
			}

			inst := &nativeInstance{
				fnd:                fndMock,
				actionMaker:        actionMaker,
				servicesMaker:      serviceMaker,
				scriptsMaker:       scriptsMaker,
				environmentMaker:   envMaker,
				configActions:      tt.instanceConfig.Actions,
				configServices:     tt.instanceConfig.Services,
				configEnvs:         tt.envsConfig,
				configInstanceEnvs: tt.instanceConfig.Environments,
				configResources:    tt.instanceConfig.Resources,
				name:               "test-instance",
				index:              instanceIdx,
				specWorkspace:      "/workspace",
				params:             testInstanceParams,
				initialized:        false,
				defaults:           dflts,
				servers:            tt.srvs,
				actionTimeout:      tt.actionTimeout,
			}

			acts := tt.setupMocks(t, actionMaker, serviceMaker, scriptsMaker, envMaker, dflts)

			err := inst.Init()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, acts, inst.actions)
				assert.Equal(t, tt.expectedServices, inst.services)
				assert.Equal(t, tt.expectedEnvironments, inst.envs)
				assert.Equal(t, tt.expectedWorkspace, inst.workspace)
				assert.True(t, inst.initialized)
			}

			actionMaker.AssertExpectations(t)
			serviceMaker.AssertExpectations(t)
			scriptsMaker.AssertExpectations(t)
			envMaker.AssertExpectations(t)
		})
	}
}

func Test_nativeInstance_Workspace(t *testing.T) {
	instance := &nativeInstance{
		name:            "testInstance",
		instanceTimeout: 10 * time.Second,
		workspace:       "/fake/workspace",
	}
	assert.Equal(t, "/fake/workspace", instance.Workspace())
}

func Test_nativeInstance_Name(t *testing.T) {
	instance := &nativeInstance{
		name:            "testInstance",
		instanceTimeout: 10 * time.Second,
		workspace:       "/fake/workspace",
	}
	assert.Equal(t, "testInstance", instance.Name())
}

func Test_nativeInstance_Run(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		abstract    bool
		initialized bool
		setupMocks  func(
			*nativeInstance,
			*appMocks.MockFoundation,
			*runtimeMocks.MockMaker,
			[]*actionMocks.MockAction,
			context.CancelFunc,
		)
		expectedCancellations int
		expectError           bool
		expectedErrorMsg      string
	}{
		{
			name:        "successful run of single action",
			count:       1,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				acts[0].On("Execute", actx, inst.runData).Return(true, nil)

				localEnv.On("Destroy", ctx).Return(nil)
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectedCancellations: 2,
		},
		{
			name:        "successful run of two success actions",
			count:       2,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[0].On("Execute", actx, inst.runData).Return(true, nil)

				actTimeout = 2 * time.Second
				acts[1].On("Timeout").Return(actTimeout)
				acts[1].On("When").Return(action.OnSuccess)
				actx, cancel = context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[1].On("Execute", actx, inst.runData).Return(true, nil)

				localEnv.On("Destroy", ctx).Return(nil)
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectedCancellations: 3,
		},
		{
			name:        "successful run of two success actions with on failure",
			count:       2,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[0].On("Execute", actx, inst.runData).Return(true, nil)

				actTimeout = 2 * time.Second
				acts[1].On("When").Return(action.OnFailure)

				localEnv.On("Destroy", ctx).Return(nil)
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectedCancellations: 2,
		},
		{
			name:        "failed run of failed and success actions with on success",
			count:       2,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[0].On("Execute", actx, inst.runData).Return(false, errors.New("failed first"))

				actTimeout = 2 * time.Second
				acts[1].On("When").Return(action.OnSuccess)

				localEnv.On("Destroy", ctx).Return(nil)
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectedCancellations: 3,
			expectError:           true,
			expectedErrorMsg:      "failed first",
		},
		{
			name:        "failed run of failed and success actions with on failure",
			count:       2,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[0].On("Execute", actx, inst.runData).Return(false, errors.New("failed first"))

				actTimeout = 2 * time.Second
				acts[1].On("Timeout").Return(actTimeout)
				acts[1].On("When").Return(action.OnFailure)
				actx, cancel = context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[1].On("Execute", actx, inst.runData).Return(true, nil)

				localEnv.On("Destroy", ctx).Return(nil)
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectedCancellations: 3,
			expectError:           true,
			expectedErrorMsg:      "failed first",
		},
		{
			name:        "failed run of failed and success actions with always",
			count:       2,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[0].On("Execute", actx, inst.runData).Return(false, errors.New("failed first"))

				actTimeout = 2 * time.Second
				acts[1].On("Timeout").Return(actTimeout)
				acts[1].On("When").Return(action.Always)
				actx, cancel = context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[1].On("Execute", actx, inst.runData).Return(true, nil)

				localEnv.On("Destroy", ctx).Return(nil)
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectedCancellations: 3,
			expectError:           true,
			expectedErrorMsg:      "failed first",
		},
		{
			name:        "failed run of failed and failed actions with always",
			count:       2,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[0].On("Execute", actx, inst.runData).Return(false, errors.New("failed first"))

				actTimeout = 2 * time.Second
				acts[1].On("Timeout").Return(actTimeout)
				acts[1].On("When").Return(action.Always)
				actx, cancel = context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)
				acts[1].On("Execute", actx, inst.runData).Return(false, errors.New("failed second"))

				localEnv.On("Destroy", ctx).Return(nil)
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectedCancellations: 3,
			expectError:           true,
			expectedErrorMsg:      "failed first",
		},
		{
			name:        "fail on env destroy",
			count:       1,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				acts[0].On("Execute", actx, inst.runData).Return(true, nil)

				localEnv.On("Destroy", ctx).Return(errors.New("local destroy err"))
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "local destroy err",
		},
		{
			name:        "fail on action error",
			count:       1,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				acts[0].On("Execute", actx, inst.runData).Return(false, errors.New("action err"))

				localEnv.On("Destroy", ctx).Return(errors.New("local destroy err"))
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "action err",
		},
		{
			name:        "fail on action false return",
			count:       1,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				acts[0].On("Execute", actx, inst.runData).Return(false, errors.New("action err"))

				localEnv.On("Destroy", ctx).Return(errors.New("local destroy err"))
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "action err",
		},
		{
			name:        "fail on action false return",
			count:       1,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.instanceTimeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				acts[0].On("Timeout").Return(actTimeout)
				acts[0].On("When").Return(action.OnSuccess)
				actx, cancel := context.WithTimeout(ctx, inst.instanceTimeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				acts[0].On("Execute", actx, inst.runData).Return(false, nil)

				localEnv.On("Destroy", ctx).Return(errors.New("local destroy err"))
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "action execution failed",
		},
		{
			name:        "fail on env init fail",
			count:       1,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(nil)
				fnd.On("Fs").Return(fsMock)

				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Maybe().Return(true)
				localEnv.On("Init", ctx).Maybe().Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(errors.New("docker fail"))

				localEnv.On("Destroy", ctx).Maybe().Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "docker fail",
		},
		{
			name:        "fail on removing workspace",
			count:       1,
			initialized: true,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("RemoveAll", "/fake/workspace").Return(errors.New("remove fail"))
				fnd.On("Fs").Return(fsMock)
			},
			expectError:      true,
			expectedErrorMsg: "failed to remove previous workspace for instance testInstance: remove fail",
		},
		{
			name:        "fail on not initialized action",
			abstract:    false,
			initialized: false,
			count:       1,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
			},
			expectError:      true,
			expectedErrorMsg: "instance testInstance is not initialized and cannot be run",
		},
		{
			name:        "fail on abstract action",
			abstract:    true,
			initialized: true,
			count:       1,
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				acts []*actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
			},
			expectError:      true,
			expectedErrorMsg: "instance testInstance is abstract and cannot be run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := externalMocks.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			runtimeMakerMock := runtimeMocks.NewMockMaker(t)
			envMock := environmentMocks.NewMockEnvironment(t)
			actMocks := make([]*actionMocks.MockAction, tt.count)
			acts := make([]action.Action, tt.count)
			for i := 0; i < tt.count; i++ {
				actMocks[i] = actionMocks.NewMockAction(t)
				acts[i] = actMocks[i]
			}

			cancelCalled := 0
			cancelFunc := func() { cancelCalled++ }

			instance := &nativeInstance{
				fnd:          fndMock,
				runtimeMaker: runtimeMakerMock,
				name:         "testInstance",
				actions:      acts,
				abstract:     tt.abstract,
				initialized:  tt.initialized,
				envs: environments.Environments{
					providers.LocalType:  environmentMocks.NewMockEnvironment(t),
					providers.DockerType: environmentMocks.NewMockEnvironment(t),
				},
				runData:         runtimeMocks.NewMockData(t),
				instanceTimeout: 10 * time.Second,
				workspace:       "/fake/workspace",
			}

			tt.setupMocks(instance, fndMock, runtimeMakerMock, actMocks, cancelFunc)

			err := instance.Run()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedCancellations, cancelCalled)
			}

			envMock.AssertExpectations(t)
			for _, actMock := range actMocks {
				actMock.AssertExpectations(t)
			}
			runtimeMakerMock.AssertExpectations(t)
			fndMock.AssertExpectations(t)
		})
	}
}
