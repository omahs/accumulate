package client

// GENERATED BY go run ./tools/cmd/gen-sdk. DO NOT EDIT.

import (
	"context"

	"gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

// Describe queries the basic configuration of the node.
func (c *Client) Describe(ctx context.Context) (*api.DescriptionResponse, error) {
	var req struct{}
	var resp api.DescriptionResponse

	err := c.RequestAPIv2(ctx, "describe", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Execute submits a transaction.
func (c *Client) Execute(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "execute", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteAddCredits submits an AddCredits transaction.
func (c *Client) ExecuteAddCredits(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "add-credits", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteAddValidator submits an AddValidator transaction.
func (c *Client) ExecuteAddValidator(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "add-validator", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteBurnTokens submits a BurnTokens transaction.
func (c *Client) ExecuteBurnTokens(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "burn-tokens", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteCreateAdi submits a CreateIdentity transaction.
func (c *Client) ExecuteCreateAdi(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "create-adi", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteCreateDataAccount submits a CreateDataAccount transaction.
func (c *Client) ExecuteCreateDataAccount(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "create-data-account", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteCreateIdentity submits a CreateIdentity transaction.
func (c *Client) ExecuteCreateIdentity(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "create-identity", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteCreateKeyBook submits a CreateKeyBook transaction.
func (c *Client) ExecuteCreateKeyBook(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "create-key-book", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteCreateKeyPage submits a CreateKeyPage transaction.
func (c *Client) ExecuteCreateKeyPage(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "create-key-page", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteCreateToken submits a CreateToken transaction.
func (c *Client) ExecuteCreateToken(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "create-token", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteCreateTokenAccount submits a CreateTokenAccount transaction.
func (c *Client) ExecuteCreateTokenAccount(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "create-token-account", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteIssueTokens submits an IssueTokens transaction.
func (c *Client) ExecuteIssueTokens(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "issue-tokens", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteRemoveValidator submits a RemoveValidator transaction.
func (c *Client) ExecuteRemoveValidator(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "remove-validator", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteSendTokens submits a SendTokens transaction.
func (c *Client) ExecuteSendTokens(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "send-tokens", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteUpdateAccountAuth submits an UpdateAccountAuth transaction.
func (c *Client) ExecuteUpdateAccountAuth(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "update-account-auth", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteUpdateKey submits an UpdateKey transaction.
func (c *Client) ExecuteUpdateKey(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "update-key", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteUpdateKeyPage submits an UpdateKeyPage transaction.
func (c *Client) ExecuteUpdateKeyPage(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "update-key-page", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteUpdateValidatorKey submits an UpdateValidatorKey transaction.
func (c *Client) ExecuteUpdateValidatorKey(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "update-validator-key", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteWriteData submits a WriteData transaction.
func (c *Client) ExecuteWriteData(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "write-data", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteWriteDataTo submits a WriteDataTo transaction.
func (c *Client) ExecuteWriteDataTo(ctx context.Context, req *api.TxRequest) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "write-data-to", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Faucet requests tokens from the ACME faucet.
func (c *Client) Faucet(ctx context.Context, req *protocol.AcmeFaucet) (*api.TxResponse, error) {
	var resp api.TxResponse

	err := c.RequestAPIv2(ctx, "faucet", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Metrics queries network metrics, such as transactions per second.
func (c *Client) Metrics(ctx context.Context, req *api.MetricsQuery) (*api.ChainQueryResponse, error) {
	var resp api.ChainQueryResponse

	err := c.RequestAPIv2(ctx, "metrics", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Query queries an account or account chain by URL.
func (c *Client) Query(ctx context.Context, req *api.GeneralQuery) (interface{}, error) {
	var resp interface{}

	err := c.RequestAPIv2(ctx, "query", req, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// QueryChain queries an account by ID.
func (c *Client) QueryChain(ctx context.Context, req *api.ChainIdQuery) (*api.ChainQueryResponse, error) {
	var resp api.ChainQueryResponse

	err := c.RequestAPIv2(ctx, "query-chain", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// QueryData queries an entry on an account's data chain.
func (c *Client) QueryData(ctx context.Context, req *api.DataEntryQuery) (*api.ChainQueryResponse, error) {
	var resp api.ChainQueryResponse

	err := c.RequestAPIv2(ctx, "query-data", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// QueryDataSet queries a range of entries on an account's data chain.
func (c *Client) QueryDataSet(ctx context.Context, req *api.DataEntrySetQuery) (*api.MultiResponse, error) {
	var resp api.MultiResponse

	err := c.RequestAPIv2(ctx, "query-data-set", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// QueryDirectory queries the directory entries of an account.
func (c *Client) QueryDirectory(ctx context.Context, req *api.DirectoryQuery) (*api.MultiResponse, error) {
	var resp api.MultiResponse

	err := c.RequestAPIv2(ctx, "query-directory", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// QueryKeyPageIndex queries the location of a key within an account's key book(s).
func (c *Client) QueryKeyPageIndex(ctx context.Context, req *api.KeyPageIndexQuery) (*api.ChainQueryResponse, error) {
	var resp api.ChainQueryResponse

	err := c.RequestAPIv2(ctx, "query-key-index", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// QueryTx queries a transaction by ID.
func (c *Client) QueryTx(ctx context.Context, req *api.TxnQuery) (*api.TransactionQueryResponse, error) {
	var resp api.TransactionQueryResponse

	err := c.RequestAPIv2(ctx, "query-tx", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// QueryTxHistory queries an account's transaction history.
func (c *Client) QueryTxHistory(ctx context.Context, req *api.TxHistoryQuery) (*api.MultiResponse, error) {
	var resp api.MultiResponse

	err := c.RequestAPIv2(ctx, "query-tx-history", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Status queries the status of the node.
func (c *Client) Status(ctx context.Context) (*api.StatusResponse, error) {
	var req struct{}
	var resp api.StatusResponse

	err := c.RequestAPIv2(ctx, "status", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Version queries the software version of the node.
func (c *Client) Version(ctx context.Context) (*api.ChainQueryResponse, error) {
	var req struct{}
	var resp api.ChainQueryResponse

	err := c.RequestAPIv2(ctx, "version", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
