// Code generated by mockery v2.40.1. DO NOT EDIT.

package environment

import (
	context "context"
	io "io"

	environment "github.com/bukka/wst/run/environments/environment"

	mock "github.com/stretchr/testify/mock"

	os "os"

	output "github.com/bukka/wst/run/environments/environment/output"

	task "github.com/bukka/wst/run/environments/task"
)

// MockEnvironment is an autogenerated mock type for the Environment type
type MockEnvironment struct {
	mock.Mock
}

type MockEnvironment_Expecter struct {
	mock *mock.Mock
}

func (_m *MockEnvironment) EXPECT() *MockEnvironment_Expecter {
	return &MockEnvironment_Expecter{mock: &_m.Mock}
}

// ContainerRegistry provides a mock function with given fields:
func (_m *MockEnvironment) ContainerRegistry() *environment.ContainerRegistry {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ContainerRegistry")
	}

	var r0 *environment.ContainerRegistry
	if rf, ok := ret.Get(0).(func() *environment.ContainerRegistry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*environment.ContainerRegistry)
		}
	}

	return r0
}

// MockEnvironment_ContainerRegistry_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ContainerRegistry'
type MockEnvironment_ContainerRegistry_Call struct {
	*mock.Call
}

// ContainerRegistry is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) ContainerRegistry() *MockEnvironment_ContainerRegistry_Call {
	return &MockEnvironment_ContainerRegistry_Call{Call: _e.mock.On("ContainerRegistry")}
}

func (_c *MockEnvironment_ContainerRegistry_Call) Run(run func()) *MockEnvironment_ContainerRegistry_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_ContainerRegistry_Call) Return(_a0 *environment.ContainerRegistry) *MockEnvironment_ContainerRegistry_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_ContainerRegistry_Call) RunAndReturn(run func() *environment.ContainerRegistry) *MockEnvironment_ContainerRegistry_Call {
	_c.Call.Return(run)
	return _c
}

// Destroy provides a mock function with given fields: ctx
func (_m *MockEnvironment) Destroy(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Destroy")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockEnvironment_Destroy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Destroy'
type MockEnvironment_Destroy_Call struct {
	*mock.Call
}

// Destroy is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockEnvironment_Expecter) Destroy(ctx interface{}) *MockEnvironment_Destroy_Call {
	return &MockEnvironment_Destroy_Call{Call: _e.mock.On("Destroy", ctx)}
}

func (_c *MockEnvironment_Destroy_Call) Run(run func(ctx context.Context)) *MockEnvironment_Destroy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockEnvironment_Destroy_Call) Return(_a0 error) *MockEnvironment_Destroy_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_Destroy_Call) RunAndReturn(run func(context.Context) error) *MockEnvironment_Destroy_Call {
	_c.Call.Return(run)
	return _c
}

// ExecTaskCommand provides a mock function with given fields: ctx, ss, target, cmd
func (_m *MockEnvironment) ExecTaskCommand(ctx context.Context, ss *environment.ServiceSettings, target task.Task, cmd *environment.Command) error {
	ret := _m.Called(ctx, ss, target, cmd)

	if len(ret) == 0 {
		panic("no return value specified for ExecTaskCommand")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *environment.ServiceSettings, task.Task, *environment.Command) error); ok {
		r0 = rf(ctx, ss, target, cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockEnvironment_ExecTaskCommand_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecTaskCommand'
type MockEnvironment_ExecTaskCommand_Call struct {
	*mock.Call
}

// ExecTaskCommand is a helper method to define mock.On call
//   - ctx context.Context
//   - ss *environment.ServiceSettings
//   - target task.Task
//   - cmd *environment.Command
func (_e *MockEnvironment_Expecter) ExecTaskCommand(ctx interface{}, ss interface{}, target interface{}, cmd interface{}) *MockEnvironment_ExecTaskCommand_Call {
	return &MockEnvironment_ExecTaskCommand_Call{Call: _e.mock.On("ExecTaskCommand", ctx, ss, target, cmd)}
}

func (_c *MockEnvironment_ExecTaskCommand_Call) Run(run func(ctx context.Context, ss *environment.ServiceSettings, target task.Task, cmd *environment.Command)) *MockEnvironment_ExecTaskCommand_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*environment.ServiceSettings), args[2].(task.Task), args[3].(*environment.Command))
	})
	return _c
}

