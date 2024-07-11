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

package runtime

import (
	"github.com/bukka/wst/app"
	"sync"
)

// Data is the interface type to allow storage and retrieval of data across different actions.
type Data interface {
	Store(key string, value interface{}) error
	Load(key string) (interface{}, bool)
}

type Maker interface {
	Make() Data
}

type syncMaker struct {
	fnd app.Foundation
}

func (s *syncMaker) Make() Data {
	return &syncData{
		fnd: s.fnd,
	}
}

func CreateMaker(fnd app.Foundation) Maker {
	return &syncMaker{
		fnd: fnd,
	}
}

// runtimeDataImpl is an implementation of the RuntimeData interface.
type syncData struct {
	fnd  app.Foundation
	data sync.Map
}

// Store stores the value for a key.
func (rt *syncData) Store(key string, value interface{}) error {
	rt.data.Store(key, value)
	return nil
}

// Load loads the value for a key.
func (rt *syncData) Load(key string) (interface{}, bool) {
	return rt.data.Load(key)
}
