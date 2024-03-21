package app

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type Command interface {
	Start() error
	Run() error
	ProcessPid() int
	ProcessSignal(sig os.Signal) error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}

func NewExecCommand(ctx context.Context, name string, args []string) Command {
	return &ExecCommand{
		cmd: exec.CommandContext(ctx, name, args...),
	}
}

type ExecCommand struct {
	cmd *exec.Cmd
}

func (c ExecCommand) Start() error {
	return c.Start()
}

func (c ExecCommand) Run() error {
	return c.Run()
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

func NewDryRunCommand() Command {
	return &DryRunCommand{}
}

type DryRunCommand struct {
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
