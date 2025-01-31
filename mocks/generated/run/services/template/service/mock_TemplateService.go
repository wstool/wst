// Code generated by mockery v2.40.1. DO NOT EDIT.

package service

import mock "github.com/stretchr/testify/mock"

// MockTemplateService is an autogenerated mock type for the TemplateService type
type MockTemplateService struct {
	mock.Mock
}

type MockTemplateService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTemplateService) EXPECT() *MockTemplateService_Expecter {
	return &MockTemplateService_Expecter{mock: &_m.Mock}
}

// ConfDir provides a mock function with given fields:
func (_m *MockTemplateService) ConfDir() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ConfDir")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTemplateService_ConfDir_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConfDir'
type MockTemplateService_ConfDir_Call struct {
	*mock.Call
}

// ConfDir is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) ConfDir() *MockTemplateService_ConfDir_Call {
	return &MockTemplateService_ConfDir_Call{Call: _e.mock.On("ConfDir")}
}

func (_c *MockTemplateService_ConfDir_Call) Run(run func()) *MockTemplateService_ConfDir_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_ConfDir_Call) Return(_a0 string, _a1 error) *MockTemplateService_ConfDir_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTemplateService_ConfDir_Call) RunAndReturn(run func() (string, error)) *MockTemplateService_ConfDir_Call {
	_c.Call.Return(run)
	return _c
}

// EnvironmentConfigPaths provides a mock function with given fields:
func (_m *MockTemplateService) EnvironmentConfigPaths() map[string]string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EnvironmentConfigPaths")
	}

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	return r0
}

// MockTemplateService_EnvironmentConfigPaths_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EnvironmentConfigPaths'
type MockTemplateService_EnvironmentConfigPaths_Call struct {
	*mock.Call
}

// EnvironmentConfigPaths is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) EnvironmentConfigPaths() *MockTemplateService_EnvironmentConfigPaths_Call {
	return &MockTemplateService_EnvironmentConfigPaths_Call{Call: _e.mock.On("EnvironmentConfigPaths")}
}

func (_c *MockTemplateService_EnvironmentConfigPaths_Call) Run(run func()) *MockTemplateService_EnvironmentConfigPaths_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_EnvironmentConfigPaths_Call) Return(_a0 map[string]string) *MockTemplateService_EnvironmentConfigPaths_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTemplateService_EnvironmentConfigPaths_Call) RunAndReturn(run func() map[string]string) *MockTemplateService_EnvironmentConfigPaths_Call {
	_c.Call.Return(run)
	return _c
}

// Executable provides a mock function with given fields:
func (_m *MockTemplateService) Executable() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Executable")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTemplateService_Executable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Executable'
type MockTemplateService_Executable_Call struct {
	*mock.Call
}

// Executable is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) Executable() *MockTemplateService_Executable_Call {
	return &MockTemplateService_Executable_Call{Call: _e.mock.On("Executable")}
}

func (_c *MockTemplateService_Executable_Call) Run(run func()) *MockTemplateService_Executable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_Executable_Call) Return(_a0 string, _a1 error) *MockTemplateService_Executable_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTemplateService_Executable_Call) RunAndReturn(run func() (string, error)) *MockTemplateService_Executable_Call {
	_c.Call.Return(run)
	return _c
}

// Group provides a mock function with given fields:
func (_m *MockTemplateService) Group() string {
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

// MockTemplateService_Group_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Group'
type MockTemplateService_Group_Call struct {
	*mock.Call
}

// Group is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) Group() *MockTemplateService_Group_Call {
	return &MockTemplateService_Group_Call{Call: _e.mock.On("Group")}
}

func (_c *MockTemplateService_Group_Call) Run(run func()) *MockTemplateService_Group_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_Group_Call) Return(_a0 string) *MockTemplateService_Group_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTemplateService_Group_Call) RunAndReturn(run func() string) *MockTemplateService_Group_Call {
	_c.Call.Return(run)
	return _c
}

