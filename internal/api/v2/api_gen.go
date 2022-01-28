package api

import (
	"context"
	"encoding/json"

	"github.com/AccumulateNetwork/accumulate/protocol"
	"github.com/AccumulateNetwork/jsonrpc2/v15"
)

func (m *JrpcMethods) populateMethodTable() jsonrpc2.MethodMap {
	if m.methods == nil {
		m.methods = make(jsonrpc2.MethodMap, 28)
	}

	m.methods["describe"] = m.Describe
	m.methods["execute"] = m.Execute
	m.methods["add-credits"] = m.ExecuteAddCredits
	m.methods["burn-tokens"] = m.ExecuteBurnTokens
	m.methods["create-adi"] = m.ExecuteCreateAdi
	m.methods["create-data-account"] = m.ExecuteCreateDataAccount
	m.methods["create-key-book"] = m.ExecuteCreateKeyBook
	m.methods["create-key-page"] = m.ExecuteCreateKeyPage
	m.methods["create-token"] = m.ExecuteCreateToken
	m.methods["create-token-account"] = m.ExecuteCreateTokenAccount
	m.methods["issue-tokens"] = m.ExecuteIssueTokens
	m.methods["send-tokens"] = m.ExecuteSendTokens
	m.methods["update-key-page"] = m.ExecuteUpdateKeyPage
	m.methods["write-data"] = m.ExecuteWriteData
	m.methods["write-data-to"] = m.ExecuteWriteDataTo
	m.methods["faucet"] = m.Faucet
	m.methods["metrics"] = m.Metrics
	m.methods["query"] = m.Query
	m.methods["query-chain"] = m.QueryChain
	m.methods["query-data"] = m.QueryData
	m.methods["query-data-set"] = m.QueryDataSet
	m.methods["query-directory"] = m.QueryDirectory
	m.methods["query-key-index"] = m.QueryKeyPageIndex
	m.methods["query-pending-tx"] = m.QueryPending
	m.methods["query-pending-txs"] = m.QueryPendingPagination
	m.methods["query-tx"] = m.QueryTx
	m.methods["query-tx-history"] = m.QueryTxHistory
	m.methods["version"] = m.Version

	return m.methods
}

func (m *JrpcMethods) parse(params json.RawMessage, target interface{}, validateFields ...string) error {
	err := json.Unmarshal(params, target)
	if err != nil {
		return validatorError(err)
	}

	// validate fields
	if len(validateFields) == 0 {
		if err = m.validate.Struct(target); err != nil {
			return validatorError(err)
		}
	} else {
		if err = m.validate.StructPartial(target, validateFields...); err != nil {
			return validatorError(err)
		}
	}

	return nil
}

func jrpcFormatResponse(res interface{}, err error) interface{} {
	if err != nil {
		return accumulateError(err)
	}

	return res
}

func (m *JrpcMethods) ExecuteAddCredits(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.AddCredits))
}

func (m *JrpcMethods) ExecuteBurnTokens(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.BurnTokens))
}

func (m *JrpcMethods) ExecuteCreateAdi(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateIdentity))
}

func (m *JrpcMethods) ExecuteCreateDataAccount(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateDataAccount))
}

func (m *JrpcMethods) ExecuteCreateKeyBook(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateKeyBook))
}

func (m *JrpcMethods) ExecuteCreateKeyPage(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateKeyPage))
}

func (m *JrpcMethods) ExecuteCreateToken(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateToken))
}

func (m *JrpcMethods) ExecuteCreateTokenAccount(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateTokenAccount))
}

func (m *JrpcMethods) ExecuteIssueTokens(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.IssueTokens))
}

func (m *JrpcMethods) ExecuteSendTokens(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.SendTokens), "From", "To")
}

func (m *JrpcMethods) ExecuteUpdateKeyPage(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.UpdateKeyPage))
}

func (m *JrpcMethods) ExecuteWriteData(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.WriteData))
}

func (m *JrpcMethods) ExecuteWriteDataTo(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.WriteDataTo))
}

func (m *JrpcMethods) Query(_ context.Context, params json.RawMessage) interface{} {
	req := new(GeneralQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryUrl(req.Url, req.QueryOptions))
}

func (m *JrpcMethods) QueryChain(_ context.Context, params json.RawMessage) interface{} {
	req := new(ChainIdQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryChain(req.ChainId))
}

func (m *JrpcMethods) QueryData(_ context.Context, params json.RawMessage) interface{} {
	req := new(DataEntryQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryData(req.Url, req.EntryHash))
}

func (m *JrpcMethods) QueryDataSet(_ context.Context, params json.RawMessage) interface{} {
	req := new(DataEntrySetQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryDataSet(req.Url, req.QueryPagination, req.QueryOptions))
}

func (m *JrpcMethods) QueryDirectory(_ context.Context, params json.RawMessage) interface{} {
	req := new(DirectoryQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryDirectory(req.Url, req.QueryPagination, req.QueryOptions))
}

func (m *JrpcMethods) QueryKeyPageIndex(_ context.Context, params json.RawMessage) interface{} {
	req := new(KeyPageIndexQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryKeyPageIndex(req.Url, req.Key))
}

func (m *JrpcMethods) QueryPending(_ context.Context, params json.RawMessage) interface{} {
	req := new(PendingTransactionQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryPending(req.Url, req.Txid, req.QueryOptions))
}

func (m *JrpcMethods) QueryPendingPagination(_ context.Context, params json.RawMessage) interface{} {
	req := new(PendingTransactionPaginationQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryPendingPagination(req.Url, req.QueryPagination, req.QueryOptions))
}

func (m *JrpcMethods) QueryTx(_ context.Context, params json.RawMessage) interface{} {
	req := new(TxnQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryTx(req.Txid, req.Wait, req.QueryOptions))
}

func (m *JrpcMethods) QueryTxHistory(_ context.Context, params json.RawMessage) interface{} {
	req := new(TxHistoryQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.opts.Query.QueryTxHistory(req.Url, req.QueryPagination))
}
