package connections

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/rpc/client"
	core "github.com/tendermint/tendermint/rpc/coretypes"
	tm "github.com/tendermint/tendermint/types"
	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type NodeStatus int

const (
	Up           NodeStatus = iota // Healthy & ready to go
	Down                           // Not reachable
	OutOfService                   // Reachable but not ready to go (IE. still syncing up)
	Unknown                        // Not checked yet
)

type NetworkGroup int

const (
	Local NetworkGroup = iota
	SamePartition
	OtherPartition
)

// ABCIClient is a subset of from TM/rpc/client.ABCIClient.
type ABCIClient interface {
	ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts client.ABCIQueryOptions) (*core.ResultABCIQuery, error)
	Status(context.Context) (*core.ResultStatus, error)
	CheckTx(ctx context.Context, tx tm.Tx) (*core.ResultCheckTx, error)
	BroadcastTxAsync(context.Context, tm.Tx) (*core.ResultBroadcastTx, error)
	BroadcastTxSync(context.Context, tm.Tx) (*core.ResultBroadcastTx, error)
}

type APIClient interface {
	RequestAPIv2(_ context.Context, method string, params, result interface{}) error
}

type ConnectionContext interface {
	GetNetworkGroup() NetworkGroup
	GetNodeType() config.NodeType
	GetPublicKey() []byte
	GetMetrics() *NodeMetrics
	GetAddress() *protocol.InternetAddress
	GetABCIClient() ABCIClient
	GetAPIClient() APIClient
	IsHealthy() bool
	ReportError(err error)
	ReportErrorStatus(status NodeStatus)
	ClearErrors()
}

type StatusChecker interface {
	IsStatusOk(connCtx ConnectionContext) bool
}

type connectionContext struct {
	partition           *protocol.PartitionDefinition
	validator           *protocol.ValidatorDefinition
	address             *protocol.InternetAddress
	nodeType            config.NodeType
	abciClient          ABCIClient
	apiClient           APIClient
	hasClient           chan struct{}
	connMgr             *connectionManager
	statusChecker       StatusChecker
	networkGroup        NetworkGroup
	resolvedIPs         []net.IP
	metrics             NodeMetrics
	lastErrorExpiryTime time.Time
}

func (cc *connectionContext) GetAddress() *protocol.InternetAddress {
	return cc.address
}

func (cc *connectionContext) GetPublicKey() []byte {
	return cc.validator.PublicKey
}

func (cc *connectionContext) GetABCIClient() ABCIClient {
	c, _ := cc.getClients()
	return c
}

func (cc *connectionContext) GetAPIClient() APIClient {
	_, c := cc.getClients()
	return c
}

func (cc *connectionContext) getClients() (ABCIClient, APIClient) {
	if cc.abciClient != nil {
		return cc.abciClient, cc.apiClient
	}

	// Client not there yet? Wait for it.
	timeout := time.After(10 * time.Second)
	select {
	case <-cc.hasClient:
		return cc.abciClient, cc.apiClient
	case <-timeout:
		panic(fmt.Sprintf("Could not obtain a client for node %s  ", cc.address))
	}
}

func (cc *connectionContext) GetNodeType() config.NodeType {
	return cc.nodeType
}

func (cc *connectionContext) GetNetworkGroup() NetworkGroup {
	return cc.networkGroup
}

func (cc *connectionContext) IsHealthy() bool {
	switch cc.metrics.status {
	case Up:
		return true
	case Unknown:
		cc.connMgr.doHealthCheckOnNode(cc)
		if cc.metrics.status == Up {
			return true
		}
	default:
		now := time.Now()
		if now.After(cc.lastErrorExpiryTime) {
			cc.lastErrorExpiryTime = now.Add(UnhealthyNodeCheckInterval) // avoid double doHealthCheckOnNode calls
			go cc.connMgr.doHealthCheckOnNode(cc)
		}
	}
	return false
}

func (cc *connectionContext) ClearErrors() {
	cc.lastErrorExpiryTime = time.Now()
}

func (cc *connectionContext) GetMetrics() *NodeMetrics {
	return &cc.metrics
}

func (cc *connectionContext) ReportError(err error) {
	// TODO Maybe we need to filter out certain errors, those should not mark the node as being out of service
	cc.lastErrorExpiryTime = time.Now().Add(UnhealthyNodeCheckInterval)
	cc.metrics.status = OutOfService

	// Set node status to Down for dial-up errors
	switch root := err.(type) {
	case *url.Error:
		if root.Err != nil {
			switch cause := root.Err.(type) {
			case *net.OpError:
				if cause.Op == "dial" {
					cc.metrics.status = Down
				}
			}
		}
	}
}

func (cc *connectionContext) ReportErrorStatus(status NodeStatus) {
	cc.metrics.status = status
	cc.lastErrorExpiryTime = time.Now().Add(UnhealthyNodeCheckInterval)
}

func (cc *connectionContext) setClient(abci ABCIClient, api APIClient) {
	shouldClose := cc.abciClient == nil
	cc.abciClient = abci
	cc.apiClient = api
	if shouldClose {
		close(cc.hasClient)
	}
}
