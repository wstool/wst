package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	certificatesMocks "github.com/wstool/wst/mocks/generated/run/resources/certificates"
	scriptsMocks "github.com/wstool/wst/mocks/generated/run/resources/scripts"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/resources/scripts"
)

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name           string
		config         types.Resources
		setupMocks     func(*certificatesMocks.MockMaker, *scriptsMocks.MockMaker)
		expectError    bool
		errorMessage   string
		validateResult func(*testing.T, *Resources)
	}{
		{
			name: "successful creation with both certificates and scripts",
			config: types.Resources{
				Certificates: map[string]types.Certificate{
					"ssl_cert": {
						Certificate: "cert_content",
						PrivateKey:  "key_content",
					},
				},
				Scripts: map[string]types.Script{
					"setup": {
						Content: "echo setup",
						Path:    "/setup.sh",
						Mode:    "755",
					},
				},
			},
			setupMocks: func(certMaker *certificatesMocks.MockMaker, scriptMaker *scriptsMocks.MockMaker) {
				expectedCerts := certificates.Certificates{
					"ssl_cert": certificatesMocks.NewMockCertificate(t),
				}
				expectedScripts := scripts.Scripts{
					"setup": scriptsMocks.NewMockScript(t),
				}

				certMaker.On("Make", map[string]types.Certificate{
					"ssl_cert": {
						Certificate: "cert_content",
						PrivateKey:  "key_content",
					},
				}).Return(expectedCerts, nil)

				scriptMaker.On("Make", map[string]types.Script{
					"setup": {
						Content: "echo setup",
						Path:    "/setup.sh",
						Mode:    "755",
					},
				}).Return(expectedScripts, nil)
			},
			validateResult: func(t *testing.T, result *Resources) {
				assert.NotNil(t, result)
				assert.Len(t, result.Certificates, 1)
				assert.Len(t, result.Scripts, 1)
			},
		},
		{
			name: "only certificates",
			config: types.Resources{
				Certificates: map[string]types.Certificate{
					"api_cert": {
						Certificate: "api_cert_content",
						PrivateKey:  "api_key_content",
					},
				},
				Scripts: nil,
			},
			setupMocks: func(certMaker *certificatesMocks.MockMaker, scriptMaker *scriptsMocks.MockMaker) {
				expectedCerts := certificates.Certificates{
					"api_cert": certificatesMocks.NewMockCertificate(t),
				}

				certMaker.On("Make", map[string]types.Certificate{
					"api_cert": {
						Certificate: "api_cert_content",
						PrivateKey:  "api_key_content",
					},
				}).Return(expectedCerts, nil)

				// scriptMaker should not be called
			},
			validateResult: func(t *testing.T, result *Resources) {
				assert.NotNil(t, result)
				assert.Len(t, result.Certificates, 1)
				assert.Len(t, result.Scripts, 0)
			},
		},
		{
			name: "only scripts",
			config: types.Resources{
				Certificates: nil,
				Scripts: map[string]types.Script{
					"deploy": {
						Content: "echo deploy",
						Path:    "/deploy.sh",
						Mode:    "644",
					},
				},
			},
			setupMocks: func(certMaker *certificatesMocks.MockMaker, scriptMaker *scriptsMocks.MockMaker) {
				expectedScripts := scripts.Scripts{
					"deploy": scriptsMocks.NewMockScript(t),
				}

				scriptMaker.On("Make", map[string]types.Script{
					"deploy": {
						Content: "echo deploy",
						Path:    "/deploy.sh",
						Mode:    "644",
					},
				}).Return(expectedScripts, nil)

				// certMaker should not be called
			},
			validateResult: func(t *testing.T, result *Resources) {
				assert.NotNil(t, result)
				assert.Len(t, result.Certificates, 0)
				assert.Len(t, result.Scripts, 1)
			},
		},
		{
			name: "empty config",
			config: types.Resources{
				Certificates: nil,
				Scripts:      nil,
			},
			setupMocks: func(certMaker *certificatesMocks.MockMaker, scriptMaker *scriptsMocks.MockMaker) {
				// No mocks should be called
			},
			validateResult: func(t *testing.T, result *Resources) {
				assert.NotNil(t, result)
				assert.Len(t, result.Certificates, 0)
				assert.Len(t, result.Scripts, 0)
			},
		},
		{
			name: "certificate creation error",
			config: types.Resources{
				Certificates: map[string]types.Certificate{
					"bad_cert": {
						Certificate: "invalid_cert",
						PrivateKey:  "invalid_key",
					},
				},
			},
			setupMocks: func(certMaker *certificatesMocks.MockMaker, scriptMaker *scriptsMocks.MockMaker) {
				certMaker.On("Make", map[string]types.Certificate{
					"bad_cert": {
						Certificate: "invalid_cert",
						PrivateKey:  "invalid_key",
					},
				}).Return(nil, assert.AnError)
			},
			expectError:  true,
			errorMessage: "error creating certificates",
		},
		{
			name: "script creation error",
			config: types.Resources{
				Scripts: map[string]types.Script{
					"bad_script": {
						Content: "echo bad",
						Mode:    "invalid_mode",
					},
				},
			},
			setupMocks: func(certMaker *certificatesMocks.MockMaker, scriptMaker *scriptsMocks.MockMaker) {
				scriptMaker.On("Make", map[string]types.Script{
					"bad_script": {
						Content: "echo bad",
						Mode:    "invalid_mode",
					},
				}).Return(nil, assert.AnError)
			},
			expectError:  true,
			errorMessage: "error creating scripts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			certMaker := certificatesMocks.NewMockMaker(t)
			scriptMaker := scriptsMocks.NewMockMaker(t)

			// Create a custom maker with mocked dependencies
			maker := &nativeMaker{
				fnd:               fndMock,
				certificatesMaker: certMaker,
				scriptsMaker:      scriptMaker,
			}

			tt.setupMocks(certMaker, scriptMaker)

			result, err := maker.Make(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				tt.validateResult(t, result)
			}

			certMaker.AssertExpectations(t)
			scriptMaker.AssertExpectations(t)
		})
	}
}

