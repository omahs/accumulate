// Code generated by MockGen. DO NOT EDIT.
// Source: abci.go

// Package mock_abci is a generated GoMock package.
package mock_abci

import (
	reflect "reflect"
	time "time"

	abci "github.com/AccumulateNetwork/accumulate/internal/abci"
	protocol "github.com/AccumulateNetwork/accumulate/protocol"
	query "github.com/AccumulateNetwork/accumulate/types/api/query"
	transactions "github.com/AccumulateNetwork/accumulate/types/api/transactions"
	gomock "github.com/golang/mock/gomock"
)

// MockChain is a mock of Chain interface.
type MockChain struct {
	ctrl     *gomock.Controller
	recorder *MockChainMockRecorder
}

// MockChainMockRecorder is the mock recorder for MockChain.
type MockChainMockRecorder struct {
	mock *MockChain
}

// NewMockChain creates a new mock instance.
func NewMockChain(ctrl *gomock.Controller) *MockChain {
	mock := &MockChain{ctrl: ctrl}
	mock.recorder = &MockChainMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChain) EXPECT() *MockChainMockRecorder {
	return m.recorder
}

// BeginBlock mocks base method.
func (m *MockChain) BeginBlock(arg0 abci.BeginBlockRequest) (abci.BeginBlockResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginBlock", arg0)
	ret0, _ := ret[0].(abci.BeginBlockResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginBlock indicates an expected call of BeginBlock.
func (mr *MockChainMockRecorder) BeginBlock(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginBlock", reflect.TypeOf((*MockChain)(nil).BeginBlock), arg0)
}

// CheckTx mocks base method.
func (m *MockChain) CheckTx(arg0 *transactions.GenTransaction) *protocol.Error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckTx", arg0)
	ret0, _ := ret[0].(*protocol.Error)
	return ret0
}

// CheckTx indicates an expected call of CheckTx.
func (mr *MockChainMockRecorder) CheckTx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckTx", reflect.TypeOf((*MockChain)(nil).CheckTx), arg0)
}

// Commit mocks base method.
func (m *MockChain) Commit() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Commit indicates an expected call of Commit.
func (mr *MockChainMockRecorder) Commit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockChain)(nil).Commit))
}

// DeliverTx mocks base method.
func (m *MockChain) DeliverTx(arg0 *transactions.GenTransaction) *protocol.Error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeliverTx", arg0)
	ret0, _ := ret[0].(*protocol.Error)
	return ret0
}

// DeliverTx indicates an expected call of DeliverTx.
func (mr *MockChainMockRecorder) DeliverTx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeliverTx", reflect.TypeOf((*MockChain)(nil).DeliverTx), arg0)
}

// EndBlock mocks base method.
func (m *MockChain) EndBlock(arg0 abci.EndBlockRequest) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "EndBlock", arg0)
}

// EndBlock indicates an expected call of EndBlock.
func (mr *MockChainMockRecorder) EndBlock(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EndBlock", reflect.TypeOf((*MockChain)(nil).EndBlock), arg0)
}

// InitChain mocks base method.
func (m *MockChain) InitChain(state []byte, time time.Time, blockIndex int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitChain", state, time, blockIndex)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitChain indicates an expected call of InitChain.
func (mr *MockChainMockRecorder) InitChain(state, time, blockIndex interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitChain", reflect.TypeOf((*MockChain)(nil).InitChain), state, time, blockIndex)
}

// Query mocks base method.
func (m *MockChain) Query(arg0 *query.Query) ([]byte, []byte, *protocol.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(*protocol.Error)
	return ret0, ret1, ret2
}

// Query indicates an expected call of Query.
func (mr *MockChainMockRecorder) Query(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockChain)(nil).Query), arg0)
}
