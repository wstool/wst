// Code generated by mockery v2.40.1. DO NOT EDIT.

package sandbox

import (
	containers "github.com/wstool/wst/run/sandboxes/containers"
	dir "github.com/wstool/wst/run/sandboxes/dir"

	hooks "github.com/wstool/wst/run/sandboxes/hooks"

	mock "github.com/stretchr/testify/mock"

	sandbox "github.com/wstool/wst/run/sandboxes/sandbox"
)

// MockSandbox is an autogenerated mock type for the Sandbox type
type MockSandbox struct {
	mock.Mock
}

type MockSandbox_Expecter struct {
	mock *mock.Mock
}

func (_m *MockSandbox) EXPECT() *MockSandbox_Expecter {
	return &MockSandbox_Expecter{mock: &_m.Mock}
}

// Available provides a mock function with given fields:
func (_m *MockSandbox) Available() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Available")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockSandbox_Available_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Available'
type MockSandbox_Available_Call struct {
	*mock.Call
}

// Available is a helper method to define mock.On call
func (_e *MockSandbox_Expecter) Available() *MockSandbox_Available_Call {
	return &MockSandbox_Available_Call{Call: _e.mock.On("Available")}
}

func (_c *MockSandbox_Available_Call) Run(run func()) *MockSandbox_Available_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockSandbox_Available_Call) Return(_a0 bool) *MockSandbox_Available_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockSandbox_Available_Call) RunAndReturn(run func() bool) *MockSandbox_Available_Call {
	_c.Call.Return(run)
	return _c
}

// ContainerConfig provides a mock function with given fields:
func (_m *MockSandbox) ContainerConfig() *containers.ContainerConfig {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ContainerConfig")
	}

	var r0 *containers.ContainerConfig
	if rf, ok := ret.Get(0).(func() *containers.ContainerConfig); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*containers.ContainerConfig)
		}
	}

	return r0
}

// MockSandbox_ContainerConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ContainerConfig'
type MockSandbox_ContainerConfig_Call struct {
	*mock.Call
}

// ContainerConfig is a helper method to define mock.On call
func (_e *MockSandbox_Expecter) ContainerConfig() *MockSandbox_ContainerConfig_Call {
	return &MockSandbox_ContainerConfig_Call{Call: _e.mock.On("ContainerConfig")}
}

func (_c *MockSandbox_ContainerConfig_Call) Run(run func()) *MockSandbox_ContainerConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockSandbox_ContainerConfig_Call) Return(_a0 *containers.ContainerConfig) *MockSandbox_ContainerConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockSandbox_ContainerConfig_Call) RunAndReturn(run func() *containers.ContainerConfig) *MockSandbox_ContainerConfig_Call {
	_c.Call.Return(run)
	return _c
}

// Dir provides a mock function with given fields: dirType
func (_m *MockSandbox) Dir(dirType dir.DirType) (string, error) {
	ret := _m.Called(dirType)

	if len(ret) == 0 {
		panic("no return value specified for Dir")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(dir.DirType) (string, error)); ok {
		return rf(dirType)
	}
	if rf, ok := ret.Get(0).(func(dir.DirType) string); ok {
		r0 = rf(dirType)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(dir.DirType) error); ok {
		r1 = rf(dirType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockSandbox_Dir_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Dir'
type MockSandbox_Dir_Call struct {
	*mock.Call
}

// Dir is a helper method to define mock.On call
//   - dirType dir.DirType
func (_e *MockSandbox_Expecter) Dir(dirType interface{}) *MockSandbox_Dir_Call {
	return &MockSandbox_Dir_Call{Call: _e.mock.On("Dir", dirType)}
}

func (_c *MockSandbox_Dir_Call) Run(run func(dirType dir.DirType)) *MockSandbox_Dir_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(dir.DirType))
	})
	return _c
}

func (_c *MockSandbox_Dir_Call) Return(_a0 string, _a1 error) *MockSandbox_Dir_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockSandbox_Dir_Call) RunAndReturn(run func(dir.DirType) (string, error)) *MockSandbox_Dir_Call {
	_c.Call.Return(run)
	return _c
}

// Dirs provides a mock function with given fields:
func (_m *MockSandbox) Dirs() map[dir.DirType]string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Dirs")
	}

	var r0 map[dir.DirType]string
	if rf, ok := ret.Get(0).(func() map[dir.DirType]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[dir.DirType]string)
		}
	}

	return r0
}

// MockSandbox_Dirs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Dirs'
type MockSandbox_Dirs_Call struct {
	*mock.Call
}