func (_c *MockEnvironment_ExecTaskCommand_Call) Return(_a0 error) *MockEnvironment_ExecTaskCommand_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_ExecTaskCommand_Call) RunAndReturn(run func(context.Context, *environment.ServiceSettings, task.Task, *environment.Command) error) *MockEnvironment_ExecTaskCommand_Call {
	_c.Call.Return(run)
	return _c
}

// ExecTaskSignal provides a mock function with given fields: ctx, ss, target, signal
func (_m *MockEnvironment) ExecTaskSignal(ctx context.Context, ss *environment.ServiceSettings, target task.Task, signal os.Signal) error {
	ret := _m.Called(ctx, ss, target, signal)

	if len(ret) == 0 {
		panic("no return value specified for ExecTaskSignal")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *environment.ServiceSettings, task.Task, os.Signal) error); ok {
		r0 = rf(ctx, ss, target, signal)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockEnvironment_ExecTaskSignal_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecTaskSignal'
type MockEnvironment_ExecTaskSignal_Call struct {
	*mock.Call
}

// ExecTaskSignal is a helper method to define mock.On call
//   - ctx context.Context
//   - ss *environment.ServiceSettings
//   - target task.Task
//   - signal os.Signal
func (_e *MockEnvironment_Expecter) ExecTaskSignal(ctx interface{}, ss interface{}, target interface{}, signal interface{}) *MockEnvironment_ExecTaskSignal_Call {
	return &MockEnvironment_ExecTaskSignal_Call{Call: _e.mock.On("ExecTaskSignal", ctx, ss, target, signal)}
}

func (_c *MockEnvironment_ExecTaskSignal_Call) Run(run func(ctx context.Context, ss *environment.ServiceSettings, target task.Task, signal os.Signal)) *MockEnvironment_ExecTaskSignal_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*environment.ServiceSettings), args[2].(task.Task), args[3].(os.Signal))
	})
	return _c
}

func (_c *MockEnvironment_ExecTaskSignal_Call) Return(_a0 error) *MockEnvironment_ExecTaskSignal_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_ExecTaskSignal_Call) RunAndReturn(run func(context.Context, *environment.ServiceSettings, task.Task, os.Signal) error) *MockEnvironment_ExecTaskSignal_Call {
	_c.Call.Return(run)
	return _c
}

// Init provides a mock function with given fields: ctx
func (_m *MockEnvironment) Init(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Init")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockEnvironment_Init_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Init'
type MockEnvironment_Init_Call struct {
	*mock.Call
}

// Init is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockEnvironment_Expecter) Init(ctx interface{}) *MockEnvironment_Init_Call {
	return &MockEnvironment_Init_Call{Call: _e.mock.On("Init", ctx)}
}

func (_c *MockEnvironment_Init_Call) Run(run func(ctx context.Context)) *MockEnvironment_Init_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockEnvironment_Init_Call) Return(_a0 error) *MockEnvironment_Init_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_Init_Call) RunAndReturn(run func(context.Context) error) *MockEnvironment_Init_Call {
	_c.Call.Return(run)
	return _c
}

// IsUsed provides a mock function with given fields:
func (_m *MockEnvironment) IsUsed() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsUsed")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockEnvironment_IsUsed_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsUsed'
type MockEnvironment_IsUsed_Call struct {
	*mock.Call
}

// IsUsed is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) IsUsed() *MockEnvironment_IsUsed_Call {
	return &MockEnvironment_IsUsed_Call{Call: _e.mock.On("IsUsed")}
}

func (_c *MockEnvironment_IsUsed_Call) Run(run func()) *MockEnvironment_IsUsed_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_IsUsed_Call) Return(_a0 bool) *MockEnvironment_IsUsed_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_IsUsed_Call) RunAndReturn(run func() bool) *MockEnvironment_IsUsed_Call {
	_c.Call.Return(run)
	return _c
}

// MarkUsed provides a mock function with given fields:
func (_m *MockEnvironment) MarkUsed() {
	_m.Called()
}

// MockEnvironment_MarkUsed_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MarkUsed'
type MockEnvironment_MarkUsed_Call struct {
	*mock.Call
}

// MarkUsed is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) MarkUsed() *MockEnvironment_MarkUsed_Call {
	return &MockEnvironment_MarkUsed_Call{Call: _e.mock.On("MarkUsed")}
}

func (_c *MockEnvironment_MarkUsed_Call) Run(run func()) *MockEnvironment_MarkUsed_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_MarkUsed_Call) Return() *MockEnvironment_MarkUsed_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockEnvironment_MarkUsed_Call) RunAndReturn(run func()) *MockEnvironment_MarkUsed_Call {
	_c.Call.Return(run)
	return _c
}

