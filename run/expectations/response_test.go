package expectations

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	"testing"
)

func Test_nativeMaker_MakeResponseExpectation(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.ResponseExpectation
		expectError bool
		expected    *ResponseExpectation
		errorMsg    string
	}{
		{
			name: "valid exact match",
			config: &types.ResponseExpectation{
				Request: "/api/data",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body: types.ResponseBody{
					Match:          "exact",
					Content:        "Expected content",
					RenderTemplate: false,
				},
				Status: 200,
			},
			expectError: false,
			expected: &ResponseExpectation{
				Request:            "/api/data",
				Headers:            map[string]string{"Content-Type": "application/json"},
				BodyContent:        "Expected content",
				BodyMatch:          MatchTypeExact,
				BodyRenderTemplate: false,
				StatusCode:         200,
			},
		},
		{
			name: "valid regexp match",
			config: &types.ResponseExpectation{
				Request: "/api/data",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body: types.ResponseBody{
					Match:          "regexp",
					Content:        "^Expected.*content$",
					RenderTemplate: true,
				},
			},
			expectError: false,
			expected: &ResponseExpectation{
				Request:            "/api/data",
				Headers:            map[string]string{"Content-Type": "application/json"},
				BodyContent:        "^Expected.*content$",
				BodyMatch:          MatchTypeRegexp,
				BodyRenderTemplate: true,
			},
		},
		{
			name: "valid none match",
			config: &types.ResponseExpectation{
				Request: "/api/data",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body: types.ResponseBody{
					Match:          "",
					Content:        "",
					RenderTemplate: false,
				},
			},
			expectError: false,
			expected: &ResponseExpectation{
				Request:            "/api/data",
				Headers:            map[string]string{"Content-Type": "application/json"},
				BodyContent:        "",
				BodyMatch:          MatchTypeNone,
				BodyRenderTemplate: false,
			},
		},
		{
			name: "invalid match type",
			config: &types.ResponseExpectation{
				Request: "/api/error",
				Body: types.ResponseBody{
					Match:   "invalid",
					Content: "Should fail",
				},
			},
			expectError: true,
			errorMsg:    "invalid MatchType: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maker := &nativeMaker{}
			result, err := maker.MakeResponseExpectation(tt.config)
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
