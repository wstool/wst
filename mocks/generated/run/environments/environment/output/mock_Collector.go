// Code generated by mockery v2.40.1. DO NOT EDIT.

package output

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	output "github.com/wstool/wst/run/environments/environment/output"
)

// MockCollector is an autogenerated mock type for the Collector type
type MockCollector struct {
	mock.Mock
}

type MockCollector_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCollector) EXPECT() *MockCollector_Expecter {
	return &MockCollector_Expecter{mock: &_m.Mock}
}

// AnyReader provides a mock function with given fields: ctx
func (_m *MockCollector) AnyReader(ctx context.Context) io.Reader {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for AnyReader")
	}

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func(context.Context) io.Reader); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	return r0
}

// MockCollector_AnyReader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AnyReader'
type MockCollector_AnyReader_Call struct {
	*mock.Call
}

// AnyReader is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockCollector_Expecter) AnyReader(ctx interface{}) *MockCollector_AnyReader_Call {
	return &MockCollector_AnyReader_Call{Call: _e.mock.On("AnyReader", ctx)}
}

func (_c *MockCollector_AnyReader_Call) Run(run func(ctx context.Context)) *MockCollector_AnyReader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockCollector_AnyReader_Call) Return(_a0 io.Reader) *MockCollector_AnyReader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollector_AnyReader_Call) RunAndReturn(run func(context.Context) io.Reader) *MockCollector_AnyReader_Call {
	_c.Call.Return(run)
	return _c
}

// Close provides a mock function with given fields:
func (_m *MockCollector) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCollector_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockCollector_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockCollector_Expecter) Close() *MockCollector_Close_Call {
	return &MockCollector_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockCollector_Close_Call) Run(run func()) *MockCollector_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCollector_Close_Call) Return(_a0 error) *MockCollector_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollector_Close_Call) RunAndReturn(run func() error) *MockCollector_Close_Call {
	_c.Call.Return(run)
	return _c
}

// LogOutput provides a mock function with given fields:
func (_m *MockCollector) LogOutput() {
	_m.Called()
}

// MockCollector_LogOutput_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LogOutput'
type MockCollector_LogOutput_Call struct {
	*mock.Call
}

// LogOutput is a helper method to define mock.On call
func (_e *MockCollector_Expecter) LogOutput() *MockCollector_LogOutput_Call {
	return &MockCollector_LogOutput_Call{Call: _e.mock.On("LogOutput")}
}

func (_c *MockCollector_LogOutput_Call) Run(run func()) *MockCollector_LogOutput_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCollector_LogOutput_Call) Return() *MockCollector_LogOutput_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockCollector_LogOutput_Call) RunAndReturn(run func()) *MockCollector_LogOutput_Call {
	_c.Call.Return(run)
	return _c
}

// Reader provides a mock function with given fields: ctx, outputType
func (_m *MockCollector) Reader(ctx context.Context, outputType output.Type) (io.Reader, error) {
	ret := _m.Called(ctx, outputType)

	if len(ret) == 0 {
		panic("no return value specified for Reader")
	}

	var r0 io.Reader
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, output.Type) (io.Reader, error)); ok {
		return rf(ctx, outputType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, output.Type) io.Reader); ok {
		r0 = rf(ctx, outputType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, output.Type) error); ok {
		r1 = rf(ctx, outputType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollector_Reader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reader'
type MockCollector_Reader_Call struct {
	*mock.Call
}

// Reader is a helper method to define mock.On call
//   - ctx context.Context
//   - outputType output.Type
func (_e *MockCollector_Expecter) Reader(ctx interface{}, outputType interface{}) *MockCollector_Reader_Call {
	return &MockCollector_Reader_Call{Call: _e.mock.On("Reader", ctx, outputType)}
}

func (_c *MockCollector_Reader_Call) Run(run func(ctx context.Context, outputType output.Type)) *MockCollector_Reader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(output.Type))
	})
	return _c
}

func (_c *MockCollector_Reader_Call) Return(_a0 io.Reader, _a1 error) *MockCollector_Reader_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollector_Reader_Call) RunAndReturn(run func(context.Context, output.Type) (io.Reader, error)) *MockCollector_Reader_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with given fields: stdoutPipe, stderrPipe
func (_m *MockCollector) Start(stdoutPipe io.ReadCloser, stderrPipe io.ReadCloser) error {
	ret := _m.Called(stdoutPipe, stderrPipe)

	if len(ret) == 0 {
		panic("no return value specified for Start")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(io.ReadCloser, io.ReadCloser) error); ok {
		r0 = rf(stdoutPipe, stderrPipe)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCollector_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type MockCollector_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
//   - stdoutPipe io.ReadCloser
//   - stderrPipe io.ReadCloser
func (_e *MockCollector_Expecter) Start(stdoutPipe interface{}, stderrPipe interface{}) *MockCollector_Start_Call {
	return &MockCollector_Start_Call{Call: _e.mock.On("Start", stdoutPipe, stderrPipe)}
}

func (_c *MockCollector_Start_Call) Run(run func(stdoutPipe io.ReadCloser, stderrPipe io.ReadCloser)) *MockCollector_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(io.ReadCloser), args[1].(io.ReadCloser))
	})
	return _c
}

