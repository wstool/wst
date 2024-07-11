package runtime

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSyncData_StoreAndLoad(t *testing.T) {
	data := CreateData()
	require.NotNil(t, data, "Data instance should not be nil")

	tests := []struct {
		name          string
		key           string
		value         interface{}
		loadKey       string
		expected      interface{}
		expectedFound bool
	}{
		{
			name:          "store and load string",
			key:           "key1",
			value:         "value1",
			loadKey:       "key1",
			expected:      "value1",
			expectedFound: true,
		},
		{
			name:          "store and load integer",
			key:           "key2",
			value:         12345,
			loadKey:       "key2",
			expected:      12345,
			expectedFound: true,
		},
		{
			name:          "non-existent key",
			key:           "key3",
			value:         "value3",
			loadKey:       "key_not_exist",
			expected:      nil,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := data.Store(tt.key, tt.value)
			assert.NoError(t, err, "Store should not return an error")

			result, found := data.Load(tt.loadKey)
			if tt.expectedFound {
				assert.True(t, found, "Expected to find a value for the key")
				assert.Equal(t, tt.expected, result, "Loaded value should match stored value")
			} else {
				assert.False(t, found, "Should not find a value for the key")
				assert.Nil(t, result, "Result should be nil when key does not exist")
			}
		})
	}
}
