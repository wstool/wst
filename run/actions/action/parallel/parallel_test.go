package parallel

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	actionMocks "github.com/wstool/wst/mocks/generated/run/actions/action"
	runtimeMocks "github.com/wstool/wst/mocks/generated/run/instances/runtime"
	servicesMocks "github.com/wstool/wst/mocks/generated/run/services"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/instances/runtime"
	"testing"
	"time"
)

func TestCreateActionMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	runtimeMock := runtimeMocks.NewMockMaker(t)
	tests := []struct {
		name        string
		fnd         app.Foundation
		runtimeMock runtime.Maker
	}{
		{
			name:        "create maker",
			fnd:         fndMock,
			runtimeMock: runtimeMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateActionMaker(tt.fnd, tt.runtimeMock)
			assert.Equal(t, tt.fnd, got.fnd)
		})
	}
}

func TestActionMaker_Make(t *testing.T) {
	tests := []struct {
		name             string
		config           *types.ParallelAction
		defaultTimeout   int
		passedTimeout    int
		expectedTimeout  time.Duration
		expectedWhen     action.When
		actionMakerErr   error
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful action creation with config timeout",
			config: &types.ParallelAction{
				Actions: []types.Action{
					&types.RequestAction{
						Service: "s1",
						Timeout: 4000,
					},
					&types.RequestAction{
						Service: "s2",
						Timeout: 4000,
					},
				},
				Timeout: 3000,
				When:    "on_success",
			},
			defaultTimeout:  5000,
			passedTimeout:   3000,
			expectedTimeout: time.Duration(3000 * 1e6),
			expectedWhen:    action.OnSuccess,
		},
		{
			name: "successful action creation with default timeout",
			config: &types.ParallelAction{
				Actions: []types.Action{
					&types.RequestAction{
						Service: "s1",
						Timeout: 4000,
					},
					&types.RequestAction{
						Service: "s2",
						Timeout: 4000,
					},
				},
				Timeout: 0,
				When:    "on_success",
			},
			defaultTimeout:  5000,
			passedTimeout:   5000,
			expectedTimeout: time.Duration(5000 * 1e6),
			expectedWhen:    action.OnSuccess,
		},
		{
			name: "failed action creation because of action maker failure",
			config: &types.ParallelAction{
				Actions: []types.Action{
					&types.RequestAction{
						Service: "s1",
						Timeout: 4000,
					},
					&types.RequestAction{
						Service: "s2",
						Timeout: 4000,
					},
				},
				Timeout: 0,
				When:    "on_success",
			},
			defaultTimeout:   5000,
			passedTimeout:    5000,
			expectedTimeout:  time.Duration(5000 * 1e6),
			actionMakerErr:   errors.New("make failed"),
			expectError:      true,
			expectedErrorMsg: "make failed",
			expectedWhen:     action.OnSuccess,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			runtimeMakerMock := runtimeMocks.NewMockMaker(t)
			m := &ActionMaker{
				fnd:          fndMock,
				runtimeMaker: runtimeMakerMock,
			}
			slMock := servicesMocks.NewMockServiceLocator(t)
			amMock := actionMocks.NewMockMaker(t)
			action1Mock := actionMocks.NewMockAction(t)
			action2Mock := actionMocks.NewMockAction(t)
			if tt.actionMakerErr != nil {
				amMock.On("MakeAction", tt.config.Actions[0], slMock, tt.passedTimeout).Return(
					nil,
					tt.actionMakerErr,
				)
			} else {
				amMock.On("MakeAction", tt.config.Actions[0], slMock, tt.passedTimeout).Return(
					action1Mock,
					nil,
				)
				amMock.On("MakeAction", tt.config.Actions[1], slMock, tt.passedTimeout).Return(
					action2Mock,
					nil,
				)
			}
			got, err := m.Make(tt.config, slMock, tt.defaultTimeout, amMock)
			assert := assert.New(t)
			if tt.expectError {
				assert.Error(err)
				assert.Nil(got)
				assert.Contains(err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(err)
				assert.NotNil(got)
				act, ok := got.(*Action)
				assert.True(ok)
				assert.Equal(runtimeMakerMock, act.runtimeMaker)
				assert.Equal([]action.Action{action1Mock, action2Mock}, act.actions)
				assert.Equal(tt.expectedTimeout, act.Timeout())
				assert.Equal(tt.expectedWhen, act.When())
			}
		})
	}
}

func TestAction_Execute(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(
			*testing.T,
			*appMocks.MockFoundation,
			[]*actionMocks.MockAction,
			*runtimeMocks.MockData,
			context.Context,
		)
		want             bool
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful execution of true action result",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("Execute", ctx, rd).Return(true, nil)
				actions[1].On("Execute", ctx, rd).Return(true, nil)
				actions[2].On("Execute", ctx, rd).Return(true, nil)
			},
			want: true,
		},
		{
			name: "successful execution of false action result of all actions",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("Execute", ctx, rd).Return(false, nil)
				actions[1].On("Execute", ctx, rd).Return(false, nil)
				actions[2].On("Execute", ctx, rd).Return(false, nil)
				fnd.On("DryRun").Return(false)
			},
			want: false,
		},
		{
			name: "successful execution of false action result of a single action",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("Execute", ctx, rd).Return(true, nil)
				actions[1].On("Execute", ctx, rd).Return(false, nil)
				actions[2].On("Execute", ctx, rd).Return(true, nil)
				fnd.On("DryRun").Return(false)
			},
			want: false,
		},
		{
			name: "successful execution of true action result with dry run",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("Execute", ctx, rd).Return(false, nil)
				actions[1].On("Execute", ctx, rd).Return(false, nil)
				actions[2].On("Execute", ctx, rd).Return(false, nil)
				fnd.On("DryRun").Return(true)
			},
			want: true,
		},
		{
			name: "failed execution",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {

				actions[0].On("Execute", ctx, rd).Return(false, errors.New("fail"))
				actions[1].On("Execute", ctx, rd).Return(false, errors.New("fail"))
				actions[2].On("Execute", ctx, rd).Return(false, errors.New("fail"))
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			baseCtx, baseCancel := context.WithTimeout(context.Background(), 5*time.Second)
			actCtx, actCancel := context.WithTimeout(baseCtx, 3*time.Second)
			defer actCancel()
			defer baseCancel()
			timeout := 3 * time.Second
			cancelCalled := false
			cancel := context.CancelFunc(func() { cancelCalled = true })
			runDataMock := runtimeMocks.NewMockData(t)
			runMakerMock := runtimeMocks.NewMockMaker(t)
			runMakerMock.On("MakeContextWithTimeout", baseCtx, timeout).Return(actCtx, cancel)
			actionMocks := []*actionMocks.MockAction{
				actionMocks.NewMockAction(t),
				actionMocks.NewMockAction(t),
				actionMocks.NewMockAction(t),
			}
			for _, actionMock := range actionMocks {
				actionMock.On("Timeout").Return(timeout).Maybe()
			}

			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)

			tt.setupMocks(t, fndMock, actionMocks, runDataMock, actCtx)

			actions := []action.Action{actionMocks[0], actionMocks[1], actionMocks[2]}

			a := &Action{
				fnd:          fndMock,
				runtimeMaker: runMakerMock,
				actions:      actions,
			}

			got, err := a.Execute(baseCtx, runDataMock)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.True(t, cancelCalled)
			}
		})
	}
}
