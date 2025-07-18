// Code generated by mockery. DO NOT EDIT.

package internal

import (
	config "github.com/dotboris/standard-backups/internal/config"
	mock "github.com/stretchr/testify/mock"
)

// MocknewBackendClienter is an autogenerated mock type for the newBackendClienter type
type MocknewBackendClienter struct {
	mock.Mock
}

type MocknewBackendClienter_Expecter struct {
	mock *mock.Mock
}

func (_m *MocknewBackendClienter) EXPECT() *MocknewBackendClienter_Expecter {
	return &MocknewBackendClienter_Expecter{mock: &_m.Mock}
}

// NewBackendClient provides a mock function with given fields: cfg, name
func (_m *MocknewBackendClienter) NewBackendClient(cfg config.Config, name string) (backuper, error) {
	ret := _m.Called(cfg, name)

	if len(ret) == 0 {
		panic("no return value specified for NewBackendClient")
	}

	var r0 backuper
	var r1 error
	if rf, ok := ret.Get(0).(func(config.Config, string) (backuper, error)); ok {
		return rf(cfg, name)
	}
	if rf, ok := ret.Get(0).(func(config.Config, string) backuper); ok {
		r0 = rf(cfg, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(backuper)
		}
	}

	if rf, ok := ret.Get(1).(func(config.Config, string) error); ok {
		r1 = rf(cfg, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MocknewBackendClienter_NewBackendClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewBackendClient'
type MocknewBackendClienter_NewBackendClient_Call struct {
	*mock.Call
}

// NewBackendClient is a helper method to define mock.On call
//   - cfg config.Config
//   - name string
func (_e *MocknewBackendClienter_Expecter) NewBackendClient(cfg interface{}, name interface{}) *MocknewBackendClienter_NewBackendClient_Call {
	return &MocknewBackendClienter_NewBackendClient_Call{Call: _e.mock.On("NewBackendClient", cfg, name)}
}

func (_c *MocknewBackendClienter_NewBackendClient_Call) Run(run func(cfg config.Config, name string)) *MocknewBackendClienter_NewBackendClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(config.Config), args[1].(string))
	})
	return _c
}

func (_c *MocknewBackendClienter_NewBackendClient_Call) Return(_a0 backuper, _a1 error) *MocknewBackendClienter_NewBackendClient_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MocknewBackendClienter_NewBackendClient_Call) RunAndReturn(run func(config.Config, string) (backuper, error)) *MocknewBackendClienter_NewBackendClient_Call {
	_c.Call.Return(run)
	return _c
}

// NewMocknewBackendClienter creates a new instance of MocknewBackendClienter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMocknewBackendClienter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MocknewBackendClienter {
	mock := &MocknewBackendClienter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