func (_c *MockCollector_Start_Call) Return(_a0 error) *MockCollector_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollector_Start_Call) RunAndReturn(run func(io.ReadCloser, io.ReadCloser) error) *MockCollector_Start_Call {
	_c.Call.Return(run)
	return _c
}

// StderrReader provides a mock function with given fields: ctx
func (_m *MockCollector) StderrReader(ctx context.Context) io.Reader {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for StderrReader")
	}

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func(context.Context) io.Reader); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	return r0
}

// MockCollector_StderrReader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StderrReader'
type MockCollector_StderrReader_Call struct {
	*mock.Call
}

// StderrReader is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockCollector_Expecter) StderrReader(ctx interface{}) *MockCollector_StderrReader_Call {
	return &MockCollector_StderrReader_Call{Call: _e.mock.On("StderrReader", ctx)}
}

func (_c *MockCollector_StderrReader_Call) Run(run func(ctx context.Context)) *MockCollector_StderrReader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockCollector_StderrReader_Call) Return(_a0 io.Reader) *MockCollector_StderrReader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollector_StderrReader_Call) RunAndReturn(run func(context.Context) io.Reader) *MockCollector_StderrReader_Call {
	_c.Call.Return(run)
	return _c
}

// StderrWriter provides a mock function with given fields:
func (_m *MockCollector) StderrWriter() io.Writer {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for StderrWriter")
	}

	var r0 io.Writer
	if rf, ok := ret.Get(0).(func() io.Writer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Writer)
		}
	}

	return r0
}

// MockCollector_StderrWriter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StderrWriter'
type MockCollector_StderrWriter_Call struct {
	*mock.Call
}

// StderrWriter is a helper method to define mock.On call
func (_e *MockCollector_Expecter) StderrWriter() *MockCollector_StderrWriter_Call {
	return &MockCollector_StderrWriter_Call{Call: _e.mock.On("StderrWriter")}
}

func (_c *MockCollector_StderrWriter_Call) Run(run func()) *MockCollector_StderrWriter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCollector_StderrWriter_Call) Return(_a0 io.Writer) *MockCollector_StderrWriter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollector_StderrWriter_Call) RunAndReturn(run func() io.Writer) *MockCollector_StderrWriter_Call {
	_c.Call.Return(run)
	return _c
}

// StdoutReader provides a mock function with given fields: ctx
func (_m *MockCollector) StdoutReader(ctx context.Context) io.Reader {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for StdoutReader")
	}

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func(context.Context) io.Reader); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	return r0
}

// MockCollector_StdoutReader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StdoutReader'
type MockCollector_StdoutReader_Call struct {
	*mock.Call
}

// StdoutReader is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockCollector_Expecter) StdoutReader(ctx interface{}) *MockCollector_StdoutReader_Call {
	return &MockCollector_StdoutReader_Call{Call: _e.mock.On("StdoutReader", ctx)}
}

func (_c *MockCollector_StdoutReader_Call) Run(run func(ctx context.Context)) *MockCollector_StdoutReader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockCollector_StdoutReader_Call) Return(_a0 io.Reader) *MockCollector_StdoutReader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollector_StdoutReader_Call) RunAndReturn(run func(context.Context) io.Reader) *MockCollector_StdoutReader_Call {
	_c.Call.Return(run)
	return _c
}

// StdoutWriter provides a mock function with given fields:
func (_m *MockCollector) StdoutWriter() io.Writer {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for StdoutWriter")
	}

	var r0 io.Writer
	if rf, ok := ret.Get(0).(func() io.Writer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Writer)
		}
	}

	return r0
}

// MockCollector_StdoutWriter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StdoutWriter'
type MockCollector_StdoutWriter_Call struct {
	*mock.Call
}

// StdoutWriter is a helper method to define mock.On call
func (_e *MockCollector_Expecter) StdoutWriter() *MockCollector_StdoutWriter_Call {
	return &MockCollector_StdoutWriter_Call{Call: _e.mock.On("StdoutWriter")}
}

func (_c *MockCollector_StdoutWriter_Call) Run(run func()) *MockCollector_StdoutWriter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCollector_StdoutWriter_Call) Return(_a0 io.Writer) *MockCollector_StdoutWriter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollector_StdoutWriter_Call) RunAndReturn(run func() io.Writer) *MockCollector_StdoutWriter_Call {
	_c.Call.Return(run)
	return _c
}

// Wait provides a mock function with given fields:
func (_m *MockCollector) Wait() {
	_m.Called()
}

// MockCollector_Wait_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Wait'
type MockCollector_Wait_Call struct {
	*mock.Call
}

// Wait is a helper method to define mock.On call
func (_e *MockCollector_Expecter) Wait() *MockCollector_Wait_Call {
	return &MockCollector_Wait_Call{Call: _e.mock.On("Wait")}
}

func (_c *MockCollector_Wait_Call) Run(run func()) *MockCollector_Wait_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCollector_Wait_Call) Return() *MockCollector_Wait_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockCollector_Wait_Call) RunAndReturn(run func()) *MockCollector_Wait_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCollector creates a new instance of MockCollector. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCollector(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCollector {
	mock := &MockCollector{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
