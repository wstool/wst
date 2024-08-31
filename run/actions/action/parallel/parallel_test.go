package parallel

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	actionMocks "github.com/bukka/wst/mocks/generated/run/actions/action"
	runtimeMocks "github.com/bukka/wst/mocks/generated/run/instances/runtime"
	servicesMocks "github.com/bukka/wst/mocks/generated/run/services"
	"github.com/bukka/wst/run/actions/action"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
			m := &ActionMaker{
				fnd: fndMock,
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
			runDataMock := runtimeMocks.NewMockData(t)
			ctx := context.Background()
			actionMocks := []*actionMocks.MockAction{
				actionMocks.NewMockAction(t),
				actionMocks.NewMockAction(t),
				actionMocks.NewMockAction(t),
			}
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)

			tt.setupMocks(t, fndMock, actionMocks, runDataMock, ctx)

			actions := []action.Action{actionMocks[0], actionMocks[1], actionMocks[2]}

			a := &Action{
				fnd:     fndMock,
				actions: actions,
			}

			got, err := a.Execute(ctx, runDataMock)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
