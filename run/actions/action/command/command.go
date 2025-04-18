package command

import (
	"context"
	"fmt"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/services"
	"time"
)

type Maker interface {
	Make(
		config *types.CommandAction,
		sl services.ServiceLocator,
		defaultTimeout int,
	) (action.Action, error)
}

type ActionMaker struct {
	fnd         app.Foundation
	outputMaker output.Maker
}

func CreateActionMaker(fnd app.Foundation) *ActionMaker {
	return &ActionMaker{
		fnd:         fnd,
		outputMaker: output.CreateMaker(fnd),
	}
}

func (m *ActionMaker) Make(
	config *types.CommandAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	svc, err := sl.Find(config.Service)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	var cmd *environment.Command
	switch c := config.Command.(type) {
	case *types.ShellCommand:
		// If it is a ShellCommand, use the specified shell and command
		cmd = &environment.Command{
			Name: config.Shell,
			Args: []string{"-c", c.Command},
		}
	case *types.ArgsCommand:
		// If it is an ArgsCommand, split the args into Name and Args
		if len(c.Args) == 0 {
			return nil, fmt.Errorf("ArgsCommand requires at least one argument")
		}
		cmd = &environment.Command{
			Name: c.Args[0],
			Args: c.Args[1:],
		}
	default:
		// Unsupported Command type - this should not happen
		return nil, fmt.Errorf("unsupported command type: %T", config.Command)
	}

	return &Action{
		fnd:         m.fnd,
		service:     svc,
		timeout:     time.Duration(config.Timeout * 1e6),
		when:        action.When(config.When),
		id:          config.Id,
		command:     cmd,
		outputMaker: m.outputMaker,
	}, nil
}

type Action struct {
	fnd         app.Foundation
	service     services.Service
	timeout     time.Duration
	when        action.When
	id          string
	command     *environment.Command
	outputMaker output.Maker
}

func (a *Action) When() action.When {
	return a.when
}

func (a *Action) Timeout() time.Duration {
	return a.timeout
}

func (a *Action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing command action")

	// Send the request.
	oc := a.outputMaker.MakeCollector(fmt.Sprintf("action %s", a.id))
	if err := a.service.ExecCommand(ctx, a.command, oc); err != nil {
		return false, err
	}

	// Store the ResponseData in runData.
	key := fmt.Sprintf("command/%s", a.id)
	a.fnd.Logger().Debugf("Storing command output %s", key)
	if err := runData.Store(key, oc); err != nil {
		return false, err
	}

	return true, nil
}
