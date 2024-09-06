// Code generated by mockery v2.40.1. DO NOT EDIT.

package stop

import (
	mock "github.com/stretchr/testify/mock"
	action "github.com/wstool/wst/run/actions/action"

	services "github.com/wstool/wst/run/services"

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

// Make provides a mock function with given fields: config, sl, defaultTimeout
func (_m *MockMaker) Make(config *types.StopAction, sl services.ServiceLocator, defaultTimeout int) (action.Action, error) {
	ret := _m.Called(config, sl, defaultTimeout)

	if len(ret) == 0 {
		panic("no return value specified for Make")
	}

	var r0 action.Action
	var r1 error
	if rf, ok := ret.Get(0).(func(*types.StopAction, services.ServiceLocator, int) (action.Action, error)); ok {
		return rf(config, sl, defaultTimeout)
	}
	if rf, ok := ret.Get(0).(func(*types.StopAction, services.ServiceLocator, int) action.Action); ok {
		r0 = rf(config, sl, defaultTimeout)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(action.Action)
		}
	}

	if rf, ok := ret.Get(1).(func(*types.StopAction, services.ServiceLocator, int) error); ok {
		r1 = rf(config, sl, defaultTimeout)
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
//   - config *types.StopAction
//   - sl services.ServiceLocator
//   - defaultTimeout int
func (_e *MockMaker_Expecter) Make(config interface{}, sl interface{}, defaultTimeout interface{}) *MockMaker_Make_Call {
	return &MockMaker_Make_Call{Call: _e.mock.On("Make", config, sl, defaultTimeout)}
}

func (_c *MockMaker_Make_Call) Run(run func(config *types.StopAction, sl services.ServiceLocator, defaultTimeout int)) *MockMaker_Make_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*types.StopAction), args[1].(services.ServiceLocator), args[2].(int))
	})
	return _c
}

func (_c *MockMaker_Make_Call) Return(_a0 action.Action, _a1 error) *MockMaker_Make_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_Make_Call) RunAndReturn(run func(*types.StopAction, services.ServiceLocator, int) (action.Action, error)) *MockMaker_Make_Call {
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
