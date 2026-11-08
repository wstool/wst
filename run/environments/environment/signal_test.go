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
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignalName(t *testing.T) {
	name, ok := SignalName(syscall.SIGTERM)
	assert.True(t, ok)
	assert.Equal(t, "TERM", name)
	name, ok = SignalName(syscall.SIGUSR2)
	assert.True(t, ok)
	assert.Equal(t, "USR2", name)
	name, ok = SignalName(syscall.SIGALRM)
	assert.False(t, ok)
	assert.Equal(t, "", name)
}
