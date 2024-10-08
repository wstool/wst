// Code generated by mockery v2.40.1. DO NOT EDIT.

package sandboxes

import (
	mock "github.com/stretchr/testify/mock"
	sandboxes "github.com/wstool/wst/run/sandboxes"

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

// MakeSandboxes provides a mock function with given fields: rootSandboxes, serverSandboxes
func (_m *MockMaker) MakeSandboxes(rootSandboxes map[string]types.Sandbox, serverSandboxes map[string]types.Sandbox) (sandboxes.Sandboxes, error) {
	ret := _m.Called(rootSandboxes, serverSandboxes)

	if len(ret) == 0 {
		panic("no return value specified for MakeSandboxes")
	}

	var r0 sandboxes.Sandboxes
	var r1 error
	if rf, ok := ret.Get(0).(func(map[string]types.Sandbox, map[string]types.Sandbox) (sandboxes.Sandboxes, error)); ok {
		return rf(rootSandboxes, serverSandboxes)
	}
	if rf, ok := ret.Get(0).(func(map[string]types.Sandbox, map[string]types.Sandbox) sandboxes.Sandboxes); ok {
		r0 = rf(rootSandboxes, serverSandboxes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sandboxes.Sandboxes)
		}
	}

	if rf, ok := ret.Get(1).(func(map[string]types.Sandbox, map[string]types.Sandbox) error); ok {
		r1 = rf(rootSandboxes, serverSandboxes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaker_MakeSandboxes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MakeSandboxes'
type MockMaker_MakeSandboxes_Call struct {
	*mock.Call
}

// MakeSandboxes is a helper method to define mock.On call
//   - rootSandboxes map[string]types.Sandbox
//   - serverSandboxes map[string]types.Sandbox
func (_e *MockMaker_Expecter) MakeSandboxes(rootSandboxes interface{}, serverSandboxes interface{}) *MockMaker_MakeSandboxes_Call {
	return &MockMaker_MakeSandboxes_Call{Call: _e.mock.On("MakeSandboxes", rootSandboxes, serverSandboxes)}
}

func (_c *MockMaker_MakeSandboxes_Call) Run(run func(rootSandboxes map[string]types.Sandbox, serverSandboxes map[string]types.Sandbox)) *MockMaker_MakeSandboxes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]types.Sandbox), args[1].(map[string]types.Sandbox))
	})
	return _c
}

func (_c *MockMaker_MakeSandboxes_Call) Return(_a0 sandboxes.Sandboxes, _a1 error) *MockMaker_MakeSandboxes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_MakeSandboxes_Call) RunAndReturn(run func(map[string]types.Sandbox, map[string]types.Sandbox) (sandboxes.Sandboxes, error)) *MockMaker_MakeSandboxes_Call {
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
