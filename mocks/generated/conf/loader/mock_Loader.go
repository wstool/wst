// Code generated by mockery v2.40.1. DO NOT EDIT.

package loader

import (
	mock "github.com/stretchr/testify/mock"
	loader "github.com/wstool/wst/conf/loader"
)

// MockLoader is an autogenerated mock type for the Loader type
type MockLoader struct {
	mock.Mock
}

type MockLoader_Expecter struct {
	mock *mock.Mock
}

func (_m *MockLoader) EXPECT() *MockLoader_Expecter {
	return &MockLoader_Expecter{mock: &_m.Mock}
}

// GlobConfigs provides a mock function with given fields: pattern, cwd
func (_m *MockLoader) GlobConfigs(pattern string, cwd string) ([]loader.LoadedConfig, error) {
	ret := _m.Called(pattern, cwd)

	if len(ret) == 0 {
		panic("no return value specified for GlobConfigs")
	}

	var r0 []loader.LoadedConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) ([]loader.LoadedConfig, error)); ok {
		return rf(pattern, cwd)
	}
	if rf, ok := ret.Get(0).(func(string, string) []loader.LoadedConfig); ok {
		r0 = rf(pattern, cwd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]loader.LoadedConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(pattern, cwd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockLoader_GlobConfigs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GlobConfigs'
type MockLoader_GlobConfigs_Call struct {
	*mock.Call
}

// GlobConfigs is a helper method to define mock.On call
//   - pattern string
//   - cwd string
func (_e *MockLoader_Expecter) GlobConfigs(pattern interface{}, cwd interface{}) *MockLoader_GlobConfigs_Call {
	return &MockLoader_GlobConfigs_Call{Call: _e.mock.On("GlobConfigs", pattern, cwd)}
}

func (_c *MockLoader_GlobConfigs_Call) Run(run func(pattern string, cwd string)) *MockLoader_GlobConfigs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockLoader_GlobConfigs_Call) Return(_a0 []loader.LoadedConfig, _a1 error) *MockLoader_GlobConfigs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockLoader_GlobConfigs_Call) RunAndReturn(run func(string, string) ([]loader.LoadedConfig, error)) *MockLoader_GlobConfigs_Call {
	_c.Call.Return(run)
	return _c
}

// LoadConfig provides a mock function with given fields: path
func (_m *MockLoader) LoadConfig(path string) (loader.LoadedConfig, error) {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for LoadConfig")
	}

	var r0 loader.LoadedConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (loader.LoadedConfig, error)); ok {
		return rf(path)
	}
	if rf, ok := ret.Get(0).(func(string) loader.LoadedConfig); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(loader.LoadedConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockLoader_LoadConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoadConfig'
type MockLoader_LoadConfig_Call struct {
	*mock.Call
}

// LoadConfig is a helper method to define mock.On call
//   - path string
func (_e *MockLoader_Expecter) LoadConfig(path interface{}) *MockLoader_LoadConfig_Call {
	return &MockLoader_LoadConfig_Call{Call: _e.mock.On("LoadConfig", path)}
}

func (_c *MockLoader_LoadConfig_Call) Run(run func(path string)) *MockLoader_LoadConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockLoader_LoadConfig_Call) Return(_a0 loader.LoadedConfig, _a1 error) *MockLoader_LoadConfig_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockLoader_LoadConfig_Call) RunAndReturn(run func(string) (loader.LoadedConfig, error)) *MockLoader_LoadConfig_Call {
	_c.Call.Return(run)
	return _c
}

// LoadConfigs provides a mock function with given fields: paths
func (_m *MockLoader) LoadConfigs(paths []string) ([]loader.LoadedConfig, error) {
	ret := _m.Called(paths)

	if len(ret) == 0 {
		panic("no return value specified for LoadConfigs")
	}

	var r0 []loader.LoadedConfig
	var r1 error
	if rf, ok := ret.Get(0).(func([]string) ([]loader.LoadedConfig, error)); ok {
		return rf(paths)
	}
	if rf, ok := ret.Get(0).(func([]string) []loader.LoadedConfig); ok {
		r0 = rf(paths)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]loader.LoadedConfig)
		}
	}

	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(paths)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockLoader_LoadConfigs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoadConfigs'
type MockLoader_LoadConfigs_Call struct {
	*mock.Call
}

// LoadConfigs is a helper method to define mock.On call
//   - paths []string
func (_e *MockLoader_Expecter) LoadConfigs(paths interface{}) *MockLoader_LoadConfigs_Call {
	return &MockLoader_LoadConfigs_Call{Call: _e.mock.On("LoadConfigs", paths)}
}

func (_c *MockLoader_LoadConfigs_Call) Run(run func(paths []string)) *MockLoader_LoadConfigs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *MockLoader_LoadConfigs_Call) Return(_a0 []loader.LoadedConfig, _a1 error) *MockLoader_LoadConfigs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockLoader_LoadConfigs_Call) RunAndReturn(run func([]string) ([]loader.LoadedConfig, error)) *MockLoader_LoadConfigs_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockLoader creates a new instance of MockLoader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockLoader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLoader {
	mock := &MockLoader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
