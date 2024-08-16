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
	"context"
	"github.com/bukka/wst/app"
	"time"
)

type Maker interface {
	MakeData() Data
	MakeBackgroundContext() context.Context
	MakeContextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc)
}

type syncMaker struct {
	fnd app.Foundation
}

func (s *syncMaker) MakeBackgroundContext() context.Context {
	return context.Background()
}

func (s *syncMaker) MakeContextWithTimeout(
	ctx context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func (s *syncMaker) MakeData() Data {
	return &syncData{
		fnd: s.fnd,
	}
}

func CreateMaker(fnd app.Foundation) Maker {
	return &syncMaker{
		fnd: fnd,
	}
}