// LocalAddress provides a mock function with given fields:
func (_m *MockTemplateService) LocalAddress() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for LocalAddress")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTemplateService_LocalAddress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LocalAddress'
type MockTemplateService_LocalAddress_Call struct {
	*mock.Call
}

// LocalAddress is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) LocalAddress() *MockTemplateService_LocalAddress_Call {
	return &MockTemplateService_LocalAddress_Call{Call: _e.mock.On("LocalAddress")}
}

func (_c *MockTemplateService_LocalAddress_Call) Run(run func()) *MockTemplateService_LocalAddress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_LocalAddress_Call) Return(_a0 string) *MockTemplateService_LocalAddress_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTemplateService_LocalAddress_Call) RunAndReturn(run func() string) *MockTemplateService_LocalAddress_Call {
	_c.Call.Return(run)
	return _c
}

// LocalPort provides a mock function with given fields:
func (_m *MockTemplateService) LocalPort() int32 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for LocalPort")
	}

	var r0 int32
	if rf, ok := ret.Get(0).(func() int32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int32)
	}

	return r0
}

// MockTemplateService_LocalPort_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LocalPort'
type MockTemplateService_LocalPort_Call struct {
	*mock.Call
}

// LocalPort is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) LocalPort() *MockTemplateService_LocalPort_Call {
	return &MockTemplateService_LocalPort_Call{Call: _e.mock.On("LocalPort")}
}

func (_c *MockTemplateService_LocalPort_Call) Run(run func()) *MockTemplateService_LocalPort_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_LocalPort_Call) Return(_a0 int32) *MockTemplateService_LocalPort_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTemplateService_LocalPort_Call) RunAndReturn(run func() int32) *MockTemplateService_LocalPort_Call {
	_c.Call.Return(run)
	return _c
}

// Pid provides a mock function with given fields:
func (_m *MockTemplateService) Pid() (int, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Pid")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func() (int, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTemplateService_Pid_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Pid'
type MockTemplateService_Pid_Call struct {
	*mock.Call
}

// Pid is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) Pid() *MockTemplateService_Pid_Call {
	return &MockTemplateService_Pid_Call{Call: _e.mock.On("Pid")}
}

func (_c *MockTemplateService_Pid_Call) Run(run func()) *MockTemplateService_Pid_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_Pid_Call) Return(_a0 int, _a1 error) *MockTemplateService_Pid_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTemplateService_Pid_Call) RunAndReturn(run func() (int, error)) *MockTemplateService_Pid_Call {
	_c.Call.Return(run)
	return _c
}

// PrivateAddress provides a mock function with given fields:
func (_m *MockTemplateService) PrivateAddress() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for PrivateAddress")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTemplateService_PrivateAddress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PrivateAddress'
type MockTemplateService_PrivateAddress_Call struct {
	*mock.Call
}

// PrivateAddress is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) PrivateAddress() *MockTemplateService_PrivateAddress_Call {
	return &MockTemplateService_PrivateAddress_Call{Call: _e.mock.On("PrivateAddress")}
}

func (_c *MockTemplateService_PrivateAddress_Call) Run(run func()) *MockTemplateService_PrivateAddress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_PrivateAddress_Call) Return(_a0 string) *MockTemplateService_PrivateAddress_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTemplateService_PrivateAddress_Call) RunAndReturn(run func() string) *MockTemplateService_PrivateAddress_Call {
	_c.Call.Return(run)
	return _c
}

// PrivateUrl provides a mock function with given fields:
func (_m *MockTemplateService) PrivateUrl() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for PrivateUrl")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTemplateService_PrivateUrl_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PrivateUrl'
type MockTemplateService_PrivateUrl_Call struct {
	*mock.Call
}

// PrivateUrl is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) PrivateUrl() *MockTemplateService_PrivateUrl_Call {
	return &MockTemplateService_PrivateUrl_Call{Call: _e.mock.On("PrivateUrl")}
}

func (_c *MockTemplateService_PrivateUrl_Call) Run(run func()) *MockTemplateService_PrivateUrl_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_PrivateUrl_Call) Return(_a0 string, _a1 error) *MockTemplateService_PrivateUrl_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTemplateService_PrivateUrl_Call) RunAndReturn(run func() (string, error)) *MockTemplateService_PrivateUrl_Call {
	_c.Call.Return(run)
	return _c
}

