package templates

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTemplates_Inherit(t *testing.T) {
	tests := []struct {
		name              string
		childTemplates    Templates
		parentTemplates   Templates
		expectedTemplates Templates
	}{
		{
			name: "inherit new templates",
			childTemplates: Templates{
				"existing": &nativeTemplate{
					filePath: "/path/existing",
				},
			},
			parentTemplates: Templates{
				"new": &nativeTemplate{
					filePath: "/path/new",
				},
			},
			expectedTemplates: Templates{
				"existing": &nativeTemplate{
					filePath: "/path/existing",
				},
				"new": &nativeTemplate{
					filePath: "/path/new",
				},
			},
		},
		{
			name: "do not override existing templates",
			childTemplates: Templates{
				"common": &nativeTemplate{
					filePath: "/path/common1",
				},
			},
			parentTemplates: Templates{
				"common": &nativeTemplate{
					filePath: "/path/common2",
				},
			},
			expectedTemplates: Templates{
				"common": &nativeTemplate{
					filePath: "/path/common1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.childTemplates.Inherit(tt.parentTemplates)
			assert.Equal(t, tt.expectedTemplates, tt.childTemplates, "Templates should be correctly inherited")
		})
	}
}

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name              string
		serverTemplates   map[string]types.ServerTemplate
		expectedTemplates Templates
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful template creation",
			serverTemplates: map[string]types.ServerTemplate{
				"template1": {File: "/path/template1"},
			},
			expectedTemplates: Templates{
				"template1": &nativeTemplate{
					filePath: "/path/template1",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			maker := CreateMaker(fndMock)

			templates, err := maker.Make(tt.serverTemplates)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTemplates, templates)
			}
		})
	}
}

func Test_nativeTemplate_FilePath(t *testing.T) {
	template := &nativeTemplate{
		filePath: "/path/template1",
	}
	assert.Equal(t, "/path/template1", template.FilePath())
}
