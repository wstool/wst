package external

import "github.com/stretchr/testify/mock"

// MockMiniParser is a minimal parser
type MockMiniParser struct {
	mock.Mock
}

type MockMiniParser_Expecter struct {
	mock *mock.Mock
}

// ParseStruct provides a mock function with given fields: data, structure, configPath
func (_m *MockMiniParser) ParseStruct(data map[string]interface{}, structure interface{}, configPath string) error {
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

// MockMiniParser_ParseStruct_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ParseStruct'
type MockMiniParser_ParseStruct_Call struct {
	*mock.Call
}

// ParseStruct is a helper method to define mock.On call
//   - data map[string]interface{}
//   - structure interface{}
//   - configPath string
func (_e *MockMiniParser_Expecter) ParseStruct(data interface{}, structure interface{}, configPath interface{}) *MockMiniParser_ParseStruct_Call {
	return &MockMiniParser_ParseStruct_Call{Call: _e.mock.On("ParseStruct", data, structure, configPath)}
}

func (_c *MockMiniParser_ParseStruct_Call) Run(run func(data map[string]interface{}, structure interface{}, configPath string)) *MockMiniParser_ParseStruct_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]interface{}), args[1].(interface{}), args[2].(string))
	})
	return _c
}

func (_c *MockMiniParser_ParseStruct_Call) Return(_a0 error) *MockMiniParser_ParseStruct_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMiniParser_ParseStruct_Call) RunAndReturn(run func(map[string]interface{}, interface{}, string) error) *MockMiniParser_ParseStruct_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockParser creates a new instance of MockParser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMiniParser(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMiniParser {
	mock := &MockMiniParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
