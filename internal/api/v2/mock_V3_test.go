// Code generated by mockery v2.14.0. DO NOT EDIT.

package api

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	private "gitlab.com/accumulatenetwork/accumulate/internal/api/private"
	v3 "gitlab.com/accumulatenetwork/accumulate/pkg/api/v3"
	messaging "gitlab.com/accumulatenetwork/accumulate/pkg/types/messaging"
	url "gitlab.com/accumulatenetwork/accumulate/pkg/url"
)

// MockV3 is an autogenerated mock type for the V3 type
type MockV3 struct {
	mock.Mock
}

type MockV3_Expecter struct {
	mock *mock.Mock
}

func (_m *MockV3) EXPECT() *MockV3_Expecter {
	return &MockV3_Expecter{mock: &_m.Mock}
}

// Metrics provides a mock function with given fields: ctx, opts
func (_m *MockV3) Metrics(ctx context.Context, opts v3.MetricsOptions) (*v3.Metrics, error) {
	ret := _m.Called(ctx, opts)

	var r0 *v3.Metrics
	if rf, ok := ret.Get(0).(func(context.Context, v3.MetricsOptions) *v3.Metrics); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3.Metrics)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, v3.MetricsOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockV3_Metrics_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Metrics'
type MockV3_Metrics_Call struct {
	*mock.Call
}

// Metrics is a helper method to define mock.On call
//   - ctx context.Context
//   - opts v3.MetricsOptions
func (_e *MockV3_Expecter) Metrics(ctx interface{}, opts interface{}) *MockV3_Metrics_Call {
	return &MockV3_Metrics_Call{Call: _e.mock.On("Metrics", ctx, opts)}
}

func (_c *MockV3_Metrics_Call) Run(run func(ctx context.Context, opts v3.MetricsOptions)) *MockV3_Metrics_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(v3.MetricsOptions))
	})
	return _c
}

func (_c *MockV3_Metrics_Call) Return(_a0 *v3.Metrics, _a1 error) *MockV3_Metrics_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// NetworkStatus provides a mock function with given fields: ctx, opts
func (_m *MockV3) NetworkStatus(ctx context.Context, opts v3.NetworkStatusOptions) (*v3.NetworkStatus, error) {
	ret := _m.Called(ctx, opts)

	var r0 *v3.NetworkStatus
	if rf, ok := ret.Get(0).(func(context.Context, v3.NetworkStatusOptions) *v3.NetworkStatus); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3.NetworkStatus)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, v3.NetworkStatusOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockV3_NetworkStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NetworkStatus'
type MockV3_NetworkStatus_Call struct {
	*mock.Call
}

// NetworkStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - opts v3.NetworkStatusOptions
func (_e *MockV3_Expecter) NetworkStatus(ctx interface{}, opts interface{}) *MockV3_NetworkStatus_Call {
	return &MockV3_NetworkStatus_Call{Call: _e.mock.On("NetworkStatus", ctx, opts)}
}

func (_c *MockV3_NetworkStatus_Call) Run(run func(ctx context.Context, opts v3.NetworkStatusOptions)) *MockV3_NetworkStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(v3.NetworkStatusOptions))
	})
	return _c
}

func (_c *MockV3_NetworkStatus_Call) Return(_a0 *v3.NetworkStatus, _a1 error) *MockV3_NetworkStatus_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// NodeStatus provides a mock function with given fields: ctx, opts
func (_m *MockV3) NodeStatus(ctx context.Context, opts v3.NodeStatusOptions) (*v3.NodeStatus, error) {
	ret := _m.Called(ctx, opts)

	var r0 *v3.NodeStatus
	if rf, ok := ret.Get(0).(func(context.Context, v3.NodeStatusOptions) *v3.NodeStatus); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3.NodeStatus)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, v3.NodeStatusOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockV3_NodeStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NodeStatus'
type MockV3_NodeStatus_Call struct {
	*mock.Call
}

// NodeStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - opts v3.NodeStatusOptions
func (_e *MockV3_Expecter) NodeStatus(ctx interface{}, opts interface{}) *MockV3_NodeStatus_Call {
	return &MockV3_NodeStatus_Call{Call: _e.mock.On("NodeStatus", ctx, opts)}
}

func (_c *MockV3_NodeStatus_Call) Run(run func(ctx context.Context, opts v3.NodeStatusOptions)) *MockV3_NodeStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(v3.NodeStatusOptions))
	})
	return _c
}

func (_c *MockV3_NodeStatus_Call) Return(_a0 *v3.NodeStatus, _a1 error) *MockV3_NodeStatus_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Private provides a mock function with given fields:
func (_m *MockV3) Private() private.Sequencer {
	ret := _m.Called()

	var r0 private.Sequencer
	if rf, ok := ret.Get(0).(func() private.Sequencer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(private.Sequencer)
		}
	}

	return r0
}

// MockV3_Private_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Private'
type MockV3_Private_Call struct {
	*mock.Call
}

