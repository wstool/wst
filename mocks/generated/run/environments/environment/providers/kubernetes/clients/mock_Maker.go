// Code generated by mockery v2.40.1. DO NOT EDIT.

package clients

import (
	mock "github.com/stretchr/testify/mock"
	clients "github.com/wstool/wst/run/environments/environment/providers/kubernetes/clients"

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

// MakeConfigMapClient provides a mock function with given fields: config
func (_m *MockMaker) MakeConfigMapClient(config *types.KubernetesEnvironment) (clients.ConfigMapClient, error) {
	ret := _m.Called(config)

	if len(ret) == 0 {
		panic("no return value specified for MakeConfigMapClient")
	}

	var r0 clients.ConfigMapClient
	var r1 error
	if rf, ok := ret.Get(0).(func(*types.KubernetesEnvironment) (clients.ConfigMapClient, error)); ok {
		return rf(config)
	}
	if rf, ok := ret.Get(0).(func(*types.KubernetesEnvironment) clients.ConfigMapClient); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(clients.ConfigMapClient)
		}
	}

	if rf, ok := ret.Get(1).(func(*types.KubernetesEnvironment) error); ok {
		r1 = rf(config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaker_MakeConfigMapClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MakeConfigMapClient'
type MockMaker_MakeConfigMapClient_Call struct {
	*mock.Call
}

// MakeConfigMapClient is a helper method to define mock.On call
//   - config *types.KubernetesEnvironment
func (_e *MockMaker_Expecter) MakeConfigMapClient(config interface{}) *MockMaker_MakeConfigMapClient_Call {
	return &MockMaker_MakeConfigMapClient_Call{Call: _e.mock.On("MakeConfigMapClient", config)}
}

func (_c *MockMaker_MakeConfigMapClient_Call) Run(run func(config *types.KubernetesEnvironment)) *MockMaker_MakeConfigMapClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*types.KubernetesEnvironment))
	})
	return _c
}

func (_c *MockMaker_MakeConfigMapClient_Call) Return(_a0 clients.ConfigMapClient, _a1 error) *MockMaker_MakeConfigMapClient_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_MakeConfigMapClient_Call) RunAndReturn(run func(*types.KubernetesEnvironment) (clients.ConfigMapClient, error)) *MockMaker_MakeConfigMapClient_Call {
	_c.Call.Return(run)
	return _c
}

// MakeDeploymentClient provides a mock function with given fields: config
func (_m *MockMaker) MakeDeploymentClient(config *types.KubernetesEnvironment) (clients.DeploymentClient, error) {
	ret := _m.Called(config)

	if len(ret) == 0 {
		panic("no return value specified for MakeDeploymentClient")
	}

	var r0 clients.DeploymentClient
	var r1 error
	if rf, ok := ret.Get(0).(func(*types.KubernetesEnvironment) (clients.DeploymentClient, error)); ok {
		return rf(config)
	}
	if rf, ok := ret.Get(0).(func(*types.KubernetesEnvironment) clients.DeploymentClient); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(clients.DeploymentClient)
		}
	}

	if rf, ok := ret.Get(1).(func(*types.KubernetesEnvironment) error); ok {
		r1 = rf(config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaker_MakeDeploymentClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MakeDeploymentClient'
type MockMaker_MakeDeploymentClient_Call struct {
	*mock.Call
}

// MakeDeploymentClient is a helper method to define mock.On call
//   - config *types.KubernetesEnvironment
func (_e *MockMaker_Expecter) MakeDeploymentClient(config interface{}) *MockMaker_MakeDeploymentClient_Call {
	return &MockMaker_MakeDeploymentClient_Call{Call: _e.mock.On("MakeDeploymentClient", config)}
}

func (_c *MockMaker_MakeDeploymentClient_Call) Run(run func(config *types.KubernetesEnvironment)) *MockMaker_MakeDeploymentClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*types.KubernetesEnvironment))
	})
	return _c
}

