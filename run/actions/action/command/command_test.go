package command

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	outputMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/output"
	runtimeMocks "github.com/wstool/wst/mocks/generated/run/instances/runtime"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	servicesMocks "github.com/wstool/wst/mocks/generated/run/services"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/services"
	"testing"
	"time"
)

func TestCreateActionMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	tests := []struct {
		name string
		fnd  app.Foundation
	}{
		{
			name: "create maker",
			fnd:  fndMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateActionMaker(tt.fnd)
			assert.Equal(t, tt.fnd, got.fnd)
		})
	}
}

func TestActionMaker_Make(t *testing.T) {
	tests := []struct {
		name              string
		config            *types.CommandAction
		defaultTimeout    int
		setupMocks        func(*testing.T, *servicesMocks.MockServiceLocator) services.Service
		getExpectedAction func(*appMocks.MockFoundation, services.Service, output.Maker) *Action
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful shell command creation with default timeout",
			config: &types.CommandAction{
				Service: "validService",
				Shell:   "/bin/bash",
				Command: &types.ShellCommand{
					Command: "echo hello",
				},
				Timeout: 0,
				When:    "on_success",
				Id:      "shell-cmd",
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Service {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", "validService").Return(svc, nil)
				return svc
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svc services.Service, outMaker output.Maker) *Action {
				return &Action{
					fnd:        fndMock,
					service:    svc,
					parameters: parameters.Parameters{},
					timeout:    5000 * time.Millisecond,
					when:       action.OnSuccess,
					id:         "shell-cmd",
					command: &environment.Command{
						Name: "/bin/bash",
						Args: []string{"-c", "echo hello"},
					},
					outputMaker: outMaker,
				}
			},
		},
		{
			name: "successful args command creation with config timeout",
			config: &types.CommandAction{
				Service: "validService",
				Command: &types.ArgsCommand{
					Args: []string{"ls", "-la"},
				},
				Timeout: 3000,
				When:    "on_success",
				Id:      "args-cmd",
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Service {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", "validService").Return(svc, nil)
				return svc
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svc services.Service, outMaker output.Maker) *Action {
				return &Action{
					fnd:        fndMock,
					service:    svc,
					parameters: parameters.Parameters{},
					timeout:    3000 * time.Millisecond,
					when:       action.OnSuccess,
					id:         "args-cmd",
					command: &environment.Command{
						Name: "ls",
						Args: []string{"-la"},
					},
					outputMaker: outMaker,
				}
			},
		},
		{
			name: "failure - service not found",
			config: &types.CommandAction{
				Service: "invalidService",
				Command: &types.ArgsCommand{
					Args: []string{"ls"},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Service {
				sl.On("Find", "invalidService").Return(nil, errors.New("service not found"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "service not found",
		},
		{
			name: "failure - empty args command",
			config: &types.CommandAction{
				Service: "validService",
				Command: &types.ArgsCommand{
					Args: []string{},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Service {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", "validService").Return(svc, nil)
				return svc
			},
			expectError:      true,
			expectedErrorMsg: "ArgsCommand requires at least one argument",
		},
		{
			name: "failure - unsupported command type",
			config: &types.CommandAction{
				Service: "validService",
				Command: &struct{}{}, // some arbitrary struct that doesn't implement known command types
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Service {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", "validService").Return(svc, nil)
				return svc
			},
			expectError:      true,
			expectedErrorMsg: "unsupported command type: *struct {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			slMock := servicesMocks.NewMockServiceLocator(t)
			outMakerMock := outputMocks.NewMockMaker(t)

			m := &ActionMaker{
				fnd:         fndMock,
				outputMaker: outMakerMock,
			}

			svc := tt.setupMocks(t, slMock)

			got, err := m.Make(tt.config, slMock, tt.defaultTimeout)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualAction, ok := got.(*Action)
				assert.True(t, ok)
				expectedAction := tt.getExpectedAction(fndMock, svc, outMakerMock)
				assert.Equal(t, expectedAction, actualAction)
			}
		})
	}
}

func TestAction_Execute(t *testing.T) {
	params := parameters.Parameters{
		"key": parameterMocks.NewMockParameter(t),
	}

	tests := []struct {
		name           string
		actionID       string
		command        *environment.Command
		parameters     parameters.Parameters
		renderTemplate bool
		setupMocks     func(t *testing.T, ctx context.Context, rd *runtimeMocks.MockData, fnd *appMocks.MockFoundation, svc *servicesMocks.MockService, outMaker *outputMocks.MockMaker, collector *outputMocks.MockCollector)
		contextSetup   func() context.Context
		want           bool
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "successful execution without template rendering",
			actionID:       "test-action-no-template",
			renderTemplate: false,
			command: &environment.Command{
				Name: "test",
				Args: []string{"-a", "-b"},
			},
			parameters: params,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				outMaker *outputMocks.MockMaker,
				collector *outputMocks.MockCollector,
			) {
				outMaker.On("MakeCollector", "action test-action-no-template").Return(collector).Once()

				// No template rendering calls expected since renderTemplate is false

				expectedCmd := &environment.Command{
					Name: "test",
					Args: []string{"-a", "-b"},
				}
				svc.On("ExecCommand", ctx, expectedCmd, collector).Return(nil).Once()
				rd.On("Store", "command/test-action-no-template", collector).Return(nil).Once()
			},
			want: true,
		},
		{
			name:           "successful execution with template rendering",
			actionID:       "test-action",
			renderTemplate: true,
			command: &environment.Command{
				Name: "test",
				Args: []string{"-a", "-b"},
			},
			parameters: params,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				outMaker *outputMocks.MockMaker,
				collector *outputMocks.MockCollector,
			) {
				outMaker.On("MakeCollector", "action test-action").Return(collector).Once()

				// Mock template rendering
				svc.On("RenderTemplate", "test", params).Return("test", nil).Once()
				svc.On("RenderTemplate", "-a", params).Return("-a", nil).Once()
				svc.On("RenderTemplate", "-b", params).Return("-b", nil).Once()

				expectedCmd := &environment.Command{
					Name: "test",
					Args: []string{"-a", "-b"},
				}
				svc.On("ExecCommand", ctx, expectedCmd, collector).Return(nil).Once()
				rd.On("Store", "command/test-action", collector).Return(nil).Once()
			},
			want: true,
		},
		{
			name:           "template rendering error in command name",
			actionID:       "template-error",
			renderTemplate: true,
			command: &environment.Command{
				Name: "test-{{.invalid}}",
				Args: []string{"-a"},
			},
			parameters: params,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				outMaker *outputMocks.MockMaker,
				collector *outputMocks.MockCollector,
			) {
				outMaker.On("MakeCollector", "action template-error").Return(collector).Once()
				svc.On("RenderTemplate", "test-{{.invalid}}", params).Return("", errors.New("template rendering error")).Once()
			},
			want:           false,
			expectError:    true,
			expectedErrMsg: "template rendering error",
		},
		{
			name:           "template rendering error in arguments",
			actionID:       "arg-template-error",
			renderTemplate: true,
			command: &environment.Command{
				Name: "test",
				Args: []string{"-a", "{{.invalid}}"},
			},
			parameters: params,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				outMaker *outputMocks.MockMaker,
				collector *outputMocks.MockCollector,
			) {
				outMaker.On("MakeCollector", "action arg-template-error").Return(collector).Once()
				svc.On("RenderTemplate", "test", params).Return("test", nil).Once()
				svc.On("RenderTemplate", "-a", params).Return("-a", nil).Once()
				svc.On("RenderTemplate", "{{.invalid}}", params).Return("", errors.New("arg template error")).Once()
			},
			want:           false,
			expectError:    true,
			expectedErrMsg: "arg template error",
		},
		{
			name:           "execution error with template rendering disabled",
			actionID:       "failed-action",
			renderTemplate: false,
			command: &environment.Command{
				Name: "test",
				Args: []string{"-a"},
			},
			parameters: params,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				outMaker *outputMocks.MockMaker,
				collector *outputMocks.MockCollector,
			) {
				outMaker.On("MakeCollector", "action failed-action").Return(collector).Once()

				// No template rendering since it's disabled

				expectedCmd := &environment.Command{
					Name: "test",
					Args: []string{"-a"},
				}
				svc.On("ExecCommand", ctx, expectedCmd, collector).Return(errors.New("execution failed")).Once()
			},
			want:           false,
			expectError:    true,
			expectedErrMsg: "execution failed",
		},
		{
			name:           "store error with template rendering",
			actionID:       "store-failed",
			renderTemplate: true,
			command: &environment.Command{
				Name: "test",
				Args: []string{"-a"},
			},
			parameters: params,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				outMaker *outputMocks.MockMaker,
				collector *outputMocks.MockCollector,
			) {
				outMaker.On("MakeCollector", "action store-failed").Return(collector).Once()

				svc.On("RenderTemplate", "test", params).Return("test", nil).Once()
				svc.On("RenderTemplate", "-a", params).Return("-a", nil).Once()

				expectedCmd := &environment.Command{
					Name: "test",
					Args: []string{"-a"},
				}
				svc.On("ExecCommand", ctx, expectedCmd, collector).Return(nil).Once()
				rd.On("Store", "command/store-failed", collector).Return(errors.New("store failed")).Once()
			},
			want:           false,
			expectError:    true,
			expectedErrMsg: "store failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			runDataMock := runtimeMocks.NewMockData(t)
			svcMock := servicesMocks.NewMockService(t)
			outMakerMock := outputMocks.NewMockMaker(t)
			collectorMock := outputMocks.NewMockCollector(t)

			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)

			var ctx context.Context
			if tt.contextSetup == nil {
				ctx = context.Background()
			} else {
				ctx = tt.contextSetup()
			}

			tt.setupMocks(t, ctx, runDataMock, fndMock, svcMock, outMakerMock, collectorMock)

			action := &Action{
				fnd:            fndMock,
				service:        svcMock,
				id:             tt.actionID,
				command:        tt.command,
				parameters:     tt.parameters,
				renderTemplate: tt.renderTemplate,
				outputMaker:    outMakerMock,
			}

			got, err := action.Execute(ctx, runDataMock)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAction_Timeout(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &Action{
		fnd:     fndMock,
		timeout: 2000 * time.Millisecond,
	}
	assert.Equal(t, 2000*time.Millisecond, a.Timeout())
}

func TestAction_When(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &Action{
		fnd:  fndMock,
		when: action.OnSuccess,
	}
	assert.Equal(t, action.OnSuccess, a.When())
}
