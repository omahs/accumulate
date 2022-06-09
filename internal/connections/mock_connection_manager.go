// Code generated by MockGen. DO NOT EDIT.
// Source: connection_manager.go

// Package connections is a generated GoMock package.
package connections

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	local "github.com/tendermint/tendermint/rpc/client/local"
)

// MockConnectionManager is a mock of ConnectionManager interface.
type MockConnectionManager struct {
	ctrl     *gomock.Controller
	recorder *MockConnectionManagerMockRecorder
}

// MockConnectionManagerMockRecorder is the mock recorder for MockConnectionManager.
type MockConnectionManagerMockRecorder struct {
	mock *MockConnectionManager
}

// NewMockConnectionManager creates a new mock instance.
func NewMockConnectionManager(ctrl *gomock.Controller) *MockConnectionManager {
	mock := &MockConnectionManager{ctrl: ctrl}
	mock.recorder = &MockConnectionManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnectionManager) EXPECT() *MockConnectionManagerMockRecorder {
	return m.recorder
}

// GetLocalNodeContext mocks base method.
func (m *MockConnectionManager) GetLocalNodeContext() ConnectionContext {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLocalNodeContext")
	ret0, _ := ret[0].(ConnectionContext)
	return ret0
}

// GetLocalNodeContext indicates an expected call of GetLocalNodeContext.
func (mr *MockConnectionManagerMockRecorder) GetLocalNodeContext() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLocalNodeContext", reflect.TypeOf((*MockConnectionManager)(nil).GetLocalNodeContext))
}

// SelectConnection mocks base method.
func (m *MockConnectionManager) SelectConnection(subnetId string, allowFollower bool) (ConnectionContext, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectConnection", subnetId, allowFollower)
	ret0, _ := ret[0].(ConnectionContext)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectConnection indicates an expected call of SelectConnection.
func (mr *MockConnectionManagerMockRecorder) SelectConnection(subnetId, allowFollower interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectConnection", reflect.TypeOf((*MockConnectionManager)(nil).SelectConnection), subnetId, allowFollower)
}

// MockConnectionInitializer is a mock of ConnectionInitializer interface.
type MockConnectionInitializer struct {
	ctrl     *gomock.Controller
	recorder *MockConnectionInitializerMockRecorder
}

// MockConnectionInitializerMockRecorder is the mock recorder for MockConnectionInitializer.
type MockConnectionInitializerMockRecorder struct {
	mock *MockConnectionInitializer
}

// NewMockConnectionInitializer creates a new mock instance.
func NewMockConnectionInitializer(ctrl *gomock.Controller) *MockConnectionInitializer {
	mock := &MockConnectionInitializer{ctrl: ctrl}
	mock.recorder = &MockConnectionInitializerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnectionInitializer) EXPECT() *MockConnectionInitializerMockRecorder {
	return m.recorder
}

// ConnectDirectly mocks base method.
func (m *MockConnectionInitializer) ConnectDirectly(other ConnectionManager) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectDirectly", other)
	ret0, _ := ret[0].(error)
	return ret0
}

// ConnectDirectly indicates an expected call of ConnectDirectly.
func (mr *MockConnectionInitializerMockRecorder) ConnectDirectly(other interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectDirectly", reflect.TypeOf((*MockConnectionInitializer)(nil).ConnectDirectly), other)
}

// GetLocalNodeContext mocks base method.
func (m *MockConnectionInitializer) GetLocalNodeContext() ConnectionContext {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLocalNodeContext")
	ret0, _ := ret[0].(ConnectionContext)
	return ret0
}

// GetLocalNodeContext indicates an expected call of GetLocalNodeContext.
func (mr *MockConnectionInitializerMockRecorder) GetLocalNodeContext() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLocalNodeContext", reflect.TypeOf((*MockConnectionInitializer)(nil).GetLocalNodeContext))
}

// InitClients mocks base method.
func (m *MockConnectionInitializer) InitClients(arg0 *local.Local, arg1 StatusChecker) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitClients", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitClients indicates an expected call of InitClients.
func (mr *MockConnectionInitializerMockRecorder) InitClients(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitClients", reflect.TypeOf((*MockConnectionInitializer)(nil).InitClients), arg0, arg1)
}

// SelectConnection mocks base method.
func (m *MockConnectionInitializer) SelectConnection(subnetId string, allowFollower bool) (ConnectionContext, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectConnection", subnetId, allowFollower)
	ret0, _ := ret[0].(ConnectionContext)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectConnection indicates an expected call of SelectConnection.
func (mr *MockConnectionInitializerMockRecorder) SelectConnection(subnetId, allowFollower interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectConnection", reflect.TypeOf((*MockConnectionInitializer)(nil).SelectConnection), subnetId, allowFollower)
}