func (_c *MockMaker_MakeDeploymentClient_Call) Return(_a0 clients.DeploymentClient, _a1 error) *MockMaker_MakeDeploymentClient_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_MakeDeploymentClient_Call) RunAndReturn(run func(*types.KubernetesEnvironment) (clients.DeploymentClient, error)) *MockMaker_MakeDeploymentClient_Call {
	_c.Call.Return(run)
	return _c
}

// MakePodClient provides a mock function with given fields: config
func (_m *MockMaker) MakePodClient(config *types.KubernetesEnvironment) (clients.PodClient, error) {
	ret := _m.Called(config)

	if len(ret) == 0 {
		panic("no return value specified for MakePodClient")
	}

	var r0 clients.PodClient
	var r1 error
	if rf, ok := ret.Get(0).(func(*types.KubernetesEnvironment) (clients.PodClient, error)); ok {
		return rf(config)
	}
	if rf, ok := ret.Get(0).(func(*types.KubernetesEnvironment) clients.PodClient); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(clients.PodClient)
		}
	}

	if rf, ok := ret.Get(1).(func(*types.KubernetesEnvironment) error); ok {
		r1 = rf(config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaker_MakePodClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MakePodClient'
type MockMaker_MakePodClient_Call struct {
	*mock.Call
}

// MakePodClient is a helper method to define mock.On call
//   - config *types.KubernetesEnvironment
func (_e *MockMaker_Expecter) MakePodClient(config interface{}) *MockMaker_MakePodClient_Call {
	return &MockMaker_MakePodClient_Call{Call: _e.mock.On("MakePodClient", config)}
}

func (_c *MockMaker_MakePodClient_Call) Run(run func(config *types.KubernetesEnvironment)) *MockMaker_MakePodClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*types.KubernetesEnvironment))
	})
	return _c
}

func (_c *MockMaker_MakePodClient_Call) Return(_a0 clients.PodClient, _a1 error) *MockMaker_MakePodClient_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_MakePodClient_Call) RunAndReturn(run func(*types.KubernetesEnvironment) (clients.PodClient, error)) *MockMaker_MakePodClient_Call {
	_c.Call.Return(run)
	return _c
}

// MakeServiceClient provides a mock function with given fields: config
func (_m *MockMaker) MakeServiceClient(config *types.KubernetesEnvironment) (clients.ServiceClient, error) {
	ret := _m.Called(config)

	if len(ret) == 0 {
		panic("no return value specified for MakeServiceClient")
	}

	var r0 clients.ServiceClient
	var r1 error
	if rf, ok := ret.Get(0).(func(*types.KubernetesEnvironment) (clients.ServiceClient, error)); ok {
		return rf(config)
	}
	if rf, ok := ret.Get(0).(func(*types.KubernetesEnvironment) clients.ServiceClient); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(clients.ServiceClient)
		}
	}

	if rf, ok := ret.Get(1).(func(*types.KubernetesEnvironment) error); ok {
		r1 = rf(config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaker_MakeServiceClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MakeServiceClient'
type MockMaker_MakeServiceClient_Call struct {
	*mock.Call
}

// MakeServiceClient is a helper method to define mock.On call
//   - config *types.KubernetesEnvironment
func (_e *MockMaker_Expecter) MakeServiceClient(config interface{}) *MockMaker_MakeServiceClient_Call {
	return &MockMaker_MakeServiceClient_Call{Call: _e.mock.On("MakeServiceClient", config)}
}

func (_c *MockMaker_MakeServiceClient_Call) Run(run func(config *types.KubernetesEnvironment)) *MockMaker_MakeServiceClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*types.KubernetesEnvironment))
	})
	return _c
}

func (_c *MockMaker_MakeServiceClient_Call) Return(_a0 clients.ServiceClient, _a1 error) *MockMaker_MakeServiceClient_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaker_MakeServiceClient_Call) RunAndReturn(run func(*types.KubernetesEnvironment) (clients.ServiceClient, error)) *MockMaker_MakeServiceClient_Call {
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
