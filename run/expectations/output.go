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

func (m *nativeMaker) MakeOutputExpectation(
	config *types.OutputExpectation,
) (*OutputExpectation, error) {
	orderType := OrderType(config.Order)
	if orderType != OrderTypeFixed && orderType != OrderTypeRandom {
		return nil, fmt.Errorf("invalid order type: %v", config.Order)
	}

	matchType := MatchType(config.Match)
	if matchType != MatchTypeExact &&
		matchType != MatchTypeRegexp &&
		matchType != MatchTypePrefix &&
		matchType != MatchTypeSuffix &&
		matchType != MatchTypeInfix {
		return nil, fmt.Errorf("invalid match type: %v", config.Match)
	}

	outputType := OutputType(config.Type)
	if outputType != OutputTypeAny && outputType != OutputTypeStdout && outputType != OutputTypeStderr {
		return nil, fmt.Errorf("invalid output type: %v", config.Type)
	}

	return &OutputExpectation{
		Command:        config.Command,
		OrderType:      orderType,
		MatchType:      matchType,
		OutputType:     outputType,
		Messages:       config.Messages,
		RenderTemplate: config.RenderTemplate,
	}, nil
}

type OutputExpectation struct {
	Command        string
	OrderType      OrderType
	MatchType      MatchType
	OutputType     OutputType
	Messages       []string
	RenderTemplate bool
}
