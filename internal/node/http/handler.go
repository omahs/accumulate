// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/multiformats/go-multiaddr"
	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/accumulatenetwork/accumulate/internal/api/p2p"
	"gitlab.com/accumulatenetwork/accumulate/internal/api/routing"
	v2 "gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/internal/node/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/node/web"
	"gitlab.com/accumulatenetwork/accumulate/pkg/api/v3"
	"gitlab.com/accumulatenetwork/accumulate/pkg/api/v3/jsonrpc"
	"gitlab.com/accumulatenetwork/accumulate/pkg/api/v3/message"
	"gitlab.com/accumulatenetwork/accumulate/pkg/api/v3/websocket"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
)

// Handler processes API requests.
type Handler struct {
	logger logging.OptionalLogger
	mux    *http.ServeMux
}

// Options are the options for a [Handler].
type Options struct {
	Logger log.Logger
	Node   *p2p.Node
	Router routing.Router

	// For API v2
	Network *config.Describe
	MaxWait time.Duration
}

// NewHandler returns a new Handler.
func NewHandler(opts Options) (*Handler, error) {
	h := new(Handler)
	h.logger.Set(opts.Logger)

	// Message clients
	selfClient := &message.Client{
		Router: unrouter(opts.Network.PartitionId),
		Dialer: opts.Node.SelfDialer(),
	}

	client := &message.Client{
		Router: routing.MessageRouter{Router: opts.Router},
		Dialer: opts.Node.Dialer(),
	}

	// JSON-RPC API v3
	v3, err := jsonrpc.NewHandler(
		opts.Logger,
		jsonrpc.NodeService{NodeService: selfClient},
		jsonrpc.NetworkService{NetworkService: client},
		jsonrpc.MetricsService{MetricsService: client},
		jsonrpc.Querier{Querier: client},
		jsonrpc.Submitter{Submitter: client},
		jsonrpc.Validator{Validator: client},
		jsonrpc.Faucet{Faucet: client},
	)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("initialize API v3: %w", err)
	}

	// WebSocket API v3
	ws, err := websocket.NewHandler(
		opts.Logger,
		message.NodeService{NodeService: selfClient},
		message.NetworkService{NetworkService: client},
		message.MetricsService{MetricsService: client},
		message.Querier{Querier: client},
		message.Submitter{Submitter: client},
		message.Validator{Validator: client},
		message.Faucet{Faucet: client},
		message.EventService{EventService: client},
	)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("initialize websocket API: %w", err)
	}

	v2, err := v2.NewJrpc(v2.Options{
		Logger:        opts.Logger,
		Describe:      opts.Network,
		TxMaxWaitTime: opts.MaxWait,
		LocalV3:       selfClient,
		NetV3:         client,
	})
	if err != nil {
		return nil, errors.UnknownError.WithFormat("initialize API v2: %v", err)
	}

	// Set up mux
	h.mux = v2.NewMux()
	h.mux.Handle("/v3", ws.FallbackTo(v3))

	h.mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/x")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	webex := web.Handler()
	h.mux.HandleFunc("/x/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/x")
		r.RequestURI = strings.TrimPrefix(r.RequestURI, "/x")
		webex.ServeHTTP(w, r)
	})

	return h, nil
}

// ServeHTTP implements [http.Handler].
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

type unrouter string

func (r unrouter) Route(msg message.Message) (multiaddr.Multiaddr, error) {
	service := new(api.ServiceAddress)
	service.Partition = string(r)
	var err error
	switch msg := msg.(type) {
	case *message.NetworkStatusRequest:
		service.Type = api.ServiceTypeNetwork
	case *message.NodeStatusRequest:
		service.Type = api.ServiceTypeNode
	case *message.MetricsRequest:
		service.Type = api.ServiceTypeMetrics
	case *message.QueryRequest:
		service.Type = api.ServiceTypeQuery
	case *message.SubmitRequest:
		service.Type = api.ServiceTypeSubmit
	case *message.ValidateRequest:
		service.Type = api.ServiceTypeValidate
	case *message.SubscribeRequest:
		service.Type = api.ServiceTypeEvent
	case *message.FaucetRequest:
		service.Type = api.ServiceTypeFaucet
	default:
		return nil, errors.BadRequest.WithFormat("%v is not routable", msg.Type())
	}
	if err != nil {
		return nil, errors.BadRequest.WithFormat("cannot route request: %w", err)
	}

	// Return /acc/{service}:{partition}
	ma, err := multiaddr.NewComponent(api.N_ACC, service.String())
	if err != nil {
		return nil, errors.BadRequest.WithFormat("build multiaddr: %w", err)
	}
	return ma, nil
}
