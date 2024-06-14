// Code generated by mockery v2.40.1. DO NOT EDIT.

package parser

import (
	parser "github.com/bukka/wst/conf/parser"
	mock "github.com/stretchr/testify/mock"

	types "github.com/bukka/wst/conf/types"
)

// MockParser is an autogenerated mock type for the Parser type
type MockParser struct {
	mock.Mock
}

type MockParser_Expecter struct {
	mock *mock.Mock
}

func (_m *MockParser) EXPECT() *MockParser_Expecter {
	return &MockParser_Expecter{mock: &_m.Mock}
}

// ParseConfig provides a mock function with given fields: data, config, configPath
func (_m *MockParser) ParseConfig(data map[string]interface{}, config *types.Config, configPath string) error {
	ret := _m.Called(data, config, configPath)

	if len(ret) == 0 {
		panic("no return value specified for ParseConfig")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(map[string]interface{}, *types.Config, string) error); ok {
		r0 = rf(data, config, configPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockParser_ParseConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ParseConfig'
type MockParser_ParseConfig_Call struct {
	*mock.Call
}

// ParseConfig is a helper method to define mock.On call
//   - data map[string]interface{}
//   - config *types.Config
//   - configPath string
func (_e *MockParser_Expecter) ParseConfig(data interface{}, config interface{}, configPath interface{}) *MockParser_ParseConfig_Call {
	return &MockParser_ParseConfig_Call{Call: _e.mock.On("ParseConfig", data, config, configPath)}
}

func (_c *MockParser_ParseConfig_Call) Run(run func(data map[string]interface{}, config *types.Config, configPath string)) *MockParser_ParseConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]interface{}), args[1].(*types.Config), args[2].(string))
	})
	return _c
}

func (_c *MockParser_ParseConfig_Call) Return(_a0 error) *MockParser_ParseConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParser_ParseConfig_Call) RunAndReturn(run func(map[string]interface{}, *types.Config, string) error) *MockParser_ParseConfig_Call {
	_c.Call.Return(run)
	return _c
}

// ParseStruct provides a mock function with given fields: data, structure, configPath
func (_m *MockParser) ParseStruct(data map[string]interface{}, structure interface{}, configPath string) error {
	ret := _m.Called(data, structure, configPath)

	if len(ret) == 0 {
		panic("no return value specified for ParseStruct")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(map[string]interface{}, interface{}, string) error); ok {
		r0 = rf(data, structure, configPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockParser_ParseStruct_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ParseStruct'
type MockParser_ParseStruct_Call struct {
	*mock.Call
}

// ParseStruct is a helper method to define mock.On call
//   - data map[string]interface{}
//   - structure interface{}
//   - configPath string
func (_e *MockParser_Expecter) ParseStruct(data interface{}, structure interface{}, configPath interface{}) *MockParser_ParseStruct_Call {
	return &MockParser_ParseStruct_Call{Call: _e.mock.On("ParseStruct", data, structure, configPath)}
}

func (_c *MockParser_ParseStruct_Call) Run(run func(data map[string]interface{}, structure interface{}, configPath string)) *MockParser_ParseStruct_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]interface{}), args[1].(interface{}), args[2].(string))
	})
	return _c
}

func (_c *MockParser_ParseStruct_Call) Return(_a0 error) *MockParser_ParseStruct_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParser_ParseStruct_Call) RunAndReturn(run func(map[string]interface{}, interface{}, string) error) *MockParser_ParseStruct_Call {
	_c.Call.Return(run)
	return _c
}

// ParseTag provides a mock function with given fields: tag
func (_m *MockParser) ParseTag(tag string) (map[parser.ConfigParam]string, error) {
	ret := _m.Called(tag)

	if len(ret) == 0 {
		panic("no return value specified for ParseTag")
	}

	var r0 map[parser.ConfigParam]string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (map[parser.ConfigParam]string, error)); ok {
		return rf(tag)
	}
	if rf, ok := ret.Get(0).(func(string) map[parser.ConfigParam]string); ok {
		r0 = rf(tag)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[parser.ConfigParam]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tag)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockParser_ParseTag_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ParseTag'
type MockParser_ParseTag_Call struct {
	*mock.Call
}

// ParseTag is a helper method to define mock.On call
//   - tag string
func (_e *MockParser_Expecter) ParseTag(tag interface{}) *MockParser_ParseTag_Call {
	return &MockParser_ParseTag_Call{Call: _e.mock.On("ParseTag", tag)}
}

func (_c *MockParser_ParseTag_Call) Run(run func(tag string)) *MockParser_ParseTag_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockParser_ParseTag_Call) Return(_a0 map[parser.ConfigParam]string, _a1 error) *MockParser_ParseTag_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockParser_ParseTag_Call) RunAndReturn(run func(string) (map[parser.ConfigParam]string, error)) *MockParser_ParseTag_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockParser creates a new instance of MockParser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockParser(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockParser {
	mock := &MockParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}