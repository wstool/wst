// Code generated by mockery v2.40.1. DO NOT EDIT.

package instances

import mock "github.com/stretchr/testify/mock"

// MockInstance is an autogenerated mock type for the Instance type
type MockInstance struct {
	mock.Mock
}

type MockInstance_Expecter struct {
	mock *mock.Mock
}

func (_m *MockInstance) EXPECT() *MockInstance_Expecter {
	return &MockInstance_Expecter{mock: &_m.Mock}
}

// Name provides a mock function with given fields:
func (_m *MockInstance) Name() string {
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

// MockInstance_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockInstance_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockInstance_Expecter) Name() *MockInstance_Name_Call {
	return &MockInstance_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockInstance_Name_Call) Run(run func()) *MockInstance_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockInstance_Name_Call) Return(_a0 string) *MockInstance_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstance_Name_Call) RunAndReturn(run func() string) *MockInstance_Name_Call {
	_c.Call.Return(run)
	return _c
}

// Run provides a mock function with given fields:
func (_m *MockInstance) Run() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Run")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstance_Run_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Run'
type MockInstance_Run_Call struct {
	*mock.Call
}

// Run is a helper method to define mock.On call
func (_e *MockInstance_Expecter) Run() *MockInstance_Run_Call {
	return &MockInstance_Run_Call{Call: _e.mock.On("Run")}
}

func (_c *MockInstance_Run_Call) Run(run func()) *MockInstance_Run_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockInstance_Run_Call) Return(_a0 error) *MockInstance_Run_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstance_Run_Call) RunAndReturn(run func() error) *MockInstance_Run_Call {
	_c.Call.Return(run)
	return _c
}

// Workspace provides a mock function with given fields:
func (_m *MockInstance) Workspace() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Workspace")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockInstance_Workspace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Workspace'
type MockInstance_Workspace_Call struct {
	*mock.Call
}

// Workspace is a helper method to define mock.On call
func (_e *MockInstance_Expecter) Workspace() *MockInstance_Workspace_Call {
	return &MockInstance_Workspace_Call{Call: _e.mock.On("Workspace")}
}

func (_c *MockInstance_Workspace_Call) Run(run func()) *MockInstance_Workspace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockInstance_Workspace_Call) Return(_a0 string) *MockInstance_Workspace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstance_Workspace_Call) RunAndReturn(run func() string) *MockInstance_Workspace_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockInstance creates a new instance of MockInstance. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockInstance(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockInstance {
	mock := &MockInstance{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}