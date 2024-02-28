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

package environment

import (
	"context"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/bukka/wst/run/services"
	"github.com/bukka/wst/run/task"
	"io"
)

type Environment interface {
	Init(ctx context.Context) error
	Destroy(ctx context.Context) error
	RunTask(ctx context.Context, service services.Service) (task.Task, error)
	ExecTask(ctx context.Context, target task.Task, hook hooks.Hook) error
	Output(ctx context.Context) (io.Reader, error)
}
