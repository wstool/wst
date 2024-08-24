package app

import (
	"context"
	"io"
	"os"
	"os/exec"
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
}

func NewExecCommand(ctx context.Context, name string, args []string) Command {
	return &ExecCommand{
		cmd: exec.CommandContext(ctx, name, args...),
	}
}

type ExecCommand struct {
	cmd *exec.Cmd
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

func NewDryRunCommand() Command {
	return &DryRunCommand{}
}

type DryRunCommand struct {
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

type DummyReaderCloser struct{}

func (drc *DummyReaderCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (drc *DummyReaderCloser) Close() error {
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
