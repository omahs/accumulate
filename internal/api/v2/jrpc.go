package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	stdlog "log"
	"mime"
	"net/http"
	"os"

	"github.com/AccumulateNetwork/jsonrpc2/v15"
	"github.com/go-playground/validator/v10"
	"github.com/tendermint/tendermint/libs/log"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/accumulatenetwork/accumulate"
	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/core"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/networks"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type JrpcMethods struct {
	Options
	querier  *queryDispatch
	methods  jsonrpc2.MethodMap
	validate *validator.Validate
	logger   log.Logger
}

func NewJrpc(opts Options) (*JrpcMethods, error) {
	var err error
	m := new(JrpcMethods)
	m.Options = opts
	m.querier = new(queryDispatch)
	m.querier.Options = opts

	if opts.Logger != nil {
		m.logger = opts.Logger.With("module", "jrpc")
	}

	m.validate, err = protocol.NewValidator()
	if err != nil {
		return nil, err
	}

	m.populateMethodTable()
	return m, nil
}

func (m *JrpcMethods) Querier_TESTONLY() Querier {
	return m.querier
}

func (m *JrpcMethods) logError(msg string, keyVals ...interface{}) {
	if m.logger != nil {
		m.logger.Error(msg, keyVals...)
	}
}

func (m *JrpcMethods) EnableDebug() {
	q := m.querier.direct(m.Network.LocalSubnetID)
	m.methods["debug-query-direct"] = func(_ context.Context, params json.RawMessage) interface{} {
		req := new(GeneralQuery)
		err := m.parse(params, req)
		if err != nil {
			return err
		}

		return jrpcFormatResponse(q.QueryUrl(req.Url, req.QueryOptions))
	}
}

func (m *JrpcMethods) NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/status", m.jrpc2http(m.Status))
	mux.Handle("/version", m.jrpc2http(m.Version))
	mux.Handle("/describe", m.jrpc2http(m.Describe))
	mux.Handle("/peers", m.jrpc2http(m.Peers))
	mux.Handle("/v2", jsonrpc2.HTTPRequestHandler(m.methods, stdlog.New(os.Stdout, "", 0)))
	return mux
}

func (m *JrpcMethods) jrpc2http(jrpc jsonrpc2.MethodFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		var params json.RawMessage
		mediatype, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if mediatype == "application/json" || mediatype == "text/json" {
			params = body
		}

		r := jrpc(req.Context(), params)
		res.Header().Add("Content-Type", "application/json")
		data, err := json.Marshal(r)
		if err != nil {
			m.logError("Failed to marshal status", "error", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = res.Write(data)
	}
}

func (m *JrpcMethods) Status(_ context.Context, params json.RawMessage) interface{} {
	return &StatusResponse{
		Ok: true,
	}
}

func (m *JrpcMethods) Version(_ context.Context, params json.RawMessage) interface{} {
	res := new(ChainQueryResponse)
	res.Type = "version"
	res.Data = VersionResponse{
		Version:        accumulate.Version,
		Commit:         accumulate.Commit,
		VersionIsKnown: accumulate.IsVersionKnown(),
	}
	return res
}

func (m *JrpcMethods) Describe(_ context.Context, params json.RawMessage) interface{} {
	res := new(DescriptionResponse)
	res.Network = m.Network

	// Load network variable values
	res.Values = new(core.GlobalValues)
	err := res.Values.Load(m.Network, func(account *url.URL, target interface{}) error {
		return m.Database.View(func(batch *database.Batch) error {
			return batch.Account(account).GetStateAs(target)
		})
	})
	if err != nil {
		res.Error = errors.Wrap(errors.StatusUnknown, err).(*errors.Error)
	}

	return res
}

func (m *JrpcMethods) Peers(ctx context.Context, _ json.RawMessage) interface{} {
	connctx, err := m.ConnectionManager.SelectConnection(protocol.Directory, true)
	if err != nil {
		return accumulateError(err)
	}

	u, err := config.OffsetPort(connctx.GetAddress(), networks.TmRpcPortOffset)
	if err != nil {
		return accumulateError(err)
	}
	nodes := []string{u.String()}

	info, err := connctx.GetABCIClient().NetInfo(ctx)
	if err != nil {
		return accumulateError(err)
	}
	for _, peer := range info.Peers {
		u, err := config.OffsetPort(peer.URL, networks.TmRpcPortOffset)
		if err != nil {
			return accumulateError(err)
		}
		u.Scheme = "http"
		u.User = nil
		nodes = append(nodes, u.String())
	}

	peers := map[[32]byte]string{}
	for _, addr := range nodes {
		client, err := rpchttp.New(addr)
		if err != nil {
			return accumulateError(err)
		}

		status, err := client.Status(ctx)
		if err != nil {
			return accumulateError(err)
		}

		hash := sha256.Sum256(status.ValidatorInfo.PubKey.Bytes())
		peers[hash] = addr
	}

	globals := new(core.GlobalValues)
	err = globals.Load(m.Network, func(account *url.URL, target interface{}) error {
		return m.Database.View(func(batch *database.Batch) error {
			return batch.Account(account).GetStateAs(target)
		})
	})
	if err != nil {
		return accumulateError(err)
	}

	type Validator struct {
		KeyHash string
		Host    string
	}
	type Subnet struct {
		ID         string
		Validators []*Validator
	}

	var subnets []*Subnet
	for _, subnet := range globals.Network.Subnets {
		vals := new(Subnet)
		vals.ID = subnet.SubnetID
		subnets = append(subnets, vals)

		for _, hash := range subnet.ValidatorKeyHashes {
			peer, ok := peers[hash]
			if !ok {
				continue
			}

			if subnet.SubnetID != protocol.Directory {
				u, err := config.OffsetPort(peer, -networks.DnnPortOffset)
				if err != nil {
					return accumulateError(err)
				}
				peer = u.String()
			}

			vals.Validators = append(vals.Validators, &Validator{
				KeyHash: hex.EncodeToString(hash[:]), //nolint:rangevarref
				Host:    peer,
			})
		}
	}

	return subnets
}
