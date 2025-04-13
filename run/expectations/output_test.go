package expectations

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	"testing"
)

func Test_nativeMaker_MakeOutputExpectation(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.OutputExpectation
		expectError bool
		expected    *OutputExpectation
		errorMsg    string
	}{
		{
			name: "valid configuration with exact match",
			config: &types.OutputExpectation{
				Order:    "fixed",
				Match:    "exact",
				Type:     "stdout",
				Messages: []string{"Hello, world!"},
			},
			expectError: false,
			expected: &OutputExpectation{
				OrderType:      OrderTypeFixed,
				MatchType:      MatchTypeExact,
				OutputType:     OutputTypeStdout,
				Messages:       []string{"Hello, world!"},
				RenderTemplate: false,
			},
		},
		{
			name: "valid configuration with prefix match and command",
			config: &types.OutputExpectation{
				Command:  "last",
				Order:    "fixed",
				Match:    "prefix",
				Type:     "stdout",
				Messages: []string{"Hello"},
			},
			expectError: false,
			expected: &OutputExpectation{
				Command:        "last",
				OrderType:      OrderTypeFixed,
				MatchType:      MatchTypePrefix,
				OutputType:     OutputTypeStdout,
				Messages:       []string{"Hello"},
				RenderTemplate: false,
			},
		},
		{
			name: "valid configuration with suffix match",
			config: &types.OutputExpectation{
				Order:    "fixed",
				Match:    "suffix",
				Type:     "stdout",
				Messages: []string{"world!"},
			},
			expectError: false,
			expected: &OutputExpectation{
				OrderType:      OrderTypeFixed,
				MatchType:      MatchTypeSuffix,
				OutputType:     OutputTypeStdout,
				Messages:       []string{"world!"},
				RenderTemplate: false,
			},
		},
		{
			name: "valid configuration with infix match",
			config: &types.OutputExpectation{
				Order:    "fixed",
				Match:    "infix",
				Type:     "stdout",
				Messages: []string{"llo, wor"},
			},
			expectError: false,
			expected: &OutputExpectation{
				OrderType:      OrderTypeFixed,
				MatchType:      MatchTypeInfix,
				OutputType:     OutputTypeStdout,
				Messages:       []string{"llo, wor"},
				RenderTemplate: false,
			},
		},
		{
			name: "invalid order type",
			config: &types.OutputExpectation{
				Order: "unknown",
				Match: "exact",
				Type:  "stdout",
			},
			expectError: true,
			errorMsg:    "invalid order type: unknown",
		},
		{
			name: "invalid match type",
			config: &types.OutputExpectation{
				Order: "fixed",
				Match: "unknown",
				Type:  "stdout",
			},
			expectError: true,
			errorMsg:    "invalid match type: unknown",
		},
		{
			name: "invalid output type",
			config: &types.OutputExpectation{
				Order: "fixed",
				Match: "exact",
				Type:  "unknown",
			},
			expectError: true,
			errorMsg:    "invalid output type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maker := &nativeMaker{}
			result, err := maker.MakeOutputExpectation(tt.config)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