// Private is a helper method to define mock.On call
func (_e *MockV3_Expecter) Private() *MockV3_Private_Call {
	return &MockV3_Private_Call{Call: _e.mock.On("Private")}
}

func (_c *MockV3_Private_Call) Run(run func()) *MockV3_Private_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockV3_Private_Call) Return(_a0 private.Sequencer) *MockV3_Private_Call {
	_c.Call.Return(_a0)
	return _c
}

// Query provides a mock function with given fields: ctx, scope, query
func (_m *MockV3) Query(ctx context.Context, scope *url.URL, query v3.Query) (v3.Record, error) {
	ret := _m.Called(ctx, scope, query)

	var r0 v3.Record
	if rf, ok := ret.Get(0).(func(context.Context, *url.URL, v3.Query) v3.Record); ok {
		r0 = rf(ctx, scope, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v3.Record)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *url.URL, v3.Query) error); ok {
		r1 = rf(ctx, scope, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockV3_Query_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Query'
type MockV3_Query_Call struct {
	*mock.Call
}

// Query is a helper method to define mock.On call
//   - ctx context.Context
//   - scope *url.URL
//   - query v3.Query
func (_e *MockV3_Expecter) Query(ctx interface{}, scope interface{}, query interface{}) *MockV3_Query_Call {
	return &MockV3_Query_Call{Call: _e.mock.On("Query", ctx, scope, query)}
}

func (_c *MockV3_Query_Call) Run(run func(ctx context.Context, scope *url.URL, query v3.Query)) *MockV3_Query_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*url.URL), args[2].(v3.Query))
	})
	return _c
}

func (_c *MockV3_Query_Call) Return(_a0 v3.Record, _a1 error) *MockV3_Query_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Submit provides a mock function with given fields: ctx, envelope, opts
func (_m *MockV3) Submit(ctx context.Context, envelope *messaging.Envelope, opts v3.SubmitOptions) ([]*v3.Submission, error) {
	ret := _m.Called(ctx, envelope, opts)

	var r0 []*v3.Submission
	if rf, ok := ret.Get(0).(func(context.Context, *messaging.Envelope, v3.SubmitOptions) []*v3.Submission); ok {
		r0 = rf(ctx, envelope, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v3.Submission)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *messaging.Envelope, v3.SubmitOptions) error); ok {
		r1 = rf(ctx, envelope, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockV3_Submit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Submit'
type MockV3_Submit_Call struct {
	*mock.Call
}

// Submit is a helper method to define mock.On call
//   - ctx context.Context
//   - envelope *messaging.Envelope
//   - opts v3.SubmitOptions
func (_e *MockV3_Expecter) Submit(ctx interface{}, envelope interface{}, opts interface{}) *MockV3_Submit_Call {
	return &MockV3_Submit_Call{Call: _e.mock.On("Submit", ctx, envelope, opts)}
}

func (_c *MockV3_Submit_Call) Run(run func(ctx context.Context, envelope *messaging.Envelope, opts v3.SubmitOptions)) *MockV3_Submit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*messaging.Envelope), args[2].(v3.SubmitOptions))
	})
	return _c
}

func (_c *MockV3_Submit_Call) Return(_a0 []*v3.Submission, _a1 error) *MockV3_Submit_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Validate provides a mock function with given fields: ctx, envelope, opts
func (_m *MockV3) Validate(ctx context.Context, envelope *messaging.Envelope, opts v3.ValidateOptions) ([]*v3.Submission, error) {
	ret := _m.Called(ctx, envelope, opts)

	var r0 []*v3.Submission
	if rf, ok := ret.Get(0).(func(context.Context, *messaging.Envelope, v3.ValidateOptions) []*v3.Submission); ok {
		r0 = rf(ctx, envelope, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v3.Submission)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *messaging.Envelope, v3.ValidateOptions) error); ok {
		r1 = rf(ctx, envelope, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockV3_Validate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Validate'
type MockV3_Validate_Call struct {
	*mock.Call
}

// Validate is a helper method to define mock.On call
//   - ctx context.Context
//   - envelope *messaging.Envelope
//   - opts v3.ValidateOptions
func (_e *MockV3_Expecter) Validate(ctx interface{}, envelope interface{}, opts interface{}) *MockV3_Validate_Call {
	return &MockV3_Validate_Call{Call: _e.mock.On("Validate", ctx, envelope, opts)}
}

func (_c *MockV3_Validate_Call) Run(run func(ctx context.Context, envelope *messaging.Envelope, opts v3.ValidateOptions)) *MockV3_Validate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*messaging.Envelope), args[2].(v3.ValidateOptions))
	})
	return _c
}

func (_c *MockV3_Validate_Call) Return(_a0 []*v3.Submission, _a1 error) *MockV3_Validate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewMockV3 interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockV3 creates a new instance of MockV3. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockV3(t mockConstructorTestingTNewMockV3) *MockV3 {
	mock := &MockV3{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
