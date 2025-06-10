// Copyright 2025 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package certificates

import (
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
)

const (
	CertificateExtension = ".crt"
	PrivateKeyExtension  = ".key"
)

type Certificate interface {
	CertificateData() string
	CertificateName() string
	PrivateKeyData() string
	PrivateKeyName() string
}

type RenderedCertificate struct {
	Certificate
	PrivateKeyFilePath  string
	CertificateFilePath string
}

func (c Certificates) Inherit(pc Certificates) Certificates {
	for key, value := range pc {
		if _, ok := c[key]; !ok {
			c[key] = value
		}
	}
	return c
}

type Certificates map[string]Certificate

type Maker interface {
	Make(config map[string]types.Certificate) (Certificates, error)
}

type nativeMaker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{
		fnd: fnd,
	}
}

func (m *nativeMaker) Make(config map[string]types.Certificate) (Certificates, error) {
	certs := make(Certificates)
	for certName, configCert := range config {
		certs[certName] = &nativeCertificate{
			name:        certName,
			privateKey:  configCert.PrivateKey,
			certificate: configCert.Certificate,
		}
	}
	return certs, nil
}

type nativeCertificate struct {
	name        string
	certificate string
	privateKey  string
}

func (s *nativeCertificate) CertificateData() string {
	return s.certificate
}

func (s *nativeCertificate) CertificateName() string {
	return s.name + CertificateExtension
}

func (s *nativeCertificate) PrivateKeyData() string {
	return s.privateKey
}

func (s *nativeCertificate) PrivateKeyName() string {
	return s.name + PrivateKeyExtension
}
