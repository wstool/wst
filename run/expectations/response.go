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

package expectations

import (
	"github.com/pkg/errors"
	"github.com/wstool/wst/conf/types"
)

func (m *nativeMaker) MakeResponseExpectation(
	config *types.ResponseExpectation,
) (*ResponseExpectation, error) {
	match := MatchType(config.Body.Match)
	if match != MatchTypeNone && match != MatchTypeExact && match != MatchTypeRegexp {
		return nil, errors.Errorf("invalid MatchType: %v", config.Body.Match)
	}

	return &ResponseExpectation{
		Request:            config.Request,
		Headers:            config.Headers,
		BodyContent:        config.Body.Content,
		BodyMatch:          match,
		BodyRenderTemplate: config.Body.RenderTemplate,
		StatusCode:         config.Status,
	}, nil
}

type ResponseExpectation struct {
	Request            string
	Headers            types.Headers
	BodyContent        string
	BodyMatch          MatchType
	BodyRenderTemplate bool
	StatusCode         int
}
