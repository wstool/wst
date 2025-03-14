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

package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Command interface {
	IsRunning() bool
	Start() error
	Run() error
	ProcessPid() int
	ProcessSignal(sig os.Signal) error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	SetStdout(stdout io.Writer)
	SetStderr(stdout io.Writer)
	SetSysProcAttr(attr *syscall.SysProcAttr)
	String() string
	Wait() error
}

func NewExecCommand(ctx context.Context, name string, args []string) Command {
	return &ExecCommand{
		cmd: exec.CommandContext(ctx, name, args...),
	}
}

type ExecCommand struct {
	cmd *exec.Cmd
}

func (c ExecCommand) String() string {
	return c.cmd.String()
}

func (c ExecCommand) IsRunning() bool {
	if c.cmd.Process == nil {
		return false
	}
	process, err := os.FindProcess(c.cmd.Process.Pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func (c ExecCommand) Start() error {
	return c.cmd.Start()
}

func (c ExecCommand) Run() error {
	return c.cmd.Run()
}

func (c ExecCommand) ProcessPid() int {
	return c.cmd.Process.Pid
}

func (c ExecCommand) ProcessSignal(sig os.Signal) error {
	return c.cmd.Process.Signal(sig)
}

func (c ExecCommand) StdoutPipe() (io.ReadCloser, error) {
	return c.cmd.StdoutPipe()
}

func (c ExecCommand) StderrPipe() (io.ReadCloser, error) {
	return c.cmd.StderrPipe()
}

func (c ExecCommand) SetStdout(stdout io.Writer) {
	c.cmd.Stdout = stdout
}

func (c ExecCommand) SetStderr(stderr io.Writer) {
	c.cmd.Stderr = stderr
}

func (c ExecCommand) SetSysProcAttr(attr *syscall.SysProcAttr) {
	c.cmd.SysProcAttr = attr
}

func (c ExecCommand) Wait() error {
	return c.cmd.Wait()
}

func NewDryRunCommand(name string, args []string) Command {
	return &DryRunCommand{}
}

type DryRunCommand struct {
	name string
	args []string
}

func (c DryRunCommand) String() string {
	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}

func (c DryRunCommand) IsRunning() bool {
	return true
}

func (c DryRunCommand) Start() error {
	return nil
}

func (c DryRunCommand) Run() error {
	return nil
}

func (c DryRunCommand) ProcessPid() int {
	return 12345
}

func (c DryRunCommand) ProcessSignal(sig os.Signal) error {
	return nil
}

func (c DryRunCommand) StdoutPipe() (io.ReadCloser, error) {
	return &DummyReaderCloser{}, nil
}

func (c DryRunCommand) StderrPipe() (io.ReadCloser, error) {
	return &DummyReaderCloser{}, nil
}

func (c DryRunCommand) SetStdout(stdout io.Writer) {}

func (c DryRunCommand) SetStderr(stdout io.Writer) {}

func (c DryRunCommand) SetSysProcAttr(attr *syscall.SysProcAttr) {}

func (c DryRunCommand) Wait() error {
	return nil
}

type DummyReaderCloser struct{}

func (drc *DummyReaderCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (drc *DummyReaderCloser) Close() error {
	return nil
}
