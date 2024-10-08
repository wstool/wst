// Code generated by mockery v2.40.1. DO NOT EDIT.

package hooks

import (
	context "context"

	environment "github.com/wstool/wst/run/environments/environment"

	mock "github.com/stretchr/testify/mock"

	task "github.com/wstool/wst/run/environments/task"

	template "github.com/wstool/wst/run/services/template"
)

// MockHook is an autogenerated mock type for the Hook type
type MockHook struct {
	mock.Mock
}

type MockHook_Expecter struct {
	mock *mock.Mock
}

func (_m *MockHook) EXPECT() *MockHook_Expecter {
	return &MockHook_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: ctx, ss, tmpl, env, st
func (_m *MockHook) Execute(ctx context.Context, ss *environment.ServiceSettings, tmpl template.Template, env environment.Environment, st task.Task) (task.Task, error) {
	ret := _m.Called(ctx, ss, tmpl, env, st)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 task.Task
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *environment.ServiceSettings, template.Template, environment.Environment, task.Task) (task.Task, error)); ok {
		return rf(ctx, ss, tmpl, env, st)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *environment.ServiceSettings, template.Template, environment.Environment, task.Task) task.Task); ok {
		r0 = rf(ctx, ss, tmpl, env, st)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(task.Task)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *environment.ServiceSettings, template.Template, environment.Environment, task.Task) error); ok {
		r1 = rf(ctx, ss, tmpl, env, st)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockHook_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockHook_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
//   - ss *environment.ServiceSettings
//   - tmpl template.Template
//   - env environment.Environment
//   - st task.Task
func (_e *MockHook_Expecter) Execute(ctx interface{}, ss interface{}, tmpl interface{}, env interface{}, st interface{}) *MockHook_Execute_Call {
	return &MockHook_Execute_Call{Call: _e.mock.On("Execute", ctx, ss, tmpl, env, st)}
}

func (_c *MockHook_Execute_Call) Run(run func(ctx context.Context, ss *environment.ServiceSettings, tmpl template.Template, env environment.Environment, st task.Task)) *MockHook_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*environment.ServiceSettings), args[2].(template.Template), args[3].(environment.Environment), args[4].(task.Task))
	})
	return _c
}

func (_c *MockHook_Execute_Call) Return(_a0 task.Task, _a1 error) *MockHook_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockHook_Execute_Call) RunAndReturn(run func(context.Context, *environment.ServiceSettings, template.Template, environment.Environment, task.Task) (task.Task, error)) *MockHook_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockHook creates a new instance of MockHook. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockHook(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockHook {
	mock := &MockHook{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
