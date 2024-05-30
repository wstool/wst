package reload

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	"github.com/bukka/wst/mocks/external"
	runtimeMocks "github.com/bukka/wst/mocks/run/instances/runtime"
	servicesMocks "github.com/bukka/wst/mocks/run/services"
	"github.com/bukka/wst/run/services"
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
		name              string
		config            *types.ReloadAction
		defaultTimeout    int
		setupMocks        func(*testing.T, *servicesMocks.MockServiceLocator) services.Services
		getExpectedAction func(*appMocks.MockFoundation, services.Services) *Action
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful reload action creation with default timeout",
			config: &types.ReloadAction{
				Service:  "validService3",
				Services: []string{"validService1", "validService2"},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Services {
				svc1 := servicesMocks.NewMockService(t)
				svc2 := servicesMocks.NewMockService(t)
				svc3 := servicesMocks.NewMockService(t)
				svc1.On("Name").Return("s1")
				svc2.On("Name").Return("s2")
				svc3.On("Name").Return("s3")
				sl.On("Find", "validService1").Return(svc1, nil)
				sl.On("Find", "validService2").Return(svc2, nil)
				sl.On("Find", "validService3").Return(svc3, nil)
				return services.Services{
					"s1": svc1,
					"s2": svc2,
					"s3": svc3,
				}
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svcs services.Services) *Action {
				return &Action{
					fnd:      fndMock,
					services: svcs,
					timeout:  5000 * time.Millisecond,
				}
			},
		},
		{
			name: "successful reload action creation with set timeout",
			config: &types.ReloadAction{
				Service: "validService",
				Timeout: 3000,
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Services {
				svc := servicesMocks.NewMockService(t)
				svc.On("Name").Return("service")
				sl.On("Find", "validService").Return(svc, nil)
				return services.Services{
					"service": svc,
				}
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svcs services.Services) *Action {
				return &Action{
					fnd:      fndMock,
					services: svcs,
					timeout:  3000 * time.Millisecond,
				}
			},
		},
		{
			name:           "successful reload action creation without any service",
			config:         &types.ReloadAction{},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Services {
				svcs := services.Services{
					"s1": servicesMocks.NewMockService(t),
					"s2": servicesMocks.NewMockService(t),
				}
				sl.On("Services").Return(svcs)
				return svcs
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svcs services.Services) *Action {
				return &Action{
					fnd:      fndMock,
					services: svcs,
					timeout:  5000 * time.Millisecond,
				}
			},
		},

		{
			name: "failed reload action creation when service not found",
			config: &types.ReloadAction{
				Service: "invalidService",
				Timeout: 3000,
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Services {
				sl.On("Find", "invalidService").Return(nil, errors.New("nf"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "nf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			slMock := servicesMocks.NewMockServiceLocator(t)
			m := &ActionMaker{
				fnd: fndMock,
			}

			svcs := tt.setupMocks(t, slMock)

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
				expectedAction := tt.getExpectedAction(fndMock, svcs)
				assert.Equal(t, expectedAction, actualAction)
			}
		})
	}
}

func TestAction_Execute(t *testing.T) {
	tests := []struct {
		name             string
		setupMocks       func(*testing.T, context.Context) services.Services
		want             bool
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful execution of all services",
			setupMocks: func(t *testing.T, ctx context.Context) services.Services {
				svc1 := servicesMocks.NewMockService(t)
				svc2 := servicesMocks.NewMockService(t)
				svc3 := servicesMocks.NewMockService(t)
				svc1.On("Name").Return("s1")
				svc2.On("Name").Return("s2")
				svc3.On("Name").Return("s3")
				svc1.On("Reload", ctx).Return(nil)
				svc2.On("Reload", ctx).Return(nil)
				svc3.On("Reload", ctx).Return(nil)
				return services.Services{
					"s1": svc1,
					"s2": svc2,
					"s3": svc3,
				}
			},
			want: true,
		},
		{
			name: "failed execution of reload",
			setupMocks: func(t *testing.T, ctx context.Context) services.Services {
				svc1 := servicesMocks.NewMockService(t)
				svc2 := servicesMocks.NewMockService(t)
				svc3 := servicesMocks.NewMockService(t)
				svc1.On("Name").Maybe().Return("s1")
				svc2.On("Name").Return("s2")
				svc3.On("Name").Maybe().Return("s3")
				svc1.On("Reload", ctx).Maybe().Return(nil)
				svc2.On("Reload", ctx).Return(errors.New("e2 failure"))
				svc3.On("Reload", ctx).Maybe().Return(nil)
				return services.Services{
					"s1": svc1,
					"s2": svc2,
					"s3": svc3,
				}
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "e2 failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			runDataMock := runtimeMocks.NewMockData(t)
			ctx := context.Background()
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)

			svcs := tt.setupMocks(t, ctx)

			a := &Action{
				fnd:      fndMock,
				services: svcs,
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

func TestAction_Timeout(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &Action{
		fnd:     fndMock,
		timeout: 2000 * time.Millisecond,
	}
	assert.Equal(t, 2000*time.Millisecond, a.Timeout())
}
