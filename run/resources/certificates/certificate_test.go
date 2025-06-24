package certificates

import (
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	"testing"
)

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]types.Certificate
		validateResult func(t *testing.T, certs Certificates)
		expectError    bool
		errorMessage   string
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
			validateResult: func(t *testing.T, certs Certificates) {
				assert.Len(t, certs, 2)

				// Test web_ssl certificate
				webCert, exists := certs["web_ssl"]
				assert.True(t, exists)
				assert.Equal(t, "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKoK\n-----END CERTIFICATE-----", webCert.CertificateData())
				assert.Equal(t, "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQ\n-----END PRIVATE KEY-----", webCert.PrivateKeyData())
				assert.Equal(t, "web_ssl.crt", webCert.CertificateName())
				assert.Equal(t, "web_ssl.key", webCert.PrivateKeyName())

				// Test api_ssl certificate
				apiCert, exists := certs["api_ssl"]
				assert.True(t, exists)
				assert.Equal(t, "/etc/ssl/certs/api.crt", apiCert.CertificateData())
				assert.Equal(t, "/etc/ssl/private/api.key", apiCert.PrivateKeyData())
				assert.Equal(t, "api_ssl.crt", apiCert.CertificateName())
				assert.Equal(t, "api_ssl.key", apiCert.PrivateKeyName())
			},
		},
		{
			name:   "empty certificate config",
			config: map[string]types.Certificate{},
			validateResult: func(t *testing.T, certs Certificates) {
				assert.Len(t, certs, 0)
				assert.NotNil(t, certs)
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
			validateResult: func(t *testing.T, certs Certificates) {
				assert.Len(t, certs, 1)

				cert, exists := certs["test_cert"]
				assert.True(t, exists)
				assert.Equal(t, "./certs/server.crt", cert.CertificateData())
				assert.Equal(t, "./certs/server.key", cert.PrivateKeyData())
				assert.Equal(t, "test_cert.crt", cert.CertificateName())
				assert.Equal(t, "test_cert.key", cert.PrivateKeyName())
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
			validateResult: func(t *testing.T, certs Certificates) {
				assert.Len(t, certs, 2)

				// Test pem_cert
				pemCert, exists := certs["pem_cert"]
				assert.True(t, exists)
				assert.Equal(t, "-----BEGIN CERTIFICATE-----\ncert_content\n-----END CERTIFICATE-----", pemCert.CertificateData())
				assert.Equal(t, "-----BEGIN PRIVATE KEY-----\nkey_content\n-----END PRIVATE KEY-----", pemCert.PrivateKeyData())
				assert.Equal(t, "pem_cert.crt", pemCert.CertificateName())
				assert.Equal(t, "pem_cert.key", pemCert.PrivateKeyName())

				// Test file_cert
				fileCert, exists := certs["file_cert"]
				assert.True(t, exists)
				assert.Equal(t, "file:-----BEGIN-CERT.crt", fileCert.CertificateData())
				assert.Equal(t, "file:-----BEGIN-KEY.key", fileCert.PrivateKeyData())
				assert.Equal(t, "file_cert.crt", fileCert.CertificateName())
				assert.Equal(t, "file_cert.key", fileCert.PrivateKeyName())
			},
		},
		{
			name: "certificate with special characters in name",
			config: map[string]types.Certificate{
				"test-cert_123": {
					Certificate: "cert-content",
					PrivateKey:  "key-content",
				},
			},
			validateResult: func(t *testing.T, certs Certificates) {
				assert.Len(t, certs, 1)

				cert, exists := certs["test-cert_123"]
				assert.True(t, exists)
				assert.Equal(t, "test-cert_123.crt", cert.CertificateName())
				assert.Equal(t, "test-cert_123.key", cert.PrivateKeyName())
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
				tt.validateResult(t, actualCerts)
			}
		})
	}
}

