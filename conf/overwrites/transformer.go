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

package overwrites

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/parser"
	"github.com/bukka/wst/conf/types"
	"reflect"
	"strings"
)

type Transformer interface {
	Transform(overwrites map[string]string) (*types.Config, error)
}

type nativeTransformer struct {
	fnd    app.Foundation
	parser parser.Parser
}

func CreateTransformer(fnd app.Foundation, parser parser.Parser) Transformer {
	return &nativeTransformer{
		fnd:    fnd,
		parser: parser,
	}
}

func (t *nativeTransformer) Transform(overwrites map[string]string) (*types.Config, error) {
	config := &types.Config{}
	dst := reflect.ValueOf(config)
	for key, val := range overwrites {
		ptr := strings.Split(key, ".")
		err := t.transformStruct(dst, ptr, val)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func (t *nativeTransformer) transformStruct(dst reflect.Value, ptr []string, val string) error {
	return nil
}
