package certificates

import (
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	"testing"
)

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name             string
		config           map[string]types.Certificate
		getExpectedCerts func() Certificates
		expectError      bool
		errorMessage     string
	}{
		{
			name: "successful certificate creation",
			config: map[string]types.Certificate{
				"web_ssl": {
					Certificate: "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKoK\n-----END CERTIFICATE-----",
					PrivateKey:  "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQ\n-----END PRIVATE KEY-----",
				},
				"api_ssl": {
					Certificate: "/etc/ssl/certs/api.crt",
					PrivateKey:  "/etc/ssl/private/api.key",
				},
			},
			getExpectedCerts: func() Certificates {
				return Certificates{
					"web_ssl": &nativeCertificate{
						certificate: "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKoK\n-----END CERTIFICATE-----",
						privateKey:  "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQ\n-----END PRIVATE KEY-----",
					},
					"api_ssl": &nativeCertificate{
						certificate: "/etc/ssl/certs/api.crt",
						privateKey:  "/etc/ssl/private/api.key",
					},
				}
			},
		},
		{
			name:   "empty certificate config",
			config: map[string]types.Certificate{},
			getExpectedCerts: func() Certificates {
				return Certificates{}
			},
		},
		{
			name: "single certificate with file paths",
			config: map[string]types.Certificate{
				"test_cert": {
					Certificate: "./certs/server.crt",
					PrivateKey:  "./certs/server.key",
				},
			},
			getExpectedCerts: func() Certificates {
				return Certificates{
					"test_cert": &nativeCertificate{
						certificate: "./certs/server.crt",
						privateKey:  "./certs/server.key",
					},
				}
			},
		},
		{
			name: "mixed content types",
			config: map[string]types.Certificate{
				"pem_cert": {
					Certificate: "-----BEGIN CERTIFICATE-----\ncert_content\n-----END CERTIFICATE-----",
					PrivateKey:  "-----BEGIN PRIVATE KEY-----\nkey_content\n-----END PRIVATE KEY-----",
				},
				"file_cert": {
					Certificate: "file:-----BEGIN-CERT.crt",
					PrivateKey:  "file:-----BEGIN-KEY.key",
				},
			},
			getExpectedCerts: func() Certificates {
				return Certificates{
					"pem_cert": &nativeCertificate{
						certificate: "-----BEGIN CERTIFICATE-----\ncert_content\n-----END CERTIFICATE-----",
						privateKey:  "-----BEGIN PRIVATE KEY-----\nkey_content\n-----END PRIVATE KEY-----",
					},
					"file_cert": &nativeCertificate{
						certificate: "file:-----BEGIN-CERT.crt",
						privateKey:  "file:-----BEGIN-KEY.key",
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			maker := CreateMaker(fndMock)

			actualCerts, err := maker.Make(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				expectedCerts := tt.getExpectedCerts()
				assert.Equal(t, expectedCerts, actualCerts)
			}
		})
	}
}

// Helper function to create a nativeCertificate instance for testing
func getTestCertificate(t *testing.T) *nativeCertificate {
	return &nativeCertificate{
		certificate: "-----BEGIN CERTIFICATE-----\ntest_cert_content\n-----END CERTIFICATE-----",
		privateKey:  "-----BEGIN PRIVATE KEY-----\ntest_key_content\n-----END PRIVATE KEY-----",
	}
}

func TestNativeCertificate_Certificate(t *testing.T) {
	cert := getTestCertificate(t)
	expected := "-----BEGIN CERTIFICATE-----\ntest_cert_content\n-----END CERTIFICATE-----"
	assert.Equal(t, expected, cert.Certificate(), "Certificate method should return the correct certificate content")
}

func TestNativeCertificate_PrivateKey(t *testing.T) {
	cert := getTestCertificate(t)
	expected := "-----BEGIN PRIVATE KEY-----\ntest_key_content\n-----END PRIVATE KEY-----"
	assert.Equal(t, expected, cert.PrivateKey(), "PrivateKey method should return the correct private key content")
}

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	maker := CreateMaker(fndMock)

	assert.NotNil(t, maker, "CreateMaker should return a non-nil Maker instance")
	assert.IsType(t, &nativeMaker{}, maker, "CreateMaker should return a nativeMaker instance")
}