func Test_nativeCertificate_methods(t *testing.T) {
	tests := []struct {
		name             string
		certName         string
		certificateData  string
		privateKeyData   string
		expectedCertName string
		expectedKeyName  string
	}{
		{
			name:             "standard certificate",
			certName:         "web_ssl",
			certificateData:  "-----BEGIN CERTIFICATE-----\ncert_data\n-----END CERTIFICATE-----",
			privateKeyData:   "-----BEGIN PRIVATE KEY-----\nkey_data\n-----END PRIVATE KEY-----",
			expectedCertName: "web_ssl.crt",
			expectedKeyName:  "web_ssl.key",
		},
		{
			name:             "certificate with underscores",
			certName:         "test_cert_name",
			certificateData:  "file_path_to_cert",
			privateKeyData:   "file_path_to_key",
			expectedCertName: "test_cert_name.crt",
			expectedKeyName:  "test_cert_name.key",
		},
		{
			name:             "certificate with hyphens",
			certName:         "api-server-ssl",
			certificateData:  "/etc/ssl/api.crt",
			privateKeyData:   "/etc/ssl/api.key",
			expectedCertName: "api-server-ssl.crt",
			expectedKeyName:  "api-server-ssl.key",
		},
		{
			name:             "empty certificate data",
			certName:         "empty_cert",
			certificateData:  "",
			privateKeyData:   "",
			expectedCertName: "empty_cert.crt",
			expectedKeyName:  "empty_cert.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := &nativeCertificate{
				name:        tt.certName,
				certificate: tt.certificateData,
				privateKey:  tt.privateKeyData,
			}

			// Test all methods
			assert.Equal(t, tt.certificateData, cert.CertificateData())
			assert.Equal(t, tt.privateKeyData, cert.PrivateKeyData())
			assert.Equal(t, tt.expectedCertName, cert.CertificateName())
			assert.Equal(t, tt.expectedKeyName, cert.PrivateKeyName())
		})
	}
}

func Test_Certificates_Inherit(t *testing.T) {
	tests := []struct {
		name           string
		baseCerts      Certificates
		parentCerts    Certificates
		expectedResult Certificates
		validateResult func(t *testing.T, result Certificates)
	}{
		{
			name: "inherit from parent with no conflicts",
			baseCerts: Certificates{
				"child_cert": &nativeCertificate{
					name:        "child_cert",
					certificate: "child_cert_data",
					privateKey:  "child_key_data",
				},
			},
			parentCerts: Certificates{
				"parent_cert": &nativeCertificate{
					name:        "parent_cert",
					certificate: "parent_cert_data",
					privateKey:  "parent_key_data",
				},
			},
			validateResult: func(t *testing.T, result Certificates) {
				assert.Len(t, result, 2)

				// Child cert should remain unchanged
				childCert, exists := result["child_cert"]
				assert.True(t, exists)
				assert.Equal(t, "child_cert_data", childCert.CertificateData())
				assert.Equal(t, "child_key_data", childCert.PrivateKeyData())

				// Parent cert should be inherited
				parentCert, exists := result["parent_cert"]
				assert.True(t, exists)
				assert.Equal(t, "parent_cert_data", parentCert.CertificateData())
				assert.Equal(t, "parent_key_data", parentCert.PrivateKeyData())
			},
		},
		{
			name: "inherit with conflicts - child takes precedence",
			baseCerts: Certificates{
				"ssl_cert": &nativeCertificate{
					name:        "ssl_cert",
					certificate: "child_cert_data",
					privateKey:  "child_key_data",
				},
			},
			parentCerts: Certificates{
				"ssl_cert": &nativeCertificate{
					name:        "ssl_cert",
					certificate: "parent_cert_data",
					privateKey:  "parent_key_data",
				},
			},
			validateResult: func(t *testing.T, result Certificates) {
				assert.Len(t, result, 1)

				// Child cert should take precedence
				cert, exists := result["ssl_cert"]
				assert.True(t, exists)
				assert.Equal(t, "child_cert_data", cert.CertificateData())
				assert.Equal(t, "child_key_data", cert.PrivateKeyData())
			},
		},
		{
			name: "inherit from empty parent",
			baseCerts: Certificates{
				"only_cert": &nativeCertificate{
					name:        "only_cert",
					certificate: "only_cert_data",
					privateKey:  "only_key_data",
				},
			},
			parentCerts: Certificates{},
			validateResult: func(t *testing.T, result Certificates) {
				assert.Len(t, result, 1)

				cert, exists := result["only_cert"]
				assert.True(t, exists)
				assert.Equal(t, "only_cert_data", cert.CertificateData())
			},
		},
		{
			name:      "inherit into empty child",
			baseCerts: Certificates{},
			parentCerts: Certificates{
				"parent_only": &nativeCertificate{
					name:        "parent_only",
					certificate: "parent_only_cert",
					privateKey:  "parent_only_key",
				},
			},
			validateResult: func(t *testing.T, result Certificates) {
				assert.Len(t, result, 1)

				cert, exists := result["parent_only"]
				assert.True(t, exists)
				assert.Equal(t, "parent_only_cert", cert.CertificateData())
			},
		},
		{
			name:        "both empty",
			baseCerts:   Certificates{},
			parentCerts: Certificates{},
			validateResult: func(t *testing.T, result Certificates) {
				assert.Len(t, result, 0)
			},
		},
		{
			name: "complex inheritance scenario",
			baseCerts: Certificates{
				"web_ssl": &nativeCertificate{
					name:        "web_ssl",
					certificate: "child_web_cert",
					privateKey:  "child_web_key",
				},
				"child_only": &nativeCertificate{
					name:        "child_only",
					certificate: "child_only_cert",
					privateKey:  "child_only_key",
				},
			},
			parentCerts: Certificates{
				"web_ssl": &nativeCertificate{
					name:        "web_ssl",
					certificate: "parent_web_cert",
					privateKey:  "parent_web_key",
				},
				"api_ssl": &nativeCertificate{
					name:        "api_ssl",
					certificate: "parent_api_cert",
					privateKey:  "parent_api_key",
				},
				"parent_only": &nativeCertificate{
					name:        "parent_only",
					certificate: "parent_only_cert",
					privateKey:  "parent_only_key",
				},
			},
			validateResult: func(t *testing.T, result Certificates) {
				assert.Len(t, result, 4)

				// Child should override parent
				webCert, exists := result["web_ssl"]
				assert.True(t, exists)
				assert.Equal(t, "child_web_cert", webCert.CertificateData())

				// Child only cert should remain
				childCert, exists := result["child_only"]
				assert.True(t, exists)
				assert.Equal(t, "child_only_cert", childCert.CertificateData())

				// Parent certs should be inherited
				apiCert, exists := result["api_ssl"]
				assert.True(t, exists)
				assert.Equal(t, "parent_api_cert", apiCert.CertificateData())

				parentCert, exists := result["parent_only"]
				assert.True(t, exists)
				assert.Equal(t, "parent_only_cert", parentCert.CertificateData())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.baseCerts.Inherit(tt.parentCerts)

			// Test that the method returns the same map instance (method receiver)
			assert.Equal(t, &tt.baseCerts, &result)

			tt.validateResult(t, result)
		})
	}
}

