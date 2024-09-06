// Code generated by mockery v2.40.1. DO NOT EDIT.

package parameter

import (
	mock "github.com/stretchr/testify/mock"
	parameter "github.com/wstool/wst/run/parameters/parameter"
)

// MockParameter is an autogenerated mock type for the Parameter type
type MockParameter struct {
	mock.Mock
}

type MockParameter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockParameter) EXPECT() *MockParameter_Expecter {
	return &MockParameter_Expecter{mock: &_m.Mock}
}

// ArrayValue provides a mock function with given fields:
func (_m *MockParameter) ArrayValue() []parameter.Parameter {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ArrayValue")
	}

	var r0 []parameter.Parameter
	if rf, ok := ret.Get(0).(func() []parameter.Parameter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]parameter.Parameter)
		}
	}

	return r0
}

// MockParameter_ArrayValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ArrayValue'
type MockParameter_ArrayValue_Call struct {
	*mock.Call
}

// ArrayValue is a helper method to define mock.On call
func (_e *MockParameter_Expecter) ArrayValue() *MockParameter_ArrayValue_Call {
	return &MockParameter_ArrayValue_Call{Call: _e.mock.On("ArrayValue")}
}

func (_c *MockParameter_ArrayValue_Call) Run(run func()) *MockParameter_ArrayValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockParameter_ArrayValue_Call) Return(_a0 []parameter.Parameter) *MockParameter_ArrayValue_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParameter_ArrayValue_Call) RunAndReturn(run func() []parameter.Parameter) *MockParameter_ArrayValue_Call {
	_c.Call.Return(run)
	return _c
}

// BoolValue provides a mock function with given fields:
func (_m *MockParameter) BoolValue() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for BoolValue")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockParameter_BoolValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BoolValue'
type MockParameter_BoolValue_Call struct {
	*mock.Call
}

// BoolValue is a helper method to define mock.On call
func (_e *MockParameter_Expecter) BoolValue() *MockParameter_BoolValue_Call {
	return &MockParameter_BoolValue_Call{Call: _e.mock.On("BoolValue")}
}

func (_c *MockParameter_BoolValue_Call) Run(run func()) *MockParameter_BoolValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockParameter_BoolValue_Call) Return(_a0 bool) *MockParameter_BoolValue_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParameter_BoolValue_Call) RunAndReturn(run func() bool) *MockParameter_BoolValue_Call {
	_c.Call.Return(run)
	return _c
}

// FloatValue provides a mock function with given fields:
func (_m *MockParameter) FloatValue() float64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for FloatValue")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// MockParameter_FloatValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FloatValue'
type MockParameter_FloatValue_Call struct {
	*mock.Call
}

// FloatValue is a helper method to define mock.On call
func (_e *MockParameter_Expecter) FloatValue() *MockParameter_FloatValue_Call {
	return &MockParameter_FloatValue_Call{Call: _e.mock.On("FloatValue")}
}

func (_c *MockParameter_FloatValue_Call) Run(run func()) *MockParameter_FloatValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockParameter_FloatValue_Call) Return(_a0 float64) *MockParameter_FloatValue_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParameter_FloatValue_Call) RunAndReturn(run func() float64) *MockParameter_FloatValue_Call {
	_c.Call.Return(run)
	return _c
}

// IntValue provides a mock function with given fields:
func (_m *MockParameter) IntValue() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IntValue")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// MockParameter_IntValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IntValue'
type MockParameter_IntValue_Call struct {
	*mock.Call
}

// IntValue is a helper method to define mock.On call
func (_e *MockParameter_Expecter) IntValue() *MockParameter_IntValue_Call {
	return &MockParameter_IntValue_Call{Call: _e.mock.On("IntValue")}
}

func (_c *MockParameter_IntValue_Call) Run(run func()) *MockParameter_IntValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockParameter_IntValue_Call) Return(_a0 int) *MockParameter_IntValue_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParameter_IntValue_Call) RunAndReturn(run func() int) *MockParameter_IntValue_Call {
	_c.Call.Return(run)
	return _c
}

// MapValue provides a mock function with given fields:
func (_m *MockParameter) MapValue() map[string]parameter.Parameter {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for MapValue")
	}

	var r0 map[string]parameter.Parameter
	if rf, ok := ret.Get(0).(func() map[string]parameter.Parameter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]parameter.Parameter)
		}
	}

	return r0
}

// MockParameter_MapValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MapValue'
type MockParameter_MapValue_Call struct {
	*mock.Call
}

// MapValue is a helper method to define mock.On call
func (_e *MockParameter_Expecter) MapValue() *MockParameter_MapValue_Call {
	return &MockParameter_MapValue_Call{Call: _e.mock.On("MapValue")}
}

func (_c *MockParameter_MapValue_Call) Run(run func()) *MockParameter_MapValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockParameter_MapValue_Call) Return(_a0 map[string]parameter.Parameter) *MockParameter_MapValue_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParameter_MapValue_Call) RunAndReturn(run func() map[string]parameter.Parameter) *MockParameter_MapValue_Call {
	_c.Call.Return(run)
	return _c
}

// StringValue provides a mock function with given fields:
func (_m *MockParameter) StringValue() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for StringValue")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockParameter_StringValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StringValue'
type MockParameter_StringValue_Call struct {
	*mock.Call
}

// StringValue is a helper method to define mock.On call
func (_e *MockParameter_Expecter) StringValue() *MockParameter_StringValue_Call {
	return &MockParameter_StringValue_Call{Call: _e.mock.On("StringValue")}
}

func (_c *MockParameter_StringValue_Call) Run(run func()) *MockParameter_StringValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockParameter_StringValue_Call) Return(_a0 string) *MockParameter_StringValue_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParameter_StringValue_Call) RunAndReturn(run func() string) *MockParameter_StringValue_Call {
	_c.Call.Return(run)
	return _c
}

// Type provides a mock function with given fields:
func (_m *MockParameter) Type() parameter.Type {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Type")
	}

	var r0 parameter.Type
	if rf, ok := ret.Get(0).(func() parameter.Type); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(parameter.Type)
	}

	return r0
}

// MockParameter_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type MockParameter_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *MockParameter_Expecter) Type() *MockParameter_Type_Call {
	return &MockParameter_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *MockParameter_Type_Call) Run(run func()) *MockParameter_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockParameter_Type_Call) Return(_a0 parameter.Type) *MockParameter_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockParameter_Type_Call) RunAndReturn(run func() parameter.Type) *MockParameter_Type_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockParameter creates a new instance of MockParameter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockParameter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockParameter {
	mock := &MockParameter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
