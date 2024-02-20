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

package types

type Parameter interface {
	IsNil() bool
	GetBool() bool
	GetInt() int
	GetFloat() float64
	GetString() string
	GetParameters() Parameters
}

type ParameterType int

const (
	ParameterTypeNil = iota
	ParameterTypeBool
	ParameterTypeInt
	ParameterTypeFloat
	ParameterTypeString
	ParameterTypeParameters
)

type ParameterData struct {
	Type        ParameterType
	BoolValue   bool
	IntValue    int
	FloatValue  float64
	StringValue string
	Parameters  Parameters
}

type Parameters map[string]Parameter

func (pd ParameterData) IsNil() bool {
	return pd.Type == ParameterTypeNil
}

func (pd ParameterData) GetBool() bool {
	if pd.Type != ParameterTypeBool {
		return false // or panic/error
	}
	return pd.BoolValue
}

func (pd ParameterData) GetInt() int {
	if pd.Type != ParameterTypeInt {
		return 0 // or panic/error
	}
	return pd.IntValue
}

func (pd ParameterData) GetFloat() float64 {
	if pd.Type != ParameterTypeFloat {
		return 0.0 // or panic/error
	}
	return pd.FloatValue
}

func (pd ParameterData) GetString() string {
	if pd.Type != ParameterTypeString {
		return "" // or panic/error
	}
	return pd.StringValue
}

func (pd ParameterData) GetParameters() Parameters {
	if pd.Type != ParameterTypeParameters || pd.Parameters == nil {
		return nil // or panic/error
	}
	return pd.Parameters
}
