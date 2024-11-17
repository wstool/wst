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

type MatchType string

const (
	MatchTypeNone   MatchType = ""
	MatchTypeExact  MatchType = "exact"
	MatchTypeRegexp MatchType = "regexp"
)

type OrderType string

const (
	OrderTypeFixed  OrderType = "fixed"
	OrderTypeRandom OrderType = "random"
)

type OutputType string

const (
	OutputTypeStdout OutputType = "stdout"
	OutputTypeStderr OutputType = "stderr"
	OutputTypeAny    OutputType = "any"
)
