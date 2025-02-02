// Code generated by mockery v2.14.0. DO NOT EDIT.

package p2p

import (
	context "context"
	io "io"

	peer "github.com/libp2p/go-libp2p/core/peer"
	mock "github.com/stretchr/testify/mock"
	api "gitlab.com/accumulatenetwork/accumulate/pkg/api/v3"
)

// mockDialerHost is an autogenerated mock type for the dialerHost type
type mockDialerHost struct {
	mock.Mock
}

type mockDialerHost_Expecter struct {
	mock *mock.Mock
}

func (_m *mockDialerHost) EXPECT() *mockDialerHost_Expecter {
	return &mockDialerHost_Expecter{mock: &_m.Mock}
}

// getOwnService provides a mock function with given fields: sa
func (_m *mockDialerHost) getOwnService(sa *api.ServiceAddress) (*service, bool) {
	ret := _m.Called(sa)

	var r0 *service
	if rf, ok := ret.Get(0).(func(*api.ServiceAddress) *service); ok {
		r0 = rf(sa)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(*api.ServiceAddress) bool); ok {
		r1 = rf(sa)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// mockDialerHost_getOwnService_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'getOwnService'
type mockDialerHost_getOwnService_Call struct {
	*mock.Call
}

// getOwnService is a helper method to define mock.On call
//   - sa *api.ServiceAddress
func (_e *mockDialerHost_Expecter) getOwnService(sa interface{}) *mockDialerHost_getOwnService_Call {
	return &mockDialerHost_getOwnService_Call{Call: _e.mock.On("getOwnService", sa)}
}

func (_c *mockDialerHost_getOwnService_Call) Run(run func(sa *api.ServiceAddress)) *mockDialerHost_getOwnService_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*api.ServiceAddress))
	})
	return _c
}

func (_c *mockDialerHost_getOwnService_Call) Return(_a0 *service, _a1 bool) *mockDialerHost_getOwnService_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// getPeerService provides a mock function with given fields: ctx, _a1, service
func (_m *mockDialerHost) getPeerService(ctx context.Context, _a1 peer.ID, service *api.ServiceAddress) (io.ReadWriteCloser, error) {
	ret := _m.Called(ctx, _a1, service)

	var r0 io.ReadWriteCloser
	if rf, ok := ret.Get(0).(func(context.Context, peer.ID, *api.ServiceAddress) io.ReadWriteCloser); ok {
		r0 = rf(ctx, _a1, service)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadWriteCloser)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, peer.ID, *api.ServiceAddress) error); ok {
		r1 = rf(ctx, _a1, service)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDialerHost_getPeerService_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'getPeerService'
type mockDialerHost_getPeerService_Call struct {
	*mock.Call
}

// getPeerService is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 peer.ID
//   - service *api.ServiceAddress
func (_e *mockDialerHost_Expecter) getPeerService(ctx interface{}, _a1 interface{}, service interface{}) *mockDialerHost_getPeerService_Call {
	return &mockDialerHost_getPeerService_Call{Call: _e.mock.On("getPeerService", ctx, _a1, service)}
}

func (_c *mockDialerHost_getPeerService_Call) Run(run func(ctx context.Context, _a1 peer.ID, service *api.ServiceAddress)) *mockDialerHost_getPeerService_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(peer.ID), args[2].(*api.ServiceAddress))
	})
	return _c
}

func (_c *mockDialerHost_getPeerService_Call) Return(_a0 io.ReadWriteCloser, _a1 error) *mockDialerHost_getPeerService_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// selfID provides a mock function with given fields:
func (_m *mockDialerHost) selfID() peer.ID {
	ret := _m.Called()

	var r0 peer.ID
	if rf, ok := ret.Get(0).(func() peer.ID); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(peer.ID)
	}

	return r0
}

// mockDialerHost_selfID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'selfID'
type mockDialerHost_selfID_Call struct {
	*mock.Call
}

// selfID is a helper method to define mock.On call
func (_e *mockDialerHost_Expecter) selfID() *mockDialerHost_selfID_Call {
	return &mockDialerHost_selfID_Call{Call: _e.mock.On("selfID")}
}

func (_c *mockDialerHost_selfID_Call) Run(run func()) *mockDialerHost_selfID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockDialerHost_selfID_Call) Return(_a0 peer.ID) *mockDialerHost_selfID_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTnewMockDialerHost interface {
	mock.TestingT
	Cleanup(func())
}

// newMockDialerHost creates a new instance of mockDialerHost. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockDialerHost(t mockConstructorTestingTnewMockDialerHost) *mockDialerHost {
	mock := &mockDialerHost{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