// Output provides a mock function with given fields: ctx, target, outputType
func (_m *MockEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	ret := _m.Called(ctx, target, outputType)

	if len(ret) == 0 {
		panic("no return value specified for Output")
	}

	var r0 io.Reader
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, task.Task, output.Type) (io.Reader, error)); ok {
		return rf(ctx, target, outputType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, task.Task, output.Type) io.Reader); ok {
		r0 = rf(ctx, target, outputType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, task.Task, output.Type) error); ok {
		r1 = rf(ctx, target, outputType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_Output_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Output'
type MockEnvironment_Output_Call struct {
	*mock.Call
}

// Output is a helper method to define mock.On call
//   - ctx context.Context
//   - target task.Task
//   - outputType output.Type
func (_e *MockEnvironment_Expecter) Output(ctx interface{}, target interface{}, outputType interface{}) *MockEnvironment_Output_Call {
	return &MockEnvironment_Output_Call{Call: _e.mock.On("Output", ctx, target, outputType)}
}

func (_c *MockEnvironment_Output_Call) Run(run func(ctx context.Context, target task.Task, outputType output.Type)) *MockEnvironment_Output_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(task.Task), args[2].(output.Type))
	})
	return _c
}

func (_c *MockEnvironment_Output_Call) Return(_a0 io.Reader, _a1 error) *MockEnvironment_Output_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_Output_Call) RunAndReturn(run func(context.Context, task.Task, output.Type) (io.Reader, error)) *MockEnvironment_Output_Call {
	_c.Call.Return(run)
	return _c
}

// PortsEnd provides a mock function with given fields:
func (_m *MockEnvironment) PortsEnd() int32 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for PortsEnd")
	}

	var r0 int32
	if rf, ok := ret.Get(0).(func() int32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int32)
	}

	return r0
}

// MockEnvironment_PortsEnd_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PortsEnd'
type MockEnvironment_PortsEnd_Call struct {
	*mock.Call
}

// PortsEnd is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) PortsEnd() *MockEnvironment_PortsEnd_Call {
	return &MockEnvironment_PortsEnd_Call{Call: _e.mock.On("PortsEnd")}
}

func (_c *MockEnvironment_PortsEnd_Call) Run(run func()) *MockEnvironment_PortsEnd_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_PortsEnd_Call) Return(_a0 int32) *MockEnvironment_PortsEnd_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_PortsEnd_Call) RunAndReturn(run func() int32) *MockEnvironment_PortsEnd_Call {
	_c.Call.Return(run)
	return _c
}

// PortsStart provides a mock function with given fields:
func (_m *MockEnvironment) PortsStart() int32 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for PortsStart")
	}

	var r0 int32
	if rf, ok := ret.Get(0).(func() int32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int32)
	}

	return r0
}

// MockEnvironment_PortsStart_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PortsStart'
type MockEnvironment_PortsStart_Call struct {
	*mock.Call
}

// PortsStart is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) PortsStart() *MockEnvironment_PortsStart_Call {
	return &MockEnvironment_PortsStart_Call{Call: _e.mock.On("PortsStart")}
}

func (_c *MockEnvironment_PortsStart_Call) Run(run func()) *MockEnvironment_PortsStart_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_PortsStart_Call) Return(_a0 int32) *MockEnvironment_PortsStart_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_PortsStart_Call) RunAndReturn(run func() int32) *MockEnvironment_PortsStart_Call {
	_c.Call.Return(run)
	return _c
}

// ReservePort provides a mock function with given fields:
func (_m *MockEnvironment) ReservePort() int32 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ReservePort")
	}

	var r0 int32
	if rf, ok := ret.Get(0).(func() int32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int32)
	}

	return r0
}

// MockEnvironment_ReservePort_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReservePort'
type MockEnvironment_ReservePort_Call struct {
	*mock.Call
}

// ReservePort is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) ReservePort() *MockEnvironment_ReservePort_Call {
	return &MockEnvironment_ReservePort_Call{Call: _e.mock.On("ReservePort")}
}

func (_c *MockEnvironment_ReservePort_Call) Run(run func()) *MockEnvironment_ReservePort_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_ReservePort_Call) Return(_a0 int32) *MockEnvironment_ReservePort_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_ReservePort_Call) RunAndReturn(run func() int32) *MockEnvironment_ReservePort_Call {
	_c.Call.Return(run)
	return _c
}

