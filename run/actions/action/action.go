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

package action

import (
	"context"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/services"
	"time"
)

type Action interface {
	Execute(ctx context.Context, runData runtime.Data) (bool, error)
	Timeout() time.Duration
	When() When
}

type Maker interface {
	MakeAction(config types.Action, sl services.ServiceLocator, defaultTimeout int) (Action, error)
}

type When string

const (
	Always    When = "always"
	OnSuccess When = "on_success"
	OnFailure When = "on_failure"
)
