// Code generated by mockery v2.40.1. DO NOT EDIT.

package container

import (
	mock "github.com/stretchr/testify/mock"
	container "github.com/wstool/wst/run/sandboxes/sandbox/container"

	types "github.com/wstool/wst/conf/types"
)

// MockMaker is an autogenerated mock type for the Maker type
type MockMaker struct {
	mock.Mock
}

type MockMaker_Expecter struct {
	mock *mock.Mock
}

func (_m *MockMaker) EXPECT() *MockMaker_Expecter {
	return &MockMaker_Expecter{mock: &_m.Mock}
}

// MakeSandbox provides a mock function with given fields: config
func (_m *MockMaker) MakeSandbox(config *types.ContainerSandbox) (*container.Sandbox, error) {
	ret := _m.Called(config)

	if len(ret) == 0 {
		panic("no return value specified for MakeSandbox")
	}

	var r0 *container.Sandbox
	var r1 error
	if rf, ok := ret.Get(0).(func(*types.ContainerSandbox) (*container.Sandbox, error)); ok {
		return rf(config)
	}
	if rf, ok := ret.Get(0).(func(*types.ContainerSandbox) *container.Sandbox); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*container.Sandbox)
		}
	}

	if rf, ok := ret.Get(1).(func(*types.ContainerSandbox) error); ok {
		r1 = rf(config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaker_MakeSandbox_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MakeSandbox'
type MockMaker_MakeSandbox_Call struct {
	*mock.Call
}

// MakeSandbox is a helper method to define mock.On call
//   - config *types.ContainerSandbox
func (_e *MockMaker_Expecter) MakeSandbox(config interface{}) *MockMaker_MakeSandbox_Call {
	return &MockMaker_MakeSandbox_Call{Call: _e.mock.On("MakeSandbox", config)}
}

func (_c *MockMaker_MakeSandbox_Call) Run(run func(config *types.ContainerSandbox)) *MockMaker_MakeSandbox_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*types.ContainerSandbox))
	})
	return _c
}

func (_c *MockMaker_MakeSandbox_Call) Return(_a0 *container.Sandbox, _a1 error) *MockMaker_MakeSandbox_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_MakeSandbox_Call) RunAndReturn(run func(*types.ContainerSandbox) (*container.Sandbox, error)) *MockMaker_MakeSandbox_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockMaker creates a new instance of MockMaker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMaker(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMaker {
	mock := &MockMaker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
