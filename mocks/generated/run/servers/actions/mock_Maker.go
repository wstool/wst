// Code generated by mockery v2.40.1. DO NOT EDIT.

package actions

import (
	actions "github.com/bukka/wst/run/servers/actions"
	mock "github.com/stretchr/testify/mock"

	types "github.com/bukka/wst/conf/types"
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

// Make provides a mock function with given fields: configActions
func (_m *MockMaker) Make(configActions *types.ServerActions) (*actions.Actions, error) {
	ret := _m.Called(configActions)

	if len(ret) == 0 {
		panic("no return value specified for Make")
	}

	var r0 *actions.Actions
	var r1 error
	if rf, ok := ret.Get(0).(func(*types.ServerActions) (*actions.Actions, error)); ok {
		return rf(configActions)
	}
	if rf, ok := ret.Get(0).(func(*types.ServerActions) *actions.Actions); ok {
		r0 = rf(configActions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*actions.Actions)
		}
	}

	if rf, ok := ret.Get(1).(func(*types.ServerActions) error); ok {
		r1 = rf(configActions)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaker_Make_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Make'
type MockMaker_Make_Call struct {
	*mock.Call
}

// Make is a helper method to define mock.On call
//   - configActions *types.ServerActions
func (_e *MockMaker_Expecter) Make(configActions interface{}) *MockMaker_Make_Call {
	return &MockMaker_Make_Call{Call: _e.mock.On("Make", configActions)}
}

func (_c *MockMaker_Make_Call) Run(run func(configActions *types.ServerActions)) *MockMaker_Make_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*types.ServerActions))
	})
	return _c
}

func (_c *MockMaker_Make_Call) Return(_a0 *actions.Actions, _a1 error) *MockMaker_Make_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_Make_Call) RunAndReturn(run func(*types.ServerActions) (*actions.Actions, error)) *MockMaker_Make_Call {
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