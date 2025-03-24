package sequential

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

// TestCreateActionMaker remains the same
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
		config           *types.SequentialAction
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
			config: &types.SequentialAction{
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
			config: &types.SequentialAction{
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
			config: &types.SequentialAction{
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

	// Test implementation remains the same as it tests the maker functionality
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
		want        bool
		expectError bool
		errorMsg    string
		execActions int
	}{
		{
			name: "successful execution of all actions",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("Execute", ctx, rd).Return(true, nil)
				actions[0].On("When").Return(action.Always)
				actions[1].On("Execute", ctx, rd).Return(true, nil)
				actions[1].On("When").Return(action.Always)
				actions[2].On("Execute", ctx, rd).Return(true, nil)
				actions[2].On("When").Return(action.Always)
				fnd.On("DryRun").Return(false)
			},
			want:        true,
			execActions: 3,
		},
		{
			name: "continue on failure in dry-run mode",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("Execute", ctx, rd).Return(false, nil)
				actions[0].On("When").Return(action.Always)
				actions[1].On("Execute", ctx, rd).Return(false, nil)
				actions[1].On("When").Return(action.Always)
				actions[2].On("Execute", ctx, rd).Return(false, nil)
				actions[2].On("When").Return(action.Always)
				fnd.On("DryRun").Return(true)
			},
			want:        true,
			execActions: 3,
		},
		{
			name: "execute on failure condition",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("Execute", ctx, rd).Return(false, nil)
				actions[0].On("When").Return(action.Always)
				actions[1].On("Execute", ctx, rd).Return(true, nil)
				actions[1].On("When").Return(action.OnFailure)
				actions[2].On("When").Return(action.OnSuccess)
				fnd.On("DryRun").Return(false)
			},
			want:        false,
			execActions: 2,
		},
		{
			name: "execute on success condition",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("Execute", ctx, rd).Return(true, nil)
				actions[0].On("When").Return(action.Always)
				actions[1].On("Execute", ctx, rd).Return(true, nil)
				actions[1].On("When").Return(action.OnSuccess)
				actions[2].On("When").Return(action.OnFailure)
				fnd.On("DryRun").Return(false)
			},
			want:        true,
			execActions: 2,
		},
		{
			name: "error stops execution",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				actions []*actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				actions[0].On("When").Return(action.Always).Once()
				actions[0].On("Execute", ctx, rd).Return(false, errors.New("act fail")).Once()
				actions[1].On("When").Return(action.OnFailure).Once()
				actions[1].On("Execute", ctx, rd).Return(true, nil).Once()
				actions[2].On("When").Return(action.OnSuccess).Once()
				// DryRun will be called once
				fnd.On("DryRun").Return(false).Once()
			},
			want:        false,
			expectError: true,
			errorMsg:    "Sequential action failed with error: act fail",
			execActions: 2,
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

			runDataMock := runtimeMocks.NewMockData(t)
			runMakerMock := runtimeMocks.NewMockMaker(t)

			actionMocks := []*actionMocks.MockAction{
				actionMocks.NewMockAction(t),
				actionMocks.NewMockAction(t),
				actionMocks.NewMockAction(t),
			}
			cancelCalled := false

			// Setup context and timeout for each action that should be executed
			for i := 0; i < tt.execActions; i++ {
				cancel := context.CancelFunc(func() { cancelCalled = true })
				runMakerMock.On("MakeContextWithTimeout", baseCtx, timeout).Return(actCtx, cancel).Once()
				actionMocks[i].On("Timeout").Return(timeout)
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
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.True(t, cancelCalled)
			}

			// Verify that all mocks' expectations were met
			for _, mock := range actionMocks {
				mock.AssertExpectations(t)
			}
			runMakerMock.AssertExpectations(t)
			fndMock.AssertExpectations(t)
		})
	}
}
