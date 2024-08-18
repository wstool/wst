// Code generated by mockery v2.40.1. DO NOT EDIT.

package task

import (
	providers "github.com/bukka/wst/run/environments/environment/providers"
	mock "github.com/stretchr/testify/mock"
)

// MockTask is an autogenerated mock type for the Task type
type MockTask struct {
	mock.Mock
}

type MockTask_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTask) EXPECT() *MockTask_Expecter {
	return &MockTask_Expecter{mock: &_m.Mock}
}

// Executable provides a mock function with given fields:
func (_m *MockTask) Executable() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Executable")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTask_Executable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Executable'
type MockTask_Executable_Call struct {
	*mock.Call
}

// Executable is a helper method to define mock.On call
func (_e *MockTask_Expecter) Executable() *MockTask_Executable_Call {
	return &MockTask_Executable_Call{Call: _e.mock.On("Executable")}
}

func (_c *MockTask_Executable_Call) Run(run func()) *MockTask_Executable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTask_Executable_Call) Return(_a0 string) *MockTask_Executable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTask_Executable_Call) RunAndReturn(run func() string) *MockTask_Executable_Call {
	_c.Call.Return(run)
	return _c
}

// Id provides a mock function with given fields:
func (_m *MockTask) Id() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Id")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTask_Id_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Id'
type MockTask_Id_Call struct {
	*mock.Call
}

// Id is a helper method to define mock.On call
func (_e *MockTask_Expecter) Id() *MockTask_Id_Call {
	return &MockTask_Id_Call{Call: _e.mock.On("Id")}
}

func (_c *MockTask_Id_Call) Run(run func()) *MockTask_Id_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTask_Id_Call) Return(_a0 string) *MockTask_Id_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTask_Id_Call) RunAndReturn(run func() string) *MockTask_Id_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *MockTask) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTask_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockTask_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockTask_Expecter) Name() *MockTask_Name_Call {
	return &MockTask_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockTask_Name_Call) Run(run func()) *MockTask_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTask_Name_Call) Return(_a0 string) *MockTask_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTask_Name_Call) RunAndReturn(run func() string) *MockTask_Name_Call {
	_c.Call.Return(run)
	return _c
}

// Pid provides a mock function with given fields:
func (_m *MockTask) Pid() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Pid")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// MockTask_Pid_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Pid'
type MockTask_Pid_Call struct {
	*mock.Call
}

// Pid is a helper method to define mock.On call
func (_e *MockTask_Expecter) Pid() *MockTask_Pid_Call {
	return &MockTask_Pid_Call{Call: _e.mock.On("Pid")}
}

func (_c *MockTask_Pid_Call) Run(run func()) *MockTask_Pid_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTask_Pid_Call) Return(_a0 int) *MockTask_Pid_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTask_Pid_Call) RunAndReturn(run func() int) *MockTask_Pid_Call {
	_c.Call.Return(run)
	return _c
}

// PrivateUrl provides a mock function with given fields:
func (_m *MockTask) PrivateUrl() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for PrivateUrl")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTask_PrivateUrl_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PrivateUrl'
type MockTask_PrivateUrl_Call struct {
	*mock.Call
}

// PrivateUrl is a helper method to define mock.On call
func (_e *MockTask_Expecter) PrivateUrl() *MockTask_PrivateUrl_Call {
	return &MockTask_PrivateUrl_Call{Call: _e.mock.On("PrivateUrl")}
}

func (_c *MockTask_PrivateUrl_Call) Run(run func()) *MockTask_PrivateUrl_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTask_PrivateUrl_Call) Return(_a0 string) *MockTask_PrivateUrl_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTask_PrivateUrl_Call) RunAndReturn(run func() string) *MockTask_PrivateUrl_Call {
	_c.Call.Return(run)
	return _c
}

// PublicUrl provides a mock function with given fields:
func (_m *MockTask) PublicUrl() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for PublicUrl")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTask_PublicUrl_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PublicUrl'
type MockTask_PublicUrl_Call struct {
	*mock.Call
}

// PublicUrl is a helper method to define mock.On call
func (_e *MockTask_Expecter) PublicUrl() *MockTask_PublicUrl_Call {
	return &MockTask_PublicUrl_Call{Call: _e.mock.On("PublicUrl")}
}

func (_c *MockTask_PublicUrl_Call) Run(run func()) *MockTask_PublicUrl_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTask_PublicUrl_Call) Return(_a0 string) *MockTask_PublicUrl_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTask_PublicUrl_Call) RunAndReturn(run func() string) *MockTask_PublicUrl_Call {
	_c.Call.Return(run)
	return _c
}

// Type provides a mock function with given fields:
func (_m *MockTask) Type() providers.Type {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Type")
	}

	var r0 providers.Type
	if rf, ok := ret.Get(0).(func() providers.Type); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(providers.Type)
	}

	return r0
}

// MockTask_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type MockTask_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *MockTask_Expecter) Type() *MockTask_Type_Call {
	return &MockTask_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *MockTask_Type_Call) Run(run func()) *MockTask_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTask_Type_Call) Return(_a0 providers.Type) *MockTask_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTask_Type_Call) RunAndReturn(run func() providers.Type) *MockTask_Type_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockTask creates a new instance of MockTask. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTask(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTask {
	mock := &MockTask{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
