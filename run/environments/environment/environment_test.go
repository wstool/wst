package environment

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	"testing"
)

func TestCommonMaker_MakeCommonEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		inputConfig    *types.CommonEnvironment
		expectedResult *CommonEnvironment
	}{
		{
			name: "Basic configuration",
			inputConfig: &types.CommonEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 3000,
					End:   4000,
				},
			},
			expectedResult: &CommonEnvironment{
				Ports: Ports{
					Start: 3000,
					Used:  3000,
					End:   4000,
				},
				Used: false,
			},
		},
		{
			name: "Empty configuration",
			inputConfig: &types.CommonEnvironment{
				Ports: types.EnvironmentPorts{},
			},
			expectedResult: &CommonEnvironment{
				Ports: Ports{
					Start: 0,
					Used:  0,
					End:   0,
				},
				Used: false,
			},
		},
		// Additional test cases can be added here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			maker := CreateCommonMaker(fndMock)
			result := maker.MakeCommonEnvironment(tt.inputConfig)

			// Assert that the foundation is properly set
			assert.Equal(t, fndMock, result.Fnd)
			assert.NotNil(t, result.OutputMaker)

			// Assert that Ports and Used are correctly set
			assert.Equal(t, tt.expectedResult.Ports.Start, result.Ports.Start)
			assert.Equal(t, tt.expectedResult.Ports.Used, result.Ports.Used)
			assert.Equal(t, tt.expectedResult.Ports.End, result.Ports.End)
			assert.Equal(t, tt.expectedResult.Used, result.Used)
		})
	}
}

func TestCommonEnvironment_MarkUsed(t *testing.T) {
	env := CommonEnvironment{}
	assert.False(t, env.Used, "Initially, Used should be false")

	env.MarkUsed()
	assert.True(t, env.Used, "MarkUsed should set Used to true")
}

func TestCommonEnvironment_IsUsed(t *testing.T) {
	env := CommonEnvironment{Used: true}
	assert.True(t, env.IsUsed(), "IsUsed should return true when Used is true")

	env.Used = false
	assert.False(t, env.IsUsed(), "IsUsed should return false when Used is false")
}

func TestCommonEnvironment_PortsStart(t *testing.T) {
	env := CommonEnvironment{
		Ports: Ports{
			Start: 3000,
		},
	}
	assert.Equal(t, int32(3000), env.PortsStart(), "PortsStart should return the correct start port")
}

func TestCommonEnvironment_PortsEnd(t *testing.T) {
	env := CommonEnvironment{
		Ports: Ports{
			End: 4000,
		},
	}
	assert.Equal(t, int32(4000), env.PortsEnd(), "PortsEnd should return the correct end port")
}

func TestCommonEnvironment_ReservePort(t *testing.T) {
	env := CommonEnvironment{
		Ports: Ports{
			Start: 3000,
			Used:  3000,
			End:   4000,
		},
	}
	assert.Equal(t, int32(3000), env.ReservePort(), "ReservePort should return the first available port")
	assert.Equal(t, int32(3001), env.Ports.Used, "ReservePort should increment the used port")

	// Test if it increments again
	assert.Equal(t, int32(3001), env.ReservePort(), "ReservePort should return the next available port")
	assert.Equal(t, int32(3002), env.Ports.Used, "ReservePort should increment the used port again")
}

func TestCommonEnvironment_ContainerRegistry(t *testing.T) {
	env := CommonEnvironment{}
	assert.Nil(t, env.ContainerRegistry(), "ContainerRegistry should return nil")
}

func TestContainerEnvironment_MakeContainerEnvironment(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t) // You'll need to create a mock for the Foundation
	maker := CreateCommonMaker(fndMock)

	tests := []struct {
		name     string
		config   *types.ContainerEnvironment
		expected *ContainerEnvironment
	}{
		{
			name: "initialize container environment with registry credentials",
			config: &types.ContainerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 3000,
					End:   3500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "user",
						Password: "pass",
					},
				},
			},
			expected: &ContainerEnvironment{
				CommonEnvironment: CommonEnvironment{
					Ports: Ports{
						Start: 3000,
						Used:  3000,
						End:   3500,
					},
				},
				Registry: ContainerRegistry{
					Auth: ContainerRegistryAuth{
						Username: "user",
						Password: "pass",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maker.MakeContainerEnvironment(tt.config)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected.Ports.Start, result.Ports.Start)
			assert.Equal(t, tt.expected.Ports.End, result.Ports.End)
			assert.Equal(t, tt.expected.Registry.Auth.Username, result.Registry.Auth.Username)
			assert.Equal(t, tt.expected.Registry.Auth.Password, result.Registry.Auth.Password)
		})
	}
}

func TestContainerEnvironment_ContainerRegistry(t *testing.T) {
	env := ContainerEnvironment{
		Registry: ContainerRegistry{
			Auth: ContainerRegistryAuth{
				Username: "testuser",
				Password: "testpass",
			},
		},
	}

	registry := env.ContainerRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, "testuser", registry.Auth.Username)
	assert.Equal(t, "testpass", registry.Auth.Password)
}
