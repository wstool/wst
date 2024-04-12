// Code generated by mockery v2.40.1. DO NOT EDIT.

package merger

import (
	types "github.com/bukka/wst/conf/types"
	mock "github.com/stretchr/testify/mock"
)

// MockMerger is an autogenerated mock type for the Merger type
type MockMerger struct {
	mock.Mock
}

type MockMerger_Expecter struct {
	mock *mock.Mock
}

func (_m *MockMerger) EXPECT() *MockMerger_Expecter {
	return &MockMerger_Expecter{mock: &_m.Mock}
}

// MergeConfigs provides a mock function with given fields: configs
func (_m *MockMerger) MergeConfigs(configs []*types.Config) (*types.Config, error) {
	ret := _m.Called(configs)

	if len(ret) == 0 {
		panic("no return value specified for MergeConfigs")
	}

	var r0 *types.Config
	var r1 error
	if rf, ok := ret.Get(0).(func([]*types.Config) (*types.Config, error)); ok {
		return rf(configs)
	}
	if rf, ok := ret.Get(0).(func([]*types.Config) *types.Config); ok {
		r0 = rf(configs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Config)
		}
	}

	if rf, ok := ret.Get(1).(func([]*types.Config) error); ok {
		r1 = rf(configs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMerger_MergeConfigs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MergeConfigs'
type MockMerger_MergeConfigs_Call struct {
	*mock.Call
}

// MergeConfigs is a helper method to define mock.On call
//   - configs []*types.Config
func (_e *MockMerger_Expecter) MergeConfigs(configs interface{}) *MockMerger_MergeConfigs_Call {
	return &MockMerger_MergeConfigs_Call{Call: _e.mock.On("MergeConfigs", configs)}
}

func (_c *MockMerger_MergeConfigs_Call) Run(run func(configs []*types.Config)) *MockMerger_MergeConfigs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]*types.Config))
	})
	return _c
}

func (_c *MockMerger_MergeConfigs_Call) Return(_a0 *types.Config, _a1 error) *MockMerger_MergeConfigs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMerger_MergeConfigs_Call) RunAndReturn(run func([]*types.Config) (*types.Config, error)) *MockMerger_MergeConfigs_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockMerger creates a new instance of MockMerger. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMerger(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMerger {
	mock := &MockMerger{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
