// Code generated by mockery v2.40.1. DO NOT EDIT.

package instances

import (
	instances "github.com/wstool/wst/run/instances"
	defaults "github.com/wstool/wst/run/spec/defaults"

	mock "github.com/stretchr/testify/mock"

	servers "github.com/wstool/wst/run/servers"

	types "github.com/wstool/wst/conf/types"
)

// MockInstanceMaker is an autogenerated mock type for the InstanceMaker type
type MockInstanceMaker struct {
	mock.Mock
}

type MockInstanceMaker_Expecter struct {
	mock *mock.Mock
}

func (_m *MockInstanceMaker) EXPECT() *MockInstanceMaker_Expecter {
	return &MockInstanceMaker_Expecter{mock: &_m.Mock}
}

// Make provides a mock function with given fields: instanceConfig, instanceId, envsConfig, dflts, srvs, specWorkspace
func (_m *MockInstanceMaker) Make(instanceConfig types.Instance, instanceId int, envsConfig map[string]types.Environment, dflts *defaults.Defaults, srvs servers.Servers, specWorkspace string) (instances.Instance, error) {
	ret := _m.Called(instanceConfig, instanceId, envsConfig, dflts, srvs, specWorkspace)

	if len(ret) == 0 {
		panic("no return value specified for Make")
	}

	var r0 instances.Instance
	var r1 error
	if rf, ok := ret.Get(0).(func(types.Instance, int, map[string]types.Environment, *defaults.Defaults, servers.Servers, string) (instances.Instance, error)); ok {
		return rf(instanceConfig, instanceId, envsConfig, dflts, srvs, specWorkspace)
	}
	if rf, ok := ret.Get(0).(func(types.Instance, int, map[string]types.Environment, *defaults.Defaults, servers.Servers, string) instances.Instance); ok {
		r0 = rf(instanceConfig, instanceId, envsConfig, dflts, srvs, specWorkspace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(instances.Instance)
		}
	}

	if rf, ok := ret.Get(1).(func(types.Instance, int, map[string]types.Environment, *defaults.Defaults, servers.Servers, string) error); ok {
		r1 = rf(instanceConfig, instanceId, envsConfig, dflts, srvs, specWorkspace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstanceMaker_Make_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Make'
type MockInstanceMaker_Make_Call struct {
	*mock.Call
}

// Make is a helper method to define mock.On call
//   - instanceConfig types.Instance
//   - instanceId int
//   - envsConfig map[string]types.Environment
//   - dflts *defaults.Defaults
//   - srvs servers.Servers
//   - specWorkspace string
func (_e *MockInstanceMaker_Expecter) Make(instanceConfig interface{}, instanceId interface{}, envsConfig interface{}, dflts interface{}, srvs interface{}, specWorkspace interface{}) *MockInstanceMaker_Make_Call {
	return &MockInstanceMaker_Make_Call{Call: _e.mock.On("Make", instanceConfig, instanceId, envsConfig, dflts, srvs, specWorkspace)}
}

func (_c *MockInstanceMaker_Make_Call) Run(run func(instanceConfig types.Instance, instanceId int, envsConfig map[string]types.Environment, dflts *defaults.Defaults, srvs servers.Servers, specWorkspace string)) *MockInstanceMaker_Make_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(types.Instance), args[1].(int), args[2].(map[string]types.Environment), args[3].(*defaults.Defaults), args[4].(servers.Servers), args[5].(string))
	})
	return _c
}

func (_c *MockInstanceMaker_Make_Call) Return(_a0 instances.Instance, _a1 error) *MockInstanceMaker_Make_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstanceMaker_Make_Call) RunAndReturn(run func(types.Instance, int, map[string]types.Environment, *defaults.Defaults, servers.Servers, string) (instances.Instance, error)) *MockInstanceMaker_Make_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockInstanceMaker creates a new instance of MockInstanceMaker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockInstanceMaker(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockInstanceMaker {
	mock := &MockInstanceMaker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
