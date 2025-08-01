// Copyright 2024 Jakub Zelenka and The WST Authors
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

package service

import "github.com/wstool/wst/run/resources/certificates"

// TemplateService defines template specific service subset
type TemplateService interface {
	LocalAddress() string
	LocalPort() int32
	PrivateAddress() string
	PrivateUrl(scheme string) (string, error)
	UdsPath(...string) (string, error)
	Executable() (string, error)
	Pid() (int, error)
	ConfDir() (string, error)
	RunDir() (string, error)
	ScriptDir() (string, error)
	Group() string
	User() string
	EnvironmentConfigPaths() map[string]string
	FindCertificate(name string) (*certificates.RenderedCertificate, error)
}