// RootPath provides a mock function with given fields: workspace
func (_m *MockEnvironment) RootPath(workspace string) string {
	ret := _m.Called(workspace)

	if len(ret) == 0 {
		panic("no return value specified for RootPath")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(workspace)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockEnvironment_RootPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RootPath'
type MockEnvironment_RootPath_Call struct {
	*mock.Call
}

// RootPath is a helper method to define mock.On call
//   - workspace string
func (_e *MockEnvironment_Expecter) RootPath(workspace interface{}) *MockEnvironment_RootPath_Call {
	return &MockEnvironment_RootPath_Call{Call: _e.mock.On("RootPath", workspace)}
}

func (_c *MockEnvironment_RootPath_Call) Run(run func(workspace string)) *MockEnvironment_RootPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockEnvironment_RootPath_Call) Return(_a0 string) *MockEnvironment_RootPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_RootPath_Call) RunAndReturn(run func(string) string) *MockEnvironment_RootPath_Call {
	_c.Call.Return(run)
	return _c
}

// RunTask provides a mock function with given fields: ctx, ss, cmd
func (_m *MockEnvironment) RunTask(ctx context.Context, ss *environment.ServiceSettings, cmd *environment.Command) (task.Task, error) {
	ret := _m.Called(ctx, ss, cmd)

	if len(ret) == 0 {
		panic("no return value specified for RunTask")
	}

	var r0 task.Task
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *environment.ServiceSettings, *environment.Command) (task.Task, error)); ok {
		return rf(ctx, ss, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *environment.ServiceSettings, *environment.Command) task.Task); ok {
		r0 = rf(ctx, ss, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(task.Task)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *environment.ServiceSettings, *environment.Command) error); ok {
		r1 = rf(ctx, ss, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_RunTask_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RunTask'
type MockEnvironment_RunTask_Call struct {
	*mock.Call
}

// RunTask is a helper method to define mock.On call
//   - ctx context.Context
//   - ss *environment.ServiceSettings
//   - cmd *environment.Command
func (_e *MockEnvironment_Expecter) RunTask(ctx interface{}, ss interface{}, cmd interface{}) *MockEnvironment_RunTask_Call {
	return &MockEnvironment_RunTask_Call{Call: _e.mock.On("RunTask", ctx, ss, cmd)}
}

func (_c *MockEnvironment_RunTask_Call) Run(run func(ctx context.Context, ss *environment.ServiceSettings, cmd *environment.Command)) *MockEnvironment_RunTask_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*environment.ServiceSettings), args[2].(*environment.Command))
	})
	return _c
}

func (_c *MockEnvironment_RunTask_Call) Return(_a0 task.Task, _a1 error) *MockEnvironment_RunTask_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_RunTask_Call) RunAndReturn(run func(context.Context, *environment.ServiceSettings, *environment.Command) (task.Task, error)) *MockEnvironment_RunTask_Call {
	_c.Call.Return(run)
	return _c
}

// ServiceAddress provides a mock function with given fields: serviceName, port
func (_m *MockEnvironment) ServiceAddress(serviceName string, port int32) string {
	ret := _m.Called(serviceName, port)

	if len(ret) == 0 {
		panic("no return value specified for ServiceAddress")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, int32) string); ok {
		r0 = rf(serviceName, port)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockEnvironment_ServiceAddress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServiceAddress'
type MockEnvironment_ServiceAddress_Call struct {
	*mock.Call
}

// ServiceAddress is a helper method to define mock.On call
//   - serviceName string
//   - port int32
func (_e *MockEnvironment_Expecter) ServiceAddress(serviceName interface{}, port interface{}) *MockEnvironment_ServiceAddress_Call {
	return &MockEnvironment_ServiceAddress_Call{Call: _e.mock.On("ServiceAddress", serviceName, port)}
}

func (_c *MockEnvironment_ServiceAddress_Call) Run(run func(serviceName string, port int32)) *MockEnvironment_ServiceAddress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int32))
	})
	return _c
}

func (_c *MockEnvironment_ServiceAddress_Call) Return(_a0 string) *MockEnvironment_ServiceAddress_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_ServiceAddress_Call) RunAndReturn(run func(string, int32) string) *MockEnvironment_ServiceAddress_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockEnvironment creates a new instance of MockEnvironment. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockEnvironment(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEnvironment {
	mock := &MockEnvironment{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
