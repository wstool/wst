package run

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	confMocks "github.com/wstool/wst/mocks/generated/conf"
	specMocks "github.com/wstool/wst/mocks/generated/run/spec"
	"testing"
)

func TestCreateRunner(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	fndMock.TestData().Set("id", "fnd")
	r := CreateRunner(fndMock)
	require.NotNil(t, r)
	assert.Equal(t, fndMock, r.fnd)
	assert.NotNil(t, r.configMaker)
	assert.NotNil(t, r.specMaker)
}

func TestRunner_Execute(t *testing.T) {
	tests := []struct {
		name           string
		options        *Options
		setupMocks     func(fm *appMocks.MockFoundation, cm *confMocks.MockMaker, sm *specMocks.MockMaker)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful execution",
			options: &Options{
				ConfigPaths: []string{"config1.yaml", "config2.yaml"},
				IncludeAll:  false,
				Overwrites:  map[string]string{"key": "value"},
				Instances:   []string{"instance1", "instance2"},
			},
			setupMocks: func(fm *appMocks.MockFoundation, cm *confMocks.MockMaker, sm *specMocks.MockMaker) {
				config := &types.Config{Spec: types.Spec{
					Workspace: "/workspace",
				}}
				specification := specMocks.NewMockSpec(t)

				cm.On(
					"Make",
					[]string{"config1.yaml", "config2.yaml"},
					map[string]string{"key": "value"},
				).Return(config, nil)

				var filteredInstances []string = nil
				sm.On("Make", &config.Spec, filteredInstances).Return(specification, nil)

				specification.On("Run", []string{"instance1", "instance2"}).Return(nil)
			},
			expectError: false,
		},
		{
			name: "successful pre filtered execution",
			options: &Options{
				ConfigPaths: []string{"config1.yaml", "config2.yaml"},
				IncludeAll:  false,
				Overwrites:  map[string]string{"key": "value"},
				PreFilter:   true,
				Instances:   []string{"instance1", "instance2"},
			},
			setupMocks: func(fm *appMocks.MockFoundation, cm *confMocks.MockMaker, sm *specMocks.MockMaker) {
				config := &types.Config{Spec: types.Spec{
					Workspace: "/workspace",
				}}
				specification := specMocks.NewMockSpec(t)

				cm.On(
					"Make",
					[]string{"config1.yaml", "config2.yaml"},
					map[string]string{"key": "value"},
				).Return(config, nil)

				sm.On("Make", &config.Spec, []string{"instance1", "instance2"}).Return(specification, nil)

				specification.On("Run", []string{"instance1", "instance2"}).Return(nil)
			},
			expectError: false,
		},
		{
			name: "include all config paths",
			options: &Options{
				ConfigPaths: []string{},
				IncludeAll:  true,
				Overwrites:  map[string]string{"key": "value"},
				Instances:   []string{"instance1", "instance2"},
			},
			setupMocks: func(fm *appMocks.MockFoundation, cm *confMocks.MockMaker, sm *specMocks.MockMaker) {
				config := &types.Config{Spec: types.Spec{
					Workspace: "/workspace",
				}}
				specification := specMocks.NewMockSpec(t)

				fm.On("UserHomeDir").Return("/home/user", nil)

				memMapFs := afero.NewMemMapFs()
				_ = afero.WriteFile(memMapFs, "/home/user/.wst/wst.yaml", []byte("wst: {}"), 0644)
				_ = afero.WriteFile(memMapFs, "/home/user/.config/wst/wst.yaml", []byte("wst: {}"), 0644)
				fm.On("Fs").Return(memMapFs)

				cm.On(
					"Make",
					[]string{
						"/home/user/.wst/wst.yaml",
						"/home/user/.config/wst/wst.yaml",
					},
					map[string]string{"key": "value"},
				).Return(config, nil)

				var filteredInstances []string = nil
				sm.On("Make", &config.Spec, filteredInstances).Return(specification, nil)

				specification.On("Run", []string{"instance1", "instance2"}).Return(nil)
			},
			expectError: false,
		},
		{
			name: "error creating config",
			options: &Options{
				ConfigPaths: []string{"config1.yaml", "config2.yaml"},
				IncludeAll:  false,
				Overwrites:  map[string]string{"key": "value"},
				Instances:   []string{"instance1", "instance2"},
			},
			setupMocks: func(fm *appMocks.MockFoundation, cm *confMocks.MockMaker, sm *specMocks.MockMaker) {
				cm.On(
					"Make",
					[]string{"config1.yaml", "config2.yaml"},
					map[string]string{"key": "value"},
				).Return(nil, errors.New("config error"))
			},
			expectError:    true,
			expectedErrMsg: "config error",
		},
		{
			name: "error creating specification",
			options: &Options{
				ConfigPaths: []string{"config1.yaml", "config2.yaml"},
				IncludeAll:  false,
				Overwrites:  map[string]string{"key": "value"},
				Instances:   []string{"instance1", "instance2"},
			},
			setupMocks: func(fm *appMocks.MockFoundation, cm *confMocks.MockMaker, sm *specMocks.MockMaker) {
				config := &types.Config{Spec: types.Spec{
					Workspace: "/workspace",
				}}

				cm.On(
					"Make",
					[]string{"config1.yaml", "config2.yaml"},
					map[string]string{"key": "value"},
				).Return(config, nil)

				var filteredInstances []string = nil
				sm.On("Make", &config.Spec, filteredInstances).Return(nil, errors.New("spec error"))
			},
			expectError:    true,
			expectedErrMsg: "spec error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)
			confMakerMock := confMocks.NewMockMaker(t)
			specMakerMock := specMocks.NewMockMaker(t)

			runner := &Runner{
				fnd:         fndMock,
				configMaker: confMakerMock,
				specMaker:   specMakerMock,
			}

			tt.setupMocks(fndMock, confMakerMock, specMakerMock)

			err := runner.Execute(tt.options)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			confMakerMock.AssertExpectations(t)
			specMakerMock.AssertExpectations(t)
			fndMock.AssertExpectations(t)
		})
	}
}
