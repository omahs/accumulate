// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package node

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"net/url"

	abciclient "github.com/tendermint/tendermint/abci/client"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/privval"
	coretypes "github.com/tendermint/tendermint/rpc/coretypes"
	corerpc "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"gitlab.com/accumulatenetwork/accumulate/config"
	web "gitlab.com/accumulatenetwork/accumulate/internal/web/static"
)

// AppFactory creates and returns an ABCI application.
type AppFactory func(*privval.FilePV) (abci.Application, error)

// Node wraps a Tendermint node.
type Node struct {
	service.Service
	Config *config.Config
	ABCI   abci.Application
	logger log.Logger
}

// New initializes a Tendermint node for the given ABCI application.
func New(config *config.Config, app abci.Application, logger log.Logger) (*Node, error) {
	node := new(Node)
	node.Config = config
	node.ABCI = app
	node.logger = logger

	// create node
	var err error
	node.Service, err = nm.New(&config.Config, logger, abciclient.NewLocalCreator(app), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new Tendermint node: %w", err)
	}

	return node, nil
}

// Start starts the Tendermint node.
func (n *Node) Start() error {
	err := n.Service.Start()
	if err != nil {
		return err
	}
	if n.Config.Accumulate.Website.Enabled {
		u, err := url.Parse(n.Config.Accumulate.Website.ListenAddress)
		if err != nil {
			return fmt.Errorf("invalid website listen address: %v", err)
		}
		if u.Scheme != "http" {
			return fmt.Errorf("invalid website listen address: expected scheme http, got %q", u.Scheme)
		}

		website := http.Server{Addr: u.Host, Handler: http.FileServer(http.FS(web.FS))}
		go func() {
			<-n.Quit()
			_ = website.Shutdown(context.Background())
		}()
		go func() {
			n.logger.Info("Listening", "host", u.Host, "module", "website")
			err := website.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				stdlog.Fatalf("Failed to start website: %v", err)
			}
		}()
	}
	_, err = n.waitForRPC()
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) waitForRPC() (*corerpc.Client, error) {
	client, err := corerpc.New(n.Config.RPC.ListenAddress)
	if err != nil {
		return nil, err
	}
	result := new(*coretypes.ResultStatus)
	for {
		_, err := client.Call(context.Background(), "status", nil, &result)
		if err == nil {
			return client, nil
		}
	}

}
