package expect

import (
	"github.com/bukka/wst/app"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	expectationsMocks "github.com/bukka/wst/mocks/generated/run/expectations"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	servicesMocks "github.com/bukka/wst/mocks/generated/run/services"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/parameters"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateExpectationActionMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	expectationsMakerMock := expectationsMocks.NewMockMaker(t)
	parametersMakerMock := parametersMocks.NewMockMaker(t)
	tests := []struct {
		name              string
		fnd               app.Foundation
		expectationsMaker expectations.Maker
		parametersMaker   parameters.Maker
	}{
		{
			name:              "create maker",
			fnd:               fndMock,
			expectationsMaker: expectationsMakerMock,
			parametersMaker:   parametersMakerMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateExpectationActionMaker(tt.fnd, tt.expectationsMaker, tt.parametersMaker)
			assert.Equal(t, tt.fnd, got.fnd)
			assert.Equal(t, tt.expectationsMaker, got.expectationsMaker)
			assert.Equal(t, tt.parametersMaker, got.parametersMaker)
		})
	}
}

func TestExpectationActionMaker_MakeCommonExpectation(t *testing.T) {
	serviceName := "sname"
	tests := []struct {
		name              string
		defaultTimeout    int
		timeout           int
		when              string
		setupMocks        func(*testing.T, *servicesMocks.MockServiceLocator) *servicesMocks.MockService
		getExpectedAction func(fndMock *appMocks.MockFoundation, svc *servicesMocks.MockService) *CommonExpectation
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name:           "successful common action creation with default timeout used",
			defaultTimeout: 5000,
			timeout:        0,
			when:           "on_success",
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) *servicesMocks.MockService {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", serviceName).Return(svc, nil)
				return svc
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svc *servicesMocks.MockService) *CommonExpectation {
				return &CommonExpectation{
					fnd:     fndMock,
					service: svc,
					timeout: 5000 * 1e6,
					when:    action.OnSuccess,
				}
			},
		},
		{
			name:           "successful common action creation with timeout used",
			defaultTimeout: 5000,
			timeout:        3000,
			when:           "on_success",
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) *servicesMocks.MockService {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", serviceName).Return(svc, nil)
				return svc
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svc *servicesMocks.MockService) *CommonExpectation {
				return &CommonExpectation{
					fnd:     fndMock,
					service: svc,
					timeout: 3000 * 1e6,
					when:    action.OnSuccess,
				}
			},
		},
		{
			name:           "failed common action because no service found",
			defaultTimeout: 5000,
			timeout:        3000,
			when:           "on_success",
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) *servicesMocks.MockService {
				sl.On("Find", serviceName).Return(nil, errors.New("no service"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "no service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			slMock := servicesMocks.NewMockServiceLocator(t)
			parametersMakerMock := parametersMocks.NewMockMaker(t)
			m := &ExpectationActionMaker{
				fnd:             fndMock,
				parametersMaker: parametersMakerMock,
			}

			svc := tt.setupMocks(t, slMock)

			got, err := m.MakeCommonExpectation(slMock, serviceName, tt.timeout, tt.defaultTimeout, tt.when)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				expectedAction := tt.getExpectedAction(fndMock, svc)
				assert.Equal(t, expectedAction, got)
			}
		})
	}
}
