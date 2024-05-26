package not

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	"github.com/bukka/wst/mocks/external"
	actionMocks "github.com/bukka/wst/mocks/run/actions/action"
	runtimeMocks "github.com/bukka/wst/mocks/run/instances/runtime"
	servicesMocks "github.com/bukka/wst/mocks/run/services"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
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
		config           *types.NotAction
		defaultTimeout   int
		passedTimeout    int
		expectedTimeout  time.Duration
		actionMakerErr   error
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful action creation with config timeout",
			config: &types.NotAction{
				Action: &types.StartAction{
					Service:  "name",
					Services: nil,
					Timeout:  4000,
				},
				Timeout: 3000,
			},
			defaultTimeout:  5000,
			passedTimeout:   3000,
			expectedTimeout: time.Duration(3000 * 1e6),
		},
		{
			name: "successful action creation with default timeout",
			config: &types.NotAction{
				Action: &types.StartAction{
					Service:  "name",
					Services: nil,
					Timeout:  4000,
				},
				Timeout: 0,
			},
			defaultTimeout:  5000,
			passedTimeout:   5000,
			expectedTimeout: time.Duration(5000 * 1e6),
		},
		{
			name: "failed action creation because of action maker failure",
			config: &types.NotAction{
				Action: &types.StartAction{
					Service:  "name",
					Services: nil,
					Timeout:  4000,
				},
				Timeout: 0,
			},
			defaultTimeout:   5000,
			passedTimeout:    5000,
			expectedTimeout:  time.Duration(5000 * 1e6),
			actionMakerErr:   errors.New("make failed"),
			expectError:      true,
			expectedErrorMsg: "make failed",
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
			actionMock := actionMocks.NewMockAction(t)
			if tt.actionMakerErr != nil {
				amMock.On("MakeAction", tt.config.Action, slMock, tt.passedTimeout).Return(
					nil,
					tt.actionMakerErr,
				)
			} else {
				amMock.On("MakeAction", tt.config.Action, slMock, tt.passedTimeout).Return(
					actionMock,
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
				assert.Equal(actionMock, act.action)
				assert.Equal(tt.expectedTimeout, act.Timeout())
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
			*actionMocks.MockAction,
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
				action *actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				action.On("Execute", ctx, rd).Return(true, nil)
				fnd.On("DryRun").Return(false)
			},
			want: false,
		},
		{
			name: "successful execution of false action result",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				action *actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				action.On("Execute", ctx, rd).Return(false, nil)
				fnd.On("DryRun").Return(false)
			},
			want: true,
		},
		{
			name: "successful execution of true action result with dry run",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				action *actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				action.On("Execute", ctx, rd).Return(true, nil)
				fnd.On("DryRun").Return(true)
			},
			want: true,
		},
		{
			name: "failed execution",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				action *actionMocks.MockAction,
				rd *runtimeMocks.MockData,
				ctx context.Context,
			) {
				action.On("Execute", ctx, rd).Return(false, errors.New("fail"))
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
			actionMock := actionMocks.NewMockAction(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)

			tt.setupMocks(t, fndMock, actionMock, runDataMock, ctx)

			a := &Action{
				fnd:    fndMock,
				action: actionMock,
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
