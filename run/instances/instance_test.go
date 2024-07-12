package instances

import (
	"context"
	"github.com/bukka/wst/conf/types"
	externalMocks "github.com/bukka/wst/mocks/authored/external"
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
	"github.com/pkg/errors"
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
		getExpectedInstance func(
			fndMock *appMocks.MockFoundation,
			acts []action.Action,
			runtimeMaker *runtimeMocks.MockMaker,
		) *nativeInstance
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
				runtimeMaker.On("MakeData").Return(testData)
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
			getExpectedInstance: func(
				fndMock *appMocks.MockFoundation,
				acts []action.Action,
				runtimeMaker *runtimeMocks.MockMaker,
			) *nativeInstance {
				return &nativeInstance{
					fnd:          fndMock,
					runtimeMaker: runtimeMaker,
					name:         "test-instance",
					timeout:      10 * time.Second,
					actions:      acts,
					envs:         testEnvironments,
					runData:      testData,
					workspace:    "/workspace/test-instance",
				}
			},
		},
		{
			name: "error creation due to action fail",
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
				actionMaker.On("MakeAction", testActions[1], sl, 5000).Return(acts[1], errors.New("action fail"))
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
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			envsConfig:        testSpecEnvironments,
			srvs:              testServers,
			instanceWorkspace: "/workspace",
			expectError:       true,
			expectedErrorMsg:  "action fail",
		},
		{
			name: "error creation due to action fail",
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
				serviceMaker.On(
					"Make",
					testServices,
					testScriptResources,
					testServers,
					testEnvironments,
					"test-instance",
					"/workspace/test-instance",
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
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			envsConfig:        testSpecEnvironments,
			srvs:              testServers,
			instanceWorkspace: "/workspace",
			expectError:       true,
			expectedErrorMsg:  "svc fail",
		},
		{
			name: "error creation due to action fail",
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
				Environments: testInstanceEnvironments,
				Services:     testServices,
			},
			envsConfig:        testSpecEnvironments,
			srvs:              testServers,
			instanceWorkspace: "/workspace",
			expectError:       true,
			expectedErrorMsg:  "env fail",
		},
		{
			name: "error creation due to action fail",
			setupMocks: func(
				t *testing.T,
				actionMaker *actionsMocks.MockActionMaker,
				serviceMaker *servicesMocks.MockMaker,
				scriptMaker *scriptsMocks.MockMaker,
				envMaker *environmentsMocks.MockMaker,
				runtimeMaker *runtimeMocks.MockMaker,
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
				expectedInstance := tt.getExpectedInstance(fndMock, acts, runtimeMaker)
				assert.Equal(t, expectedInstance, actualInstance)
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
		name:      "testInstance",
		timeout:   10 * time.Second,
		workspace: "/fake/workspace",
	}
	assert.Equal(t, "/fake/workspace", instance.Workspace())
}

func Test_nativeInstance_Name(t *testing.T) {
	instance := &nativeInstance{
		name:      "testInstance",
		timeout:   10 * time.Second,
		workspace: "/fake/workspace",
	}
	assert.Equal(t, "testInstance", instance.Name())
}

func Test_nativeInstance_Run(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(
			*nativeInstance,
			*appMocks.MockFoundation,
			*runtimeMocks.MockMaker,
			*actionMocks.MockAction,
			context.CancelFunc,
		)
		expectedCancellations int
		expectError           bool
		expectedErrorMsg      string
	}{
		{
			name: "successful run",
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				act *actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.timeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				act.On("Timeout").Return(actTimeout)
				actx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				act.On("Execute", actx, inst.runData).Return(true, nil)

				localEnv.On("Destroy", ctx).Return(nil)
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectedCancellations: 2,
		},
		{
			name: "fail on env destroy",
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				act *actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.timeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				act.On("Timeout").Return(actTimeout)
				actx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				act.On("Execute", actx, inst.runData).Return(true, nil)

				localEnv.On("Destroy", ctx).Return(errors.New("local destroy err"))
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "local destroy err",
		},
		{
			name: "fail on action error",
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				act *actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.timeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				act.On("Timeout").Return(actTimeout)
				actx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				act.On("Execute", actx, inst.runData).Return(false, errors.New("action err"))

				localEnv.On("Destroy", ctx).Return(errors.New("local destroy err"))
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "action err",
		},
		{
			name: "fail on action false return",
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				act *actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.timeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				act.On("Timeout").Return(actTimeout)
				actx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				act.On("Execute", actx, inst.runData).Return(false, errors.New("action err"))

				localEnv.On("Destroy", ctx).Return(errors.New("local destroy err"))
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "action err",
		},
		{
			name: "fail on action false return",
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				act *actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(nil)

				tctx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", ctx, inst.timeout).Return(tctx, cancelFunc)

				actTimeout := 1 * time.Second
				act.On("Timeout").Return(actTimeout)
				actx, cancel := context.WithTimeout(ctx, inst.timeout)
				defer cancel()
				rm.On("MakeContextWithTimeout", tctx, actTimeout).Return(actx, cancelFunc)

				act.On("Execute", actx, inst.runData).Return(false, nil)

				localEnv.On("Destroy", ctx).Return(errors.New("local destroy err"))
				dockerEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "action execution failed",
		},
		{
			name: "fail on env init fail",
			setupMocks: func(
				inst *nativeInstance,
				fnd *appMocks.MockFoundation,
				rm *runtimeMocks.MockMaker,
				act *actionMocks.MockAction,
				cancelFunc context.CancelFunc,
			) {
				ctx := context.Background()
				rm.On("MakeBackgroundContext").Return(ctx)

				localEnv := inst.envs[providers.LocalType].(*environmentMocks.MockEnvironment)
				localEnv.On("IsUsed").Return(true)
				localEnv.On("Init", ctx).Return(nil)

				dockerEnv := inst.envs[providers.DockerType].(*environmentMocks.MockEnvironment)
				dockerEnv.On("IsUsed").Return(true)
				dockerEnv.On("Init", ctx).Return(errors.New("docker fail"))

				localEnv.On("Destroy", ctx).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "docker fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := externalMocks.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			runtimeMakerMock := runtimeMocks.NewMockMaker(t)
			envMock := environmentMocks.NewMockEnvironment(t)
			actionMock := actionMocks.NewMockAction(t)

			cancelCalled := 0
			cancelFunc := func() { cancelCalled++ }

			instance := &nativeInstance{
				fnd:          fndMock,
				runtimeMaker: runtimeMakerMock,
				name:         "testInstance",
				actions:      []action.Action{actionMock},
				envs: environments.Environments{
					providers.LocalType:  environmentMocks.NewMockEnvironment(t),
					providers.DockerType: environmentMocks.NewMockEnvironment(t),
				},
				runData:   runtimeMocks.NewMockData(t),
				timeout:   10 * time.Second,
				workspace: "/fake/workspace",
			}

			tt.setupMocks(instance, fndMock, runtimeMakerMock, actionMock, cancelFunc)

			err := instance.Run()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedCancellations, cancelCalled)
			}

			envMock.AssertExpectations(t)
			actionMock.AssertExpectations(t)
			runtimeMakerMock.AssertExpectations(t)
			fndMock.AssertExpectations(t)
		})
	}
}
