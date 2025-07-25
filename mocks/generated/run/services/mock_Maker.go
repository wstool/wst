// Code generated by mockery v2.40.1. DO NOT EDIT.

package services

import (
	environments "github.com/wstool/wst/run/environments"
	defaults "github.com/wstool/wst/run/spec/defaults"

	mock "github.com/stretchr/testify/mock"

	parameters "github.com/wstool/wst/run/parameters"

	resources "github.com/wstool/wst/run/resources"

	servers "github.com/wstool/wst/run/servers"

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

// Make provides a mock function with given fields: config, dflts, rscrs, srvs, _a4, instanceName, instanceIdx, instanceWorkspace, instanceParameters
func (_m *MockMaker) Make(config map[string]types.Service, dflts *defaults.Defaults, rscrs *resources.Resources, srvs servers.Servers, _a4 environments.Environments, instanceName string, instanceIdx int, instanceWorkspace string, instanceParameters parameters.Parameters) (services.ServiceLocator, error) {
	ret := _m.Called(config, dflts, rscrs, srvs, _a4, instanceName, instanceIdx, instanceWorkspace, instanceParameters)

	if len(ret) == 0 {
		panic("no return value specified for Make")
	}

	var r0 services.ServiceLocator
	var r1 error
	if rf, ok := ret.Get(0).(func(map[string]types.Service, *defaults.Defaults, *resources.Resources, servers.Servers, environments.Environments, string, int, string, parameters.Parameters) (services.ServiceLocator, error)); ok {
		return rf(config, dflts, rscrs, srvs, _a4, instanceName, instanceIdx, instanceWorkspace, instanceParameters)
	}
	if rf, ok := ret.Get(0).(func(map[string]types.Service, *defaults.Defaults, *resources.Resources, servers.Servers, environments.Environments, string, int, string, parameters.Parameters) services.ServiceLocator); ok {
		r0 = rf(config, dflts, rscrs, srvs, _a4, instanceName, instanceIdx, instanceWorkspace, instanceParameters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(services.ServiceLocator)
		}
	}

	if rf, ok := ret.Get(1).(func(map[string]types.Service, *defaults.Defaults, *resources.Resources, servers.Servers, environments.Environments, string, int, string, parameters.Parameters) error); ok {
		r1 = rf(config, dflts, rscrs, srvs, _a4, instanceName, instanceIdx, instanceWorkspace, instanceParameters)
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
//   - config map[string]types.Service
//   - dflts *defaults.Defaults
//   - rscrs *resources.Resources
//   - srvs servers.Servers
//   - _a4 environments.Environments
//   - instanceName string
//   - instanceIdx int
//   - instanceWorkspace string
//   - instanceParameters parameters.Parameters
func (_e *MockMaker_Expecter) Make(config interface{}, dflts interface{}, rscrs interface{}, srvs interface{}, _a4 interface{}, instanceName interface{}, instanceIdx interface{}, instanceWorkspace interface{}, instanceParameters interface{}) *MockMaker_Make_Call {
	return &MockMaker_Make_Call{Call: _e.mock.On("Make", config, dflts, rscrs, srvs, _a4, instanceName, instanceIdx, instanceWorkspace, instanceParameters)}
}

func (_c *MockMaker_Make_Call) Run(run func(config map[string]types.Service, dflts *defaults.Defaults, rscrs *resources.Resources, srvs servers.Servers, _a4 environments.Environments, instanceName string, instanceIdx int, instanceWorkspace string, instanceParameters parameters.Parameters)) *MockMaker_Make_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]types.Service), args[1].(*defaults.Defaults), args[2].(*resources.Resources), args[3].(servers.Servers), args[4].(environments.Environments), args[5].(string), args[6].(int), args[7].(string), args[8].(parameters.Parameters))
	})
	return _c
}

func (_c *MockMaker_Make_Call) Return(_a0 services.ServiceLocator, _a1 error) *MockMaker_Make_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_Make_Call) RunAndReturn(run func(map[string]types.Service, *defaults.Defaults, *resources.Resources, servers.Servers, environments.Environments, string, int, string, parameters.Parameters) (services.ServiceLocator, error)) *MockMaker_Make_Call {
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
