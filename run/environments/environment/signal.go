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
	"os"
	"syscall"
)

// signalNames maps the supported signals to their names without the SIG prefix as understood by the kill command
// and by the container runtimes.
var signalNames = map[os.Signal]string{
	syscall.SIGHUP:  "HUP",
	syscall.SIGINT:  "INT",
	syscall.SIGQUIT: "QUIT",
	syscall.SIGKILL: "KILL",
	syscall.SIGUSR1: "USR1",
	syscall.SIGUSR2: "USR2",
	syscall.SIGTERM: "TERM",
}

// SignalName returns the name of the signal without the SIG prefix (e.g. TERM for SIGTERM). The second value is
// false if the signal is not supported.
func SignalName(signal os.Signal) (string, bool) {
	name, ok := signalNames[signal]
	return name, ok
}
