// Code generated by mockery v2.40.1. DO NOT EDIT.

package servers

import (
	actions "github.com/wstool/wst/run/servers/actions"
	configs "github.com/wstool/wst/run/servers/configs"

	mock "github.com/stretchr/testify/mock"

	parameters "github.com/wstool/wst/run/parameters"

	providers "github.com/wstool/wst/run/environments/environment/providers"

	sandbox "github.com/wstool/wst/run/sandboxes/sandbox"

	templates "github.com/wstool/wst/run/servers/templates"
)

// MockServer is an autogenerated mock type for the Server type
type MockServer struct {
	mock.Mock
}

type MockServer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockServer) EXPECT() *MockServer_Expecter {
	return &MockServer_Expecter{mock: &_m.Mock}
}

// Config provides a mock function with given fields: name
func (_m *MockServer) Config(name string) (configs.Config, bool) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Config")
	}

	var r0 configs.Config
	var r1 bool
	if rf, ok := ret.Get(0).(func(string) (configs.Config, bool)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) configs.Config); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(configs.Config)
		}
	}

	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// MockServer_Config_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Config'
type MockServer_Config_Call struct {
	*mock.Call
}

// Config is a helper method to define mock.On call
//   - name string
func (_e *MockServer_Expecter) Config(name interface{}) *MockServer_Config_Call {
	return &MockServer_Config_Call{Call: _e.mock.On("Config", name)}
}

func (_c *MockServer_Config_Call) Run(run func(name string)) *MockServer_Config_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockServer_Config_Call) Return(_a0 configs.Config, _a1 bool) *MockServer_Config_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServer_Config_Call) RunAndReturn(run func(string) (configs.Config, bool)) *MockServer_Config_Call {
	_c.Call.Return(run)
	return _c
}

// Configs provides a mock function with given fields:
func (_m *MockServer) Configs() configs.Configs {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Configs")
	}

	var r0 configs.Configs
	if rf, ok := ret.Get(0).(func() configs.Configs); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(configs.Configs)
		}
	}

	return r0
}

// MockServer_Configs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Configs'
type MockServer_Configs_Call struct {
	*mock.Call
}

// Configs is a helper method to define mock.On call
func (_e *MockServer_Expecter) Configs() *MockServer_Configs_Call {
	return &MockServer_Configs_Call{Call: _e.mock.On("Configs")}
}

func (_c *MockServer_Configs_Call) Run(run func()) *MockServer_Configs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockServer_Configs_Call) Return(_a0 configs.Configs) *MockServer_Configs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockServer_Configs_Call) RunAndReturn(run func() configs.Configs) *MockServer_Configs_Call {
	_c.Call.Return(run)
	return _c
}

// ExpectAction provides a mock function with given fields: name
func (_m *MockServer) ExpectAction(name string) (actions.ExpectAction, bool) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for ExpectAction")
	}

	var r0 actions.ExpectAction
	var r1 bool
	if rf, ok := ret.Get(0).(func(string) (actions.ExpectAction, bool)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) actions.ExpectAction); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(actions.ExpectAction)
		}
	}

	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// MockServer_ExpectAction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExpectAction'
type MockServer_ExpectAction_Call struct {
	*mock.Call
}

// ExpectAction is a helper method to define mock.On call
//   - name string
func (_e *MockServer_Expecter) ExpectAction(name interface{}) *MockServer_ExpectAction_Call {
	return &MockServer_ExpectAction_Call{Call: _e.mock.On("ExpectAction", name)}
}

func (_c *MockServer_ExpectAction_Call) Run(run func(name string)) *MockServer_ExpectAction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockServer_ExpectAction_Call) Return(_a0 actions.ExpectAction, _a1 bool) *MockServer_ExpectAction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServer_ExpectAction_Call) RunAndReturn(run func(string) (actions.ExpectAction, bool)) *MockServer_ExpectAction_Call {
	_c.Call.Return(run)
	return _c
}

