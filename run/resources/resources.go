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

package resources

import (
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/resources/scripts"
)

type Resources struct {
	Certificates certificates.Certificates
	Scripts      scripts.Scripts
}

func (r Resources) Inherit(pr Resources) Resources {
	r.Certificates.Inherit(pr.Certificates)
	r.Scripts.Inherit(pr.Scripts)
	return r
}

type Maker interface {
	Make(config types.Resources) (*Resources, error)
}

type nativeMaker struct {
	fnd               app.Foundation
	certificatesMaker certificates.Maker
	scriptsMaker      scripts.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker parameters.Maker) Maker {
	return &nativeMaker{
		fnd:               fnd,
		certificatesMaker: certificates.CreateMaker(fnd),
		scriptsMaker:      scripts.CreateMaker(fnd, parametersMaker),
	}
}

func (m *nativeMaker) Make(config types.Resources) (*Resources, error) {
	var certs certificates.Certificates
	var scts scripts.Scripts
	var err error

	// Create certificates if config provided
	if config.Certificates != nil {
		certs, err = m.certificatesMaker.Make(config.Certificates)
		if err != nil {
			return nil, errors.Errorf("error creating certificates: %v", err)
		}
	}

	// Create scripts if config provided
	if config.Scripts != nil {
		scts, err = m.scriptsMaker.Make(config.Scripts)
		if err != nil {
			return nil, errors.Errorf("error creating scripts: %v", err)
		}
	}

	return &Resources{
		Certificates: certs,
		Scripts:      scts,
	}, nil
}
