// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package connections

type FakeConnectionManager interface {
	ConnectionManager
	SetClients(clients map[string]FakeClient)
}

type fakeConnectionManager struct {
	ctxMap map[string]ConnectionContext
}

func (fcm *fakeConnectionManager) SelectConnection(partitionId string, allowFollower bool) (ConnectionContext, error) {
	context, ok := fcm.ctxMap[partitionId]
	if ok {
		return context, nil
	} else {
		return nil, errUnknownPartition(partitionId)
	}
}

type FakeClient = struct {
	ABCI ABCIClient
	API  APIClient
}

func NewFakeConnectionManager(partitions []string) FakeConnectionManager {
	fcm := new(fakeConnectionManager)
	fcm.ctxMap = make(map[string]ConnectionContext)
	for _, partition := range partitions {
		connCtx := &connectionContext{
			partitionId: partition,
			hasClient:   make(chan struct{}),
			metrics:     NodeMetrics{status: Up}}
		fcm.ctxMap[partition] = connCtx
	}
	return fcm
}

func (fcm *fakeConnectionManager) SetClients(clients map[string]FakeClient) {
	for partition, client := range clients {
		fcm.ctxMap[partition].(*connectionContext).setClient(client.ABCI, client.API)
	}
}

func (fcm *fakeConnectionManager) GetBVNContextMap() map[string][]ConnectionContext {
	return nil
}

func (fcm *fakeConnectionManager) GetDNContextList() []ConnectionContext {
	return nil
}

func (fcm *fakeConnectionManager) GetFNContextList() []ConnectionContext {
	return nil
}

func (fcm *fakeConnectionManager) GetAllNodeContexts() []ConnectionContext {
	return nil
}

func (fcm *fakeConnectionManager) GetLocalNodeContext() ConnectionContext {
	return nil
}

func (fcm *fakeConnectionManager) ResetErrors() {
}
