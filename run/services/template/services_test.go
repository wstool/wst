package template

import (
	serviceMocks "github.com/bukka/wst/mocks/generated/run/services/template/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServices_Find(t *testing.T) {
	mockService := serviceMocks.NewMockTemplateService(t)
	mockService.TestData().Set("name", "testService")

	services := Services{
		"testService": mockService,
	}

	tests := []struct {
		name          string
		serviceName   string
		expectedError string
	}{
		{
			name:          "Service found",
			serviceName:   "testService",
			expectedError: "",
		},
		{
			name:          "Service not found",
			serviceName:   "unknownService",
			expectedError: "service not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := services.Find(tt.serviceName)

			if tt.expectedError != "" {
				assert.Nil(t, svc)
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NotNil(t, svc)
				assert.NoError(t, err)
				assert.Equal(t, mockService, svc)
			}
		})
	}
}
