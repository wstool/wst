package instances

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	actionsMocks "github.com/bukka/wst/mocks/generated/run/actions"
	actionMocks "github.com/bukka/wst/mocks/generated/run/actions/action"
	environmentsMocks "github.com/bukka/wst/mocks/generated/run/environments"
	environmentMocks "github.com/bukka/wst/mocks/generated/run/environments/environment"
	expectationsMocks "github.com/bukka/wst/mocks/generated/run/expectations"
	runtimeMocks "github.com/bukka/wst/mocks/generated/run/instances/runtime"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	scriptsMocks "github.com/bukka/wst/mocks/generated/run/resources/scripts"
	serversMocks "github.com/bukka/wst/mocks/generated/run/servers"
	servicesMocks "github.com/bukka/wst/mocks/generated/run/services"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/environments"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/servers"
	"github.com/stretchr/testify/assert"
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
	testData := runtimeMocks.NewMockData(t)
	tests := []struct {
		name       string
		setupMocks func(
			*testing.T,
			*actionsMocks.MockActionMaker,
			*servicesMocks.MockMaker,
			*scriptsMocks.MockMaker,
			*environmentsMocks.MockMaker,
			*runtimeMocks.MockMaker,
		) []action.Action
		expectError         bool
		expectedErrorMsg    string
		instanceConfig      types.Instance
		envsConfig          map[string]types.Environment
		srvs                servers.Servers
		instanceWorkspace   string
		getExpectedInstance func(fndMock *appMocks.MockFoundation, acts []action.Action) *nativeInstance
	}{
		{
			name: "successful creation",
			setupMocks: func(
				t *testing.T,
				actionMaker *actionsMocks.MockActionMaker,
				serviceMaker *servicesMocks.MockMaker,
				scriptMaker *scriptsMocks.MockMaker,
				envMaker *environmentsMocks.MockMaker,
				runtimeMaker *runtimeMocks.MockMaker,
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
					testScriptResources,
					testServers,
					testEnvironments,
					"test-instance",
					"/workspace/test-instance",
				).Return(sl, nil)
				acts := []action.Action{
					actionMocks.NewMockAction(t),
					actionMocks.NewMockAction(t),
					actionMocks.NewMockAction(t),
				}
				actionMaker.On("MakeAction", testActions[0], sl, 5000).Return(acts[0], nil)
				actionMaker.On("MakeAction", testActions[1], sl, 5000).Return(acts[1], nil)
				actionMaker.On("MakeAction", testActions[2], sl, 5000).Return(acts[2], nil)
				runtimeMaker.On("Make").Return(testData)
				return acts
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
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			envsConfig:        testSpecEnvironments,
			srvs:              testServers,
			instanceWorkspace: "/workspace",
			getExpectedInstance: func(fndMock *appMocks.MockFoundation, acts []action.Action) *nativeInstance {
				return &nativeInstance{
					fnd:       fndMock,
					name:      "test-instance",
					timeout:   10 * time.Second,
					actions:   acts,
					envs:      testEnvironments,
					runData:   testData,
					workspace: "/workspace/test-instance",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			actionMaker := actionsMocks.NewMockActionMaker(t)
			serviceMaker := servicesMocks.NewMockMaker(t)
			scriptsMaker := scriptsMocks.NewMockMaker(t)
			envMaker := environmentsMocks.NewMockMaker(t)
			runtimeMaker := runtimeMocks.NewMockMaker(t)

			maker := &nativeInstanceMaker{
				fnd:              fndMock,
				actionMaker:      actionMaker,
				servicesMaker:    serviceMaker,
				scriptsMaker:     scriptsMaker,
				environmentMaker: envMaker,
				runtimeMaker:     runtimeMaker,
			}

			acts := tt.setupMocks(t, actionMaker, serviceMaker, scriptsMaker, envMaker, runtimeMaker)

			got, err := maker.Make(tt.instanceConfig, tt.envsConfig, tt.srvs, tt.instanceWorkspace)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				actualInstance, ok := got.(*nativeInstance)
				assert.True(t, ok)
				expectedInstance := tt.getExpectedInstance(fndMock, acts)
				assert.Equal(t, expectedInstance, actualInstance)
			}

			actionMaker.AssertExpectations(t)
			serviceMaker.AssertExpectations(t)
			scriptsMaker.AssertExpectations(t)
			envMaker.AssertExpectations(t)
		})
	}
}
