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
	"fmt"
	"github.com/wstool/wst/conf/types"
)

func (m *nativeMaker) MakeResponseExpectation(
	config *types.ResponseExpectation,
) (*ResponseExpectation, error) {
	matchType := MatchType(config.Body.Match)
	if matchType != MatchTypeNone &&
		matchType != MatchTypeExact &&
		matchType != MatchTypeRegexp &&
		matchType != MatchTypePrefix &&
		matchType != MatchTypeSuffix &&
		matchType != MatchTypeInfix {
		return nil, fmt.Errorf("invalid match type: %v", config.Body.Match)
	}

	return &ResponseExpectation{
		Request:            config.Request,
		Headers:            config.Headers,
		BodyContent:        config.Body.Content,
		BodyMatch:          matchType,
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