func Test_constants(t *testing.T) {
	assert.Equal(t, ".crt", CertificateExtension)
	assert.Equal(t, ".key", PrivateKeyExtension)
}

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	maker := CreateMaker(fndMock)

	assert.NotNil(t, maker, "CreateMaker should return a non-nil Maker instance")
	assert.IsType(t, &nativeMaker{}, maker, "CreateMaker should return a nativeMaker instance")

	// Test that the foundation is properly set
	nativeMakerInstance := maker.(*nativeMaker)
	assert.Equal(t, fndMock, nativeMakerInstance.fnd)
}

func Test_RenderedCertificate_struct(t *testing.T) {
	// Test that RenderedCertificate can be created and fields are accessible
	cert := &nativeCertificate{
		name:        "test",
		certificate: "cert_data",
		privateKey:  "key_data",
	}

	rendered := RenderedCertificate{
		Certificate:               cert,
		PrivateKeyFilePath:        "/tmp/test.key",
		CertificateFilePath:       "/tmp/test.crt",
		PrivateKeySourceFilePath:  "/src/test.key",
		CertificateSourceFilePath: "/src/test.crt",
	}

	assert.Equal(t, cert, rendered.Certificate)
	assert.Equal(t, "/tmp/test.key", rendered.PrivateKeyFilePath)
	assert.Equal(t, "/tmp/test.crt", rendered.CertificateFilePath)
	assert.Equal(t, "/src/test.key", rendered.PrivateKeySourceFilePath)
	assert.Equal(t, "/src/test.crt", rendered.CertificateSourceFilePath)

	// Test that embedded interface methods are accessible
	assert.Equal(t, "cert_data", rendered.CertificateData())
	assert.Equal(t, "key_data", rendered.PrivateKeyData())
	assert.Equal(t, "test.crt", rendered.CertificateName())
	assert.Equal(t, "test.key", rendered.PrivateKeyName())
}
