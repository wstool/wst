// Code generated by mockery v2.40.1. DO NOT EDIT.

package app

import (
	time "time"

	mock "github.com/stretchr/testify/mock"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// MockVegetaAttacker is an autogenerated mock type for the VegetaAttacker type
type MockVegetaAttacker struct {
	mock.Mock
}

type MockVegetaAttacker_Expecter struct {
	mock *mock.Mock
}

func (_m *MockVegetaAttacker) EXPECT() *MockVegetaAttacker_Expecter {
	return &MockVegetaAttacker_Expecter{mock: &_m.Mock}
}

// Attack provides a mock function with given fields: targeter, rate, duration, name
func (_m *MockVegetaAttacker) Attack(targeter vegeta.Targeter, rate vegeta.ConstantPacer, duration time.Duration, name string) <-chan *vegeta.Result {
	ret := _m.Called(targeter, rate, duration, name)

	if len(ret) == 0 {
		panic("no return value specified for Attack")
	}

	var r0 <-chan *vegeta.Result
	if rf, ok := ret.Get(0).(func(vegeta.Targeter, vegeta.ConstantPacer, time.Duration, string) <-chan *vegeta.Result); ok {
		r0 = rf(targeter, rate, duration, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *vegeta.Result)
		}
	}

	return r0
}

// MockVegetaAttacker_Attack_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Attack'
type MockVegetaAttacker_Attack_Call struct {
	*mock.Call
}

// Attack is a helper method to define mock.On call
//   - targeter vegeta.Targeter
//   - rate vegeta.ConstantPacer
//   - duration time.Duration
//   - name string
func (_e *MockVegetaAttacker_Expecter) Attack(targeter interface{}, rate interface{}, duration interface{}, name interface{}) *MockVegetaAttacker_Attack_Call {
	return &MockVegetaAttacker_Attack_Call{Call: _e.mock.On("Attack", targeter, rate, duration, name)}
}

func (_c *MockVegetaAttacker_Attack_Call) Run(run func(targeter vegeta.Targeter, rate vegeta.ConstantPacer, duration time.Duration, name string)) *MockVegetaAttacker_Attack_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(vegeta.Targeter), args[1].(vegeta.ConstantPacer), args[2].(time.Duration), args[3].(string))
	})
	return _c
}

func (_c *MockVegetaAttacker_Attack_Call) Return(_a0 <-chan *vegeta.Result) *MockVegetaAttacker_Attack_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockVegetaAttacker_Attack_Call) RunAndReturn(run func(vegeta.Targeter, vegeta.ConstantPacer, time.Duration, string) <-chan *vegeta.Result) *MockVegetaAttacker_Attack_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockVegetaAttacker creates a new instance of MockVegetaAttacker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockVegetaAttacker(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockVegetaAttacker {
	mock := &MockVegetaAttacker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
