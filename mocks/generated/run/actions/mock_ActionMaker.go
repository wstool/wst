// Code generated by mockery v2.40.1. DO NOT EDIT.

package actions

import (
	action "github.com/wstool/wst/run/actions/action"

	mock "github.com/stretchr/testify/mock"

	services "github.com/wstool/wst/run/services"

	types "github.com/wstool/wst/conf/types"
)

// MockActionMaker is an autogenerated mock type for the ActionMaker type
type MockActionMaker struct {
	mock.Mock
}

type MockActionMaker_Expecter struct {
	mock *mock.Mock
}

func (_m *MockActionMaker) EXPECT() *MockActionMaker_Expecter {
	return &MockActionMaker_Expecter{mock: &_m.Mock}
}

// MakeAction provides a mock function with given fields: config, sl, defaultTimeout
func (_m *MockActionMaker) MakeAction(config types.Action, sl services.ServiceLocator, defaultTimeout int) (action.Action, error) {
	ret := _m.Called(config, sl, defaultTimeout)

	if len(ret) == 0 {
		panic("no return value specified for MakeAction")
	}

	var r0 action.Action
	var r1 error
	if rf, ok := ret.Get(0).(func(types.Action, services.ServiceLocator, int) (action.Action, error)); ok {
		return rf(config, sl, defaultTimeout)
	}
	if rf, ok := ret.Get(0).(func(types.Action, services.ServiceLocator, int) action.Action); ok {
		r0 = rf(config, sl, defaultTimeout)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(action.Action)
		}
	}

	if rf, ok := ret.Get(1).(func(types.Action, services.ServiceLocator, int) error); ok {
		r1 = rf(config, sl, defaultTimeout)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockActionMaker_MakeAction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MakeAction'
type MockActionMaker_MakeAction_Call struct {
	*mock.Call
}

// MakeAction is a helper method to define mock.On call
//   - config types.Action
//   - sl services.ServiceLocator
//   - defaultTimeout int
func (_e *MockActionMaker_Expecter) MakeAction(config interface{}, sl interface{}, defaultTimeout interface{}) *MockActionMaker_MakeAction_Call {
	return &MockActionMaker_MakeAction_Call{Call: _e.mock.On("MakeAction", config, sl, defaultTimeout)}
}

func (_c *MockActionMaker_MakeAction_Call) Run(run func(config types.Action, sl services.ServiceLocator, defaultTimeout int)) *MockActionMaker_MakeAction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(types.Action), args[1].(services.ServiceLocator), args[2].(int))
	})
	return _c
}

func (_c *MockActionMaker_MakeAction_Call) Return(_a0 action.Action, _a1 error) *MockActionMaker_MakeAction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockActionMaker_MakeAction_Call) RunAndReturn(run func(types.Action, services.ServiceLocator, int) (action.Action, error)) *MockActionMaker_MakeAction_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockActionMaker creates a new instance of MockActionMaker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockActionMaker(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockActionMaker {
	mock := &MockActionMaker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