// Group provides a mock function with given fields:
func (_m *MockServer) Group() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Group")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockServer_Group_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Group'
type MockServer_Group_Call struct {
	*mock.Call
}

// Group is a helper method to define mock.On call
func (_e *MockServer_Expecter) Group() *MockServer_Group_Call {
	return &MockServer_Group_Call{Call: _e.mock.On("Group")}
}

func (_c *MockServer_Group_Call) Run(run func()) *MockServer_Group_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockServer_Group_Call) Return(_a0 string) *MockServer_Group_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockServer_Group_Call) RunAndReturn(run func() string) *MockServer_Group_Call {
	_c.Call.Return(run)
	return _c
}

// Parameters provides a mock function with given fields:
func (_m *MockServer) Parameters() parameters.Parameters {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Parameters")
	}

	var r0 parameters.Parameters
	if rf, ok := ret.Get(0).(func() parameters.Parameters); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(parameters.Parameters)
		}
	}

	return r0
}

// MockServer_Parameters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Parameters'
type MockServer_Parameters_Call struct {
	*mock.Call
}

// Parameters is a helper method to define mock.On call
func (_e *MockServer_Expecter) Parameters() *MockServer_Parameters_Call {
	return &MockServer_Parameters_Call{Call: _e.mock.On("Parameters")}
}

func (_c *MockServer_Parameters_Call) Run(run func()) *MockServer_Parameters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockServer_Parameters_Call) Return(_a0 parameters.Parameters) *MockServer_Parameters_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockServer_Parameters_Call) RunAndReturn(run func() parameters.Parameters) *MockServer_Parameters_Call {
	_c.Call.Return(run)
	return _c
}

// Port provides a mock function with given fields:
func (_m *MockServer) Port() int32 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Port")
	}

	var r0 int32
	if rf, ok := ret.Get(0).(func() int32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int32)
	}

	return r0
}

// MockServer_Port_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Port'
type MockServer_Port_Call struct {
	*mock.Call
}

// Port is a helper method to define mock.On call
func (_e *MockServer_Expecter) Port() *MockServer_Port_Call {
	return &MockServer_Port_Call{Call: _e.mock.On("Port")}
}

func (_c *MockServer_Port_Call) Run(run func()) *MockServer_Port_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockServer_Port_Call) Return(_a0 int32) *MockServer_Port_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockServer_Port_Call) RunAndReturn(run func() int32) *MockServer_Port_Call {
	_c.Call.Return(run)
	return _c
}

// Sandbox provides a mock function with given fields: name
func (_m *MockServer) Sandbox(name providers.Type) (sandbox.Sandbox, bool) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Sandbox")
	}

	var r0 sandbox.Sandbox
	var r1 bool
	if rf, ok := ret.Get(0).(func(providers.Type) (sandbox.Sandbox, bool)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(providers.Type) sandbox.Sandbox); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sandbox.Sandbox)
		}
	}

	if rf, ok := ret.Get(1).(func(providers.Type) bool); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// MockServer_Sandbox_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Sandbox'
type MockServer_Sandbox_Call struct {
	*mock.Call
}

// Sandbox is a helper method to define mock.On call
//   - name providers.Type
func (_e *MockServer_Expecter) Sandbox(name interface{}) *MockServer_Sandbox_Call {
	return &MockServer_Sandbox_Call{Call: _e.mock.On("Sandbox", name)}
}

func (_c *MockServer_Sandbox_Call) Run(run func(name providers.Type)) *MockServer_Sandbox_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(providers.Type))
	})
	return _c
}

func (_c *MockServer_Sandbox_Call) Return(_a0 sandbox.Sandbox, _a1 bool) *MockServer_Sandbox_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServer_Sandbox_Call) RunAndReturn(run func(providers.Type) (sandbox.Sandbox, bool)) *MockServer_Sandbox_Call {
	_c.Call.Return(run)
	return _c
}

