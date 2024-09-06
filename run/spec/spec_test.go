package spec

import (
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	instancesMocks "github.com/bukka/wst/mocks/generated/run/instances"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	serversMocks "github.com/bukka/wst/mocks/generated/run/servers"
	defaultsMocks "github.com/bukka/wst/mocks/generated/run/spec/defaults"
	"github.com/bukka/wst/run/instances"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/spec/defaults"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	fndMock.TestData().Set("id", "fnd")
	m := CreateMaker(fndMock)
	require.NotNil(t, m)
	nm, ok := m.(*nativeMaker)
	assert.True(t, ok)
	assert.Equal(t, fndMock, nm.fnd)
	assert.NotNil(t, nm.defaultsMaker)
	assert.NotNil(t, nm.instanceMaker)
	assert.NotNil(t, nm.serversMaker)
}

func Test_nativeMaker_Make(t *testing.T) {
	envsConfig := map[string]types.Environment{
		"local": types.LocalEnvironment{},
	}
	defaultsConfig := types.SpecDefaults{
		Service: types.SpecServiceDefaults{
			Sandbox: "docker",
			Server: types.SpecServiceServerDefaults{
				Tag: "latest",
			},
		},
		Timeouts: types.SpecTimeouts{
			Actions: 11000,
			Action:  7000,
		},
		Parameters: types.Parameters{
			"dk": "dv",
		},
	}
	dflts := &defaults.Defaults{
		Service: defaults.ServiceDefaults{
			Sandbox: "docker",
			Server: defaults.ServiceServerDefaults{
				Tag: "latest",
			},
		},
		Timeouts: defaults.TimeoutsDefaults{
			Actions: 11000,
			Action:  7000,
		},
		Parameters: parameters.Parameters{
			"dk": parameterMocks.NewMockParameter(t),
		},
	}
	tests := []struct {
		name       string
		config     *types.Spec
		setupMocks func(
			*types.Spec,
			*defaultsMocks.MockMaker,
			*serversMocks.MockMaker,
			*instancesMocks.MockInstanceMaker,
		) []instances.Instance
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful spec creation",
			config: &types.Spec{
				Instances:    []types.Instance{{Name: "i1"}, {Name: "i2"}},
				Environments: envsConfig,
				Defaults:     defaultsConfig,
				Workspace:    "/workspace",
			},
			setupMocks: func(
				cfg *types.Spec,
				dm *defaultsMocks.MockMaker,
				sm *serversMocks.MockMaker,
				im *instancesMocks.MockInstanceMaker,
			) []instances.Instance {
				srvs := servers.Servers{
					"php": {
						"base": serversMocks.NewMockServer(t),
					},
				}
				sm.On("Make", cfg).Return(srvs, nil)
				dm.On("Make", &cfg.Defaults).Return(dflts, nil)
				i1 := instancesMocks.NewMockInstance(t)
				i1.TestData().Set("id", "i1")
				i2 := instancesMocks.NewMockInstance(t)
				i2.TestData().Set("id", "i1")
				im.On("Make", types.Instance{Name: "i1"}, envsConfig, dflts, srvs, "/workspace").Return(i1, nil)
				im.On("Make", types.Instance{Name: "i2"}, envsConfig, dflts, srvs, "/workspace").Return(i2, nil)
				return []instances.Instance{i1, i2}
			},
			expectError: false,
		},
		{
			name: "failed spec creation on instance make fail",
			config: &types.Spec{
				Instances:    []types.Instance{{Name: "i1"}, {Name: "i2"}},
				Environments: envsConfig,
				Defaults:     defaultsConfig,
				Workspace:    "/workspace",
			},
			setupMocks: func(
				cfg *types.Spec,
				dm *defaultsMocks.MockMaker,
				sm *serversMocks.MockMaker,
				im *instancesMocks.MockInstanceMaker,
			) []instances.Instance {
				srvs := servers.Servers{
					"php": {
						"base": serversMocks.NewMockServer(t),
					},
				}
				sm.On("Make", cfg).Return(srvs, nil)
				dm.On("Make", &cfg.Defaults).Return(dflts, nil)
				i1 := instancesMocks.NewMockInstance(t)
				i1.TestData().Set("id", "i1")
				im.On("Make", types.Instance{Name: "i1"}, envsConfig, dflts, srvs, "/workspace").Return(
					nil,
					errors.New("instance fail"),
				)
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "instance fail",
		},
		{
			name: "failed spec creation on instance with empty name",
			config: &types.Spec{
				Instances:    []types.Instance{{Name: ""}, {Name: "i2"}},
				Environments: envsConfig,
				Defaults:     defaultsConfig,
				Workspace:    "/workspace",
			},
			setupMocks: func(
				cfg *types.Spec,
				dm *defaultsMocks.MockMaker,
				sm *serversMocks.MockMaker,
				im *instancesMocks.MockInstanceMaker,
			) []instances.Instance {
				srvs := servers.Servers{
					"php": {
						"base": serversMocks.NewMockServer(t),
					},
				}
				sm.On("Make", cfg).Return(srvs, nil)
				dm.On("Make", &cfg.Defaults).Return(dflts, nil)
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "instance 0 name is empty",
		},
		{
			name: "failed spec creation on defaults make fail",
			config: &types.Spec{
				Instances:    []types.Instance{{Name: "i1"}, {Name: "i2"}},
				Environments: envsConfig,
				Defaults:     defaultsConfig,
				Workspace:    "/workspace",
			},
			setupMocks: func(
				cfg *types.Spec,
				dm *defaultsMocks.MockMaker,
				sm *serversMocks.MockMaker,
				im *instancesMocks.MockInstanceMaker,
			) []instances.Instance {
				srvs := servers.Servers{
					"php": {
						"base": serversMocks.NewMockServer(t),
					},
				}
				sm.On("Make", cfg).Return(srvs, nil)
				dm.On("Make", &cfg.Defaults).Return(nil, errors.New("defaults fail"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "defaults fail",
		},
		{
			name: "failed spec creation on servers make fail",
			config: &types.Spec{
				Instances:    []types.Instance{{Name: "i1"}, {Name: "i2"}},
				Environments: envsConfig,
				Defaults:     defaultsConfig,
				Workspace:    "/workspace",
			},
			setupMocks: func(
				cfg *types.Spec,
				dm *defaultsMocks.MockMaker,
				sm *serversMocks.MockMaker,
				im *instancesMocks.MockInstanceMaker,
			) []instances.Instance {
				sm.On("Make", cfg).Return(nil, errors.New("servers fail"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "servers fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			serverMakerMock := serversMocks.NewMockMaker(t)
			instanceMakerMock := instancesMocks.NewMockInstanceMaker(t)
			defaultsMaker := defaultsMocks.NewMockMaker(t)
			maker := &nativeMaker{
				fnd:           fndMock,
				defaultsMaker: defaultsMaker,
				serversMaker:  serverMakerMock,
				instanceMaker: instanceMakerMock,
			}
			expectedInstances := tt.setupMocks(tt.config, defaultsMaker, serverMakerMock, instanceMakerMock)

			result, err := maker.Make(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
				specResult := result.(*nativeSpec)
				assert.Equal(t, fndMock, specResult.fnd)
				assert.Equal(t, tt.config.Workspace, specResult.workspace)
				assert.Equal(t, expectedInstances, specResult.instances)
			}
		})
	}
}

func Test_nativeSpec_Run(t *testing.T) {
	instance1 := instancesMocks.NewMockInstance(t)
	instance1.On("Name").Return("instance1").Maybe()
	instance1.On("Run").Return(nil).Maybe()

	instance2 := instancesMocks.NewMockInstance(t)
	instance2.On("Name").Return("instance2").Maybe()
	instance2.On("Run").Return(errors.New("failure in instance2")).Maybe()

	instance3 := instancesMocks.NewMockInstance(t)
	instance3.On("Name").Return("instance3").Maybe()
	instance3.On("Run").Return(nil).Maybe()

	tests := []struct {
		name               string
		instances          []instances.Instance
		filteredInstances  []string
		expectedRun        []string
		expectedSkip       []string
		expectError        bool
		expectedError      string
		expectedErrorCount int
	}{
		{
			name:              "Run all instances with empty filter",
			instances:         []instances.Instance{instance1, instance2, instance3},
			filteredInstances: nil,
			expectedRun:       []string{"instance1", "instance2", "instance3"},
			expectedSkip:      nil,
			expectError:       true,
			expectedError:     "failure in instance2",
		},
		{
			name:              "Run filtered instances only",
			instances:         []instances.Instance{instance1, instance2, instance3},
			filteredInstances: []string{"instance1", "instance3"},
			expectedRun:       []string{"instance1", "instance3"},
			expectedSkip:      []string{"instance2"},
			expectError:       false,
		},
		{
			name:              "Handle no instances to run",
			instances:         []instances.Instance{instance2},    // instance2 will fail
			filteredInstances: []string{"instance1", "instance3"}, // they are not in the list
			expectedRun:       nil,
			expectedSkip:      []string{"instance2"},
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()

			fndMock.On("Logger").Return(mockLogger.SugaredLogger)

			spec := &nativeSpec{
				fnd:       fndMock,
				workspace: "test_workspace",
				instances: tt.instances,
			}

			// Run the spec
			err := spec.Run(tt.filteredInstances)

			// Check for expected errors
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			for _, instance := range tt.instances {
				instance.(*instancesMocks.MockInstance).AssertExpectations(t)
			}
		})
	}
}
