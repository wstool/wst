package merger

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_nativeMerger_MergeConfigs(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}

	tests := []struct {
		name           string
		configs        []*types.Config
		overwrites     map[string]string
		expectedConfig *types.Config
		wantErr        bool
		errMsg         string
	}{
		{
			name: "Merge basic fields and complex structures",
			configs: []*types.Config{
				{
					Version:     "1.0",
					Name:        "Config 1",
					Description: "Description 1",
					Spec: types.Spec{
						Environments: map[string]types.Environment{
							"common": &types.CommonEnvironment{
								Ports: types.EnvironmentPorts{Start: 8000, End: 8080},
							},
						},
						Servers: []types.Server{
							{Name: "Server 1", User: "user1"},
						},
					},
				},
				{
					Version:     "1.0",
					Name:        "Config 2",
					Description: "Description 2",
					Spec: types.Spec{
						Environments: map[string]types.Environment{
							"docker": &types.DockerEnvironment{
								NamePrefix: "prefix-",
							},
						},
						Servers: []types.Server{
							{Name: "Server 2", User: "user2"},
						},
					},
				},
			},
			expectedConfig: &types.Config{
				Version:     "1.0",
				Name:        "Config 2",
				Description: "Description 2",
				Spec: types.Spec{
					Environments: map[string]types.Environment{
						"common": &types.CommonEnvironment{
							Ports: types.EnvironmentPorts{Start: 8000, End: 8080},
						},
						"docker": &types.DockerEnvironment{
							NamePrefix: "prefix-",
						},
					},
					Servers: []types.Server{
						{Name: "Server 1", User: "user1"},
						{Name: "Server 2", User: "user2"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := &nativeMerger{
				fnd: fndMock,
			}
			config, err := merger.MergeConfigs(tt.configs, tt.overwrites)

			if tt.errMsg != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tt.expectedConfig, config)
				}
			}
		})
	}
}
