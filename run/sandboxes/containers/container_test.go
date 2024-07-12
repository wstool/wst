package containers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainerConfig_Image(t *testing.T) {
	tests := []struct {
		name     string
		config   ContainerConfig
		expected string
	}{
		{
			name: "simple image name and tag",
			config: ContainerConfig{
				ImageName: "ubuntu",
				ImageTag:  "20.04",
			},
			expected: "ubuntu:20.04",
		},
		{
			name: "image name with registry and tag",
			config: ContainerConfig{
				ImageName: "docker.io/ubuntu",
				ImageTag:  "latest",
			},
			expected: "docker.io/ubuntu:latest",
		},
		{
			name: "image name without tag",
			config: ContainerConfig{
				ImageName: "nginx",
				ImageTag:  "",
			},
			expected: "nginx:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.Image(), "Image method should concatenate image name and tag correctly")
		})
	}
}