// SequentialAction provides a mock function with given fields: name
func (_m *MockServer) SequentialAction(name string) (actions.SequentialAction, bool) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for SequentialAction")
	}

	var r0 actions.SequentialAction
	var r1 bool
	if rf, ok := ret.Get(0).(func(string) (actions.SequentialAction, bool)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) actions.SequentialAction); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(actions.SequentialAction)
		}
	}

	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// MockServer_SequentialAction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SequentialAction'
type MockServer_SequentialAction_Call struct {
	*mock.Call
}

// SequentialAction is a helper method to define mock.On call
//   - name string
func (_e *MockServer_Expecter) SequentialAction(name interface{}) *MockServer_SequentialAction_Call {
	return &MockServer_SequentialAction_Call{Call: _e.mock.On("SequentialAction", name)}
}

func (_c *MockServer_SequentialAction_Call) Run(run func(name string)) *MockServer_SequentialAction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockServer_SequentialAction_Call) Return(_a0 actions.SequentialAction, _a1 bool) *MockServer_SequentialAction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServer_SequentialAction_Call) RunAndReturn(run func(string) (actions.SequentialAction, bool)) *MockServer_SequentialAction_Call {
	_c.Call.Return(run)
	return _c
}

// Template provides a mock function with given fields: name
func (_m *MockServer) Template(name string) (templates.Template, bool) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Template")
	}

	var r0 templates.Template
	var r1 bool
	if rf, ok := ret.Get(0).(func(string) (templates.Template, bool)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) templates.Template); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(templates.Template)
		}
	}

	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// MockServer_Template_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Template'
type MockServer_Template_Call struct {
	*mock.Call
}

// Template is a helper method to define mock.On call
//   - name string
func (_e *MockServer_Expecter) Template(name interface{}) *MockServer_Template_Call {
	return &MockServer_Template_Call{Call: _e.mock.On("Template", name)}
}

func (_c *MockServer_Template_Call) Run(run func(name string)) *MockServer_Template_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockServer_Template_Call) Return(_a0 templates.Template, _a1 bool) *MockServer_Template_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServer_Template_Call) RunAndReturn(run func(string) (templates.Template, bool)) *MockServer_Template_Call {
	_c.Call.Return(run)
	return _c
}

// Templates provides a mock function with given fields:
func (_m *MockServer) Templates() templates.Templates {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Templates")
	}

	var r0 templates.Templates
	if rf, ok := ret.Get(0).(func() templates.Templates); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(templates.Templates)
		}
	}

	return r0
}

// MockServer_Templates_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Templates'
type MockServer_Templates_Call struct {
	*mock.Call
}

// Templates is a helper method to define mock.On call
func (_e *MockServer_Expecter) Templates() *MockServer_Templates_Call {
	return &MockServer_Templates_Call{Call: _e.mock.On("Templates")}
}

func (_c *MockServer_Templates_Call) Run(run func()) *MockServer_Templates_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockServer_Templates_Call) Return(_a0 templates.Templates) *MockServer_Templates_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockServer_Templates_Call) RunAndReturn(run func() templates.Templates) *MockServer_Templates_Call {
	_c.Call.Return(run)
	return _c
}

// User provides a mock function with given fields:
func (_m *MockServer) User() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for User")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockServer_User_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'User'
type MockServer_User_Call struct {
	*mock.Call
}

// User is a helper method to define mock.On call
func (_e *MockServer_Expecter) User() *MockServer_User_Call {
	return &MockServer_User_Call{Call: _e.mock.On("User")}
}

func (_c *MockServer_User_Call) Run(run func()) *MockServer_User_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockServer_User_Call) Return(_a0 string) *MockServer_User_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockServer_User_Call) RunAndReturn(run func() string) *MockServer_User_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockServer creates a new instance of MockServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockServer {
	mock := &MockServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