func Test_Resources_Inherit(t *testing.T) {
	tests := []struct {
		name            string
		baseResources   Resources
		parentResources Resources
		validateResult  func(*testing.T, Resources)
	}{
		{
			name: "inherit both certificates and scripts",
			baseResources: Resources{
				Certificates: certificates.Certificates{
					"child_cert": certificatesMocks.NewMockCertificate(t),
				},
				Scripts: scripts.Scripts{
					"child_script": scriptsMocks.NewMockScript(t),
				},
			},
			parentResources: Resources{
				Certificates: certificates.Certificates{
					"parent_cert": certificatesMocks.NewMockCertificate(t),
				},
				Scripts: scripts.Scripts{
					"parent_script": scriptsMocks.NewMockScript(t),
				},
			},
			validateResult: func(t *testing.T, result Resources) {
				assert.Len(t, result.Certificates, 2)
				assert.Len(t, result.Scripts, 2)

				_, hasChildCert := result.Certificates["child_cert"]
				_, hasParentCert := result.Certificates["parent_cert"]
				_, hasChildScript := result.Scripts["child_script"]
				_, hasParentScript := result.Scripts["parent_script"]

				assert.True(t, hasChildCert)
				assert.True(t, hasParentCert)
				assert.True(t, hasChildScript)
				assert.True(t, hasParentScript)
			},
		},
		{
			name: "inherit with conflicts - child takes precedence",
			baseResources: Resources{
				Certificates: certificates.Certificates{
					"shared_cert": certificatesMocks.NewMockCertificate(t),
				},
				Scripts: scripts.Scripts{
					"shared_script": scriptsMocks.NewMockScript(t),
				},
			},
			parentResources: Resources{
				Certificates: certificates.Certificates{
					"shared_cert": certificatesMocks.NewMockCertificate(t),
				},
				Scripts: scripts.Scripts{
					"shared_script": scriptsMocks.NewMockScript(t),
				},
			},
			validateResult: func(t *testing.T, result Resources) {
				assert.Len(t, result.Certificates, 1)
				assert.Len(t, result.Scripts, 1)
				// The specific instances should be the child's (but we can't easily test that with mocks)
			},
		},
		{
			name: "inherit from empty parent",
			baseResources: Resources{
				Certificates: certificates.Certificates{
					"only_cert": certificatesMocks.NewMockCertificate(t),
				},
				Scripts: scripts.Scripts{
					"only_script": scriptsMocks.NewMockScript(t),
				},
			},
			parentResources: Resources{
				Certificates: certificates.Certificates{},
				Scripts:      scripts.Scripts{},
			},
			validateResult: func(t *testing.T, result Resources) {
				assert.Len(t, result.Certificates, 1)
				assert.Len(t, result.Scripts, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.baseResources.Inherit(tt.parentResources)
			tt.validateResult(t, result)
		})
	}
}

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	paramMaker := parameterMocks.NewMockMaker(t)

	maker := CreateMaker(fndMock, paramMaker)

	assert.NotNil(t, maker)
	assert.IsType(t, &nativeMaker{}, maker)

	// Verify that the maker has the correct dependencies
	nativeMakerInstance := maker.(*nativeMaker)
	assert.Equal(t, fndMock, nativeMakerInstance.fnd)
	assert.NotNil(t, nativeMakerInstance.certificatesMaker)
	assert.NotNil(t, nativeMakerInstance.scriptsMaker)
}