// RunDir provides a mock function with given fields:
func (_m *MockTemplateService) RunDir() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RunDir")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTemplateService_RunDir_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RunDir'
type MockTemplateService_RunDir_Call struct {
	*mock.Call
}

// RunDir is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) RunDir() *MockTemplateService_RunDir_Call {
	return &MockTemplateService_RunDir_Call{Call: _e.mock.On("RunDir")}
}

func (_c *MockTemplateService_RunDir_Call) Run(run func()) *MockTemplateService_RunDir_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_RunDir_Call) Return(_a0 string, _a1 error) *MockTemplateService_RunDir_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTemplateService_RunDir_Call) RunAndReturn(run func() (string, error)) *MockTemplateService_RunDir_Call {
	_c.Call.Return(run)
	return _c
}

// ScriptDir provides a mock function with given fields:
func (_m *MockTemplateService) ScriptDir() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ScriptDir")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTemplateService_ScriptDir_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ScriptDir'
type MockTemplateService_ScriptDir_Call struct {
	*mock.Call
}

// ScriptDir is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) ScriptDir() *MockTemplateService_ScriptDir_Call {
	return &MockTemplateService_ScriptDir_Call{Call: _e.mock.On("ScriptDir")}
}

func (_c *MockTemplateService_ScriptDir_Call) Run(run func()) *MockTemplateService_ScriptDir_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_ScriptDir_Call) Return(_a0 string, _a1 error) *MockTemplateService_ScriptDir_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTemplateService_ScriptDir_Call) RunAndReturn(run func() (string, error)) *MockTemplateService_ScriptDir_Call {
	_c.Call.Return(run)
	return _c
}

// UdsPath provides a mock function with given fields: _a0
func (_m *MockTemplateService) UdsPath(_a0 ...string) (string, error) {
	_va := make([]interface{}, len(_a0))
	for _i := range _a0 {
		_va[_i] = _a0[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UdsPath")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(...string) (string, error)); ok {
		return rf(_a0...)
	}
	if rf, ok := ret.Get(0).(func(...string) string); ok {
		r0 = rf(_a0...)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(...string) error); ok {
		r1 = rf(_a0...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTemplateService_UdsPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UdsPath'
type MockTemplateService_UdsPath_Call struct {
	*mock.Call
}

// UdsPath is a helper method to define mock.On call
//   - _a0 ...string
func (_e *MockTemplateService_Expecter) UdsPath(_a0 ...interface{}) *MockTemplateService_UdsPath_Call {
	return &MockTemplateService_UdsPath_Call{Call: _e.mock.On("UdsPath",
		append([]interface{}{}, _a0...)...)}
}

func (_c *MockTemplateService_UdsPath_Call) Run(run func(_a0 ...string)) *MockTemplateService_UdsPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *MockTemplateService_UdsPath_Call) Return(_a0 string, _a1 error) *MockTemplateService_UdsPath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTemplateService_UdsPath_Call) RunAndReturn(run func(...string) (string, error)) *MockTemplateService_UdsPath_Call {
	_c.Call.Return(run)
	return _c
}

// User provides a mock function with given fields:
func (_m *MockTemplateService) User() string {
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

// MockTemplateService_User_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'User'
type MockTemplateService_User_Call struct {
	*mock.Call
}

// User is a helper method to define mock.On call
func (_e *MockTemplateService_Expecter) User() *MockTemplateService_User_Call {
	return &MockTemplateService_User_Call{Call: _e.mock.On("User")}
}

func (_c *MockTemplateService_User_Call) Run(run func()) *MockTemplateService_User_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTemplateService_User_Call) Return(_a0 string) *MockTemplateService_User_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTemplateService_User_Call) RunAndReturn(run func() string) *MockTemplateService_User_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockTemplateService creates a new instance of MockTemplateService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTemplateService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTemplateService {
	mock := &MockTemplateService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