// Dirs is a helper method to define mock.On call
func (_e *MockSandbox_Expecter) Dirs() *MockSandbox_Dirs_Call {
	return &MockSandbox_Dirs_Call{Call: _e.mock.On("Dirs")}
}

func (_c *MockSandbox_Dirs_Call) Run(run func()) *MockSandbox_Dirs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockSandbox_Dirs_Call) Return(_a0 map[dir.DirType]string) *MockSandbox_Dirs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockSandbox_Dirs_Call) RunAndReturn(run func() map[dir.DirType]string) *MockSandbox_Dirs_Call {
	_c.Call.Return(run)
	return _c
}

// Hook provides a mock function with given fields: hookType
func (_m *MockSandbox) Hook(hookType hooks.HookType) (hooks.Hook, error) {
	ret := _m.Called(hookType)

	if len(ret) == 0 {
		panic("no return value specified for Hook")
	}

	var r0 hooks.Hook
	var r1 error
	if rf, ok := ret.Get(0).(func(hooks.HookType) (hooks.Hook, error)); ok {
		return rf(hookType)
	}
	if rf, ok := ret.Get(0).(func(hooks.HookType) hooks.Hook); ok {
		r0 = rf(hookType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(hooks.Hook)
		}
	}

	if rf, ok := ret.Get(1).(func(hooks.HookType) error); ok {
		r1 = rf(hookType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockSandbox_Hook_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Hook'
type MockSandbox_Hook_Call struct {
	*mock.Call
}

// Hook is a helper method to define mock.On call
//   - hookType hooks.HookType
func (_e *MockSandbox_Expecter) Hook(hookType interface{}) *MockSandbox_Hook_Call {
	return &MockSandbox_Hook_Call{Call: _e.mock.On("Hook", hookType)}
}

func (_c *MockSandbox_Hook_Call) Run(run func(hookType hooks.HookType)) *MockSandbox_Hook_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(hooks.HookType))
	})
	return _c
}

func (_c *MockSandbox_Hook_Call) Return(_a0 hooks.Hook, _a1 error) *MockSandbox_Hook_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockSandbox_Hook_Call) RunAndReturn(run func(hooks.HookType) (hooks.Hook, error)) *MockSandbox_Hook_Call {
	_c.Call.Return(run)
	return _c
}

// Hooks provides a mock function with given fields:
func (_m *MockSandbox) Hooks() hooks.Hooks {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Hooks")
	}

	var r0 hooks.Hooks
	if rf, ok := ret.Get(0).(func() hooks.Hooks); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(hooks.Hooks)
		}
	}

	return r0
}

// MockSandbox_Hooks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Hooks'
type MockSandbox_Hooks_Call struct {
	*mock.Call
}

// Hooks is a helper method to define mock.On call
func (_e *MockSandbox_Expecter) Hooks() *MockSandbox_Hooks_Call {
	return &MockSandbox_Hooks_Call{Call: _e.mock.On("Hooks")}
}

func (_c *MockSandbox_Hooks_Call) Run(run func()) *MockSandbox_Hooks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockSandbox_Hooks_Call) Return(_a0 hooks.Hooks) *MockSandbox_Hooks_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockSandbox_Hooks_Call) RunAndReturn(run func() hooks.Hooks) *MockSandbox_Hooks_Call {
	_c.Call.Return(run)
	return _c
}

// Inherit provides a mock function with given fields: parentSandbox
func (_m *MockSandbox) Inherit(parentSandbox sandbox.Sandbox) error {
	ret := _m.Called(parentSandbox)

	if len(ret) == 0 {
		panic("no return value specified for Inherit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(sandbox.Sandbox) error); ok {
		r0 = rf(parentSandbox)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockSandbox_Inherit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Inherit'
type MockSandbox_Inherit_Call struct {
	*mock.Call
}

// Inherit is a helper method to define mock.On call
//   - parentSandbox sandbox.Sandbox
func (_e *MockSandbox_Expecter) Inherit(parentSandbox interface{}) *MockSandbox_Inherit_Call {
	return &MockSandbox_Inherit_Call{Call: _e.mock.On("Inherit", parentSandbox)}
}

func (_c *MockSandbox_Inherit_Call) Run(run func(parentSandbox sandbox.Sandbox)) *MockSandbox_Inherit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(sandbox.Sandbox))
	})
	return _c
}

func (_c *MockSandbox_Inherit_Call) Return(_a0 error) *MockSandbox_Inherit_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockSandbox_Inherit_Call) RunAndReturn(run func(sandbox.Sandbox) error) *MockSandbox_Inherit_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockSandbox creates a new instance of MockSandbox. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockSandbox(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSandbox {
	mock := &MockSandbox{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
