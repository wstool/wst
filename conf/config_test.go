package conf

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/loader"
	"github.com/wstool/wst/conf/parser"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	loader2 "github.com/wstool/wst/mocks/generated/conf/loader"
	mergerMocks "github.com/wstool/wst/mocks/generated/conf/merger"
	overwritesMocks "github.com/wstool/wst/mocks/generated/conf/overwrites"
	parserMocks "github.com/wstool/wst/mocks/generated/conf/parser"
	"testing"
)

func TestCreateConfigMaker(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	tests := []struct {
		name   string
		fnd    app.Foundation
		parser parser.Parser
	}{
		{
			name: "create merger",
			fnd:  fndMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := CreateConfigMaker(tt.fnd)
			cm, ok := m.(*ConfigMaker)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, cm.fnd)
			assert.NotNil(t, cm.merger)
			assert.NotNil(t, cm.parser)
			assert.NotNil(t, cm.loader)
			assert.NotNil(t, cm.overwriter)
		})
	}
}

func TestConfigMaker_Make(t *testing.T) {
	tests := []struct {
		name                    string
		configPaths             []string
		overwrites              map[string]string
		loadedConfigsMocksSetup func() []loader.LoadedConfig // Function to set up and return mocks
		parseConfigResults      []error
		mergeConfigResultConfig *types.Config
		mergeConfigResultError  error
		overwriterResult        error
		wantErr                 bool
	}{
		// Example test case
		{
			name:        "successful config creation with overwrites",
			configPaths: []string{"path/to/config1.yaml", "path/to/config2.yaml"},
			overwrites:  map[string]string{"path.to.value": "new value"},
			loadedConfigsMocksSetup: func() []loader.LoadedConfig {
				config1Mock := loader2.NewMockLoadedConfig(t)
				config1Mock.On("Path").Return("path/to/config1.yaml")
				config1Mock.On("Data").Return(map[string]interface{}{"config1Key": "config1Value"})

				config2Mock := loader2.NewMockLoadedConfig(t)
				config2Mock.On("Path").Return("path/to/config2.yaml")
				config2Mock.On("Data").Return(map[string]interface{}{"config2Key": "config2Value"})

				return []loader.LoadedConfig{config1Mock, config2Mock}
			},
			parseConfigResults:      []error{nil, nil},
			mergeConfigResultConfig: &types.Config{ /* populated config */ },
			mergeConfigResultError:  nil,
			overwriterResult:        nil,
			wantErr:                 false,
		},
		// Test case when loading configs returns an error
		{
			name:        "error loading configs",
			configPaths: []string{"path/to/nonexistent.yaml"},
			wantErr:     true,
			loadedConfigsMocksSetup: func() []loader.LoadedConfig {
				return nil // Simulating failure in loading configs
			},
			parseConfigResults:      nil, // Not reached due to loading error
			mergeConfigResultConfig: nil,
			mergeConfigResultError:  nil,
			overwriterResult:        nil,
		},
		// Test case when parsing a config returns an error
		{
			name:        "error parsing config",
			configPaths: []string{"path/to/config1.yaml"},
			wantErr:     true,
			loadedConfigsMocksSetup: func() []loader.LoadedConfig {
				config1Mock := loader2.NewMockLoadedConfig(t)
				config1Mock.On("Path").Return("path/to/config1.yaml")
				config1Mock.On("Data").Return(map[string]interface{}{"config1Key": "config1Value"})
				return []loader.LoadedConfig{config1Mock}
			},
			parseConfigResults:      []error{errors.New("parsing error")}, // Simulating failure in parsing
			mergeConfigResultConfig: nil,
			mergeConfigResultError:  nil,
			overwriterResult:        nil,
		},
		// Test case when merging configs returns an error
		{
			name:        "error merging configs",
			configPaths: []string{"path/to/config1.yaml", "path/to/config2.yaml"},
			wantErr:     true,
			loadedConfigsMocksSetup: func() []loader.LoadedConfig {
				config1Mock := loader2.NewMockLoadedConfig(t)
				config1Mock.On("Path").Return("path/to/config1.yaml")
				config1Mock.On("Data").Return(map[string]interface{}{"config1Key": "config1Value"})

				config2Mock := loader2.NewMockLoadedConfig(t)
				config2Mock.On("Path").Return("path/to/config2.yaml")
				config2Mock.On("Data").Return(map[string]interface{}{"config2Key": "config2Value"})

				return []loader.LoadedConfig{config1Mock, config2Mock}
			},
			parseConfigResults:      []error{nil, nil},
			mergeConfigResultConfig: nil,
			mergeConfigResultError:  errors.New("merging error"), // Simulating failure in merging
			overwriterResult:        nil,
		},
		// Test case when applying overwrites returns an error
		{
			name:        "error applying overwrites",
			configPaths: []string{"path/to/config1.yaml"},
			overwrites:  map[string]string{"invalid.path": "value"},
			wantErr:     true,
			loadedConfigsMocksSetup: func() []loader.LoadedConfig {
				config1Mock := loader2.NewMockLoadedConfig(t)
				config1Mock.On("Path").Return("path/to/config1.yaml")
				config1Mock.On("Data").Return(map[string]interface{}{"config1Key": "config1Value"})

				return []loader.LoadedConfig{config1Mock}
			},
			parseConfigResults:      []error{nil},
			mergeConfigResultConfig: &types.Config{},
			mergeConfigResultError:  nil,
			overwriterResult:        errors.New("overwriting error"), // Simulating failure in applying overwrites
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			loaderMock := loader2.NewMockLoader(t)
			parserMock := parserMocks.NewMockParser(t)
			mergerMock := mergerMocks.NewMockMerger(t)
			overwriterMock := overwritesMocks.NewMockOverwriter(t)

			cm := &ConfigMaker{
				fnd:        fndMock,
				loader:     loaderMock,
				parser:     parserMock,
				merger:     mergerMock,
				overwriter: overwriterMock,
			}

			loadedConfigs := tt.loadedConfigsMocksSetup()

			// Set up loader mock expectations
			if loadedConfigs == nil {
				loaderMock.On("LoadConfigs", mock.Anything).Return(nil, errors.New("no configs"))
			} else {
				loaderMock.On("LoadConfigs", mock.Anything).Return(loadedConfigs, nil)

				// Set up parser mock expectations for each loaded config
				isParserError := false
				for i, loadedConfig := range loadedConfigs {
					parserResult := tt.parseConfigResults[i]
					if parserResult != nil {
						isParserError = true
					}
					parserMock.On("ParseConfig", loadedConfig.Data(), mock.AnythingOfType("*types.Config"), loadedConfig.Path()).Return(parserResult)
				}

				if !isParserError {
					// Set up merger mock expectations
					mergerMock.On("MergeConfigs", mock.AnythingOfType("[]*types.Config")).Return(tt.mergeConfigResultConfig, tt.mergeConfigResultError)

					// Set up overwriter mock expectations, if there are overwrites
					if len(tt.overwrites) > 0 {
						overwriterMock.On("Overwrite", mock.AnythingOfType("*types.Config"), tt.overwrites).Return(tt.overwriterResult)
					}
				}
			}

			// Call the Make function
			resultConfig, err := cm.Make(tt.configPaths, tt.overwrites)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mergeConfigResultConfig, resultConfig)
			}

			loaderMock.AssertExpectations(t)
			parserMock.AssertExpectations(t)
			mergerMock.AssertExpectations(t)
			overwriterMock.AssertExpectations(t)
		})
	}
}
