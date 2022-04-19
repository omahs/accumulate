package api

// GENERATED BY go run ./tools/cmd/gen-api. DO NOT EDIT.

import (
	"context"
	"encoding/json"

	"github.com/AccumulateNetwork/jsonrpc2/v15"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

func (m *JrpcMethods) populateMethodTable() jsonrpc2.MethodMap {
	if m.methods == nil {
		m.methods = make(jsonrpc2.MethodMap, 34)
	}

	m.methods["describe"] = m.Describe
	m.methods["execute"] = m.Execute
	m.methods["add-credits"] = m.ExecuteAddCredits
	m.methods["add-validator"] = m.ExecuteAddValidator
	m.methods["burn-tokens"] = m.ExecuteBurnTokens
	m.methods["create-adi"] = m.ExecuteCreateAdi
	m.methods["create-data-account"] = m.ExecuteCreateDataAccount
	m.methods["create-identity"] = m.ExecuteCreateIdentity
	m.methods["create-key-book"] = m.ExecuteCreateKeyBook
	m.methods["create-key-page"] = m.ExecuteCreateKeyPage
	m.methods["create-token"] = m.ExecuteCreateToken
	m.methods["create-token-account"] = m.ExecuteCreateTokenAccount
	m.methods["issue-tokens"] = m.ExecuteIssueTokens
	m.methods["remove-validator"] = m.ExecuteRemoveValidator
	m.methods["send-tokens"] = m.ExecuteSendTokens
	m.methods["update-account-auth"] = m.ExecuteUpdateAccountAuth
	m.methods["update-key"] = m.ExecuteUpdateKey
	m.methods["update-key-page"] = m.ExecuteUpdateKeyPage
	m.methods["update-validator-key"] = m.ExecuteUpdateValidatorKey
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
	m.methods["query-minor-blocks"] = m.QueryMinorBlocks
	m.methods["query-tx"] = m.QueryTx
	m.methods["query-tx-history"] = m.QueryTxHistory
	m.methods["status"] = m.Status
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

// ExecuteAddCredits submits an AddCredits transaction.
func (m *JrpcMethods) ExecuteAddCredits(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.AddCredits))
}

// ExecuteAddValidator submits an AddValidator transaction.
func (m *JrpcMethods) ExecuteAddValidator(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.AddValidator))
}

// ExecuteBurnTokens submits a BurnTokens transaction.
func (m *JrpcMethods) ExecuteBurnTokens(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.BurnTokens))
}

// ExecuteCreateAdi submits a CreateIdentity transaction.
func (m *JrpcMethods) ExecuteCreateAdi(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateIdentity))
}

// ExecuteCreateDataAccount submits a CreateDataAccount transaction.
func (m *JrpcMethods) ExecuteCreateDataAccount(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateDataAccount))
}

// ExecuteCreateIdentity submits a CreateIdentity transaction.
func (m *JrpcMethods) ExecuteCreateIdentity(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateIdentity))
}

// ExecuteCreateKeyBook submits a CreateKeyBook transaction.
func (m *JrpcMethods) ExecuteCreateKeyBook(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateKeyBook))
}

// ExecuteCreateKeyPage submits a CreateKeyPage transaction.
func (m *JrpcMethods) ExecuteCreateKeyPage(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateKeyPage))
}

// ExecuteCreateToken submits a CreateToken transaction.
func (m *JrpcMethods) ExecuteCreateToken(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateToken))
}

// ExecuteCreateTokenAccount submits a CreateTokenAccount transaction.
func (m *JrpcMethods) ExecuteCreateTokenAccount(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.CreateTokenAccount))
}

// ExecuteIssueTokens submits an IssueTokens transaction.
func (m *JrpcMethods) ExecuteIssueTokens(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.IssueTokens))
}

// ExecuteRemoveValidator submits a RemoveValidator transaction.
func (m *JrpcMethods) ExecuteRemoveValidator(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.RemoveValidator))
}

// ExecuteSendTokens submits a SendTokens transaction.
func (m *JrpcMethods) ExecuteSendTokens(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.SendTokens), "From", "To")
}

// ExecuteUpdateAccountAuth submits an UpdateAccountAuth transaction.
func (m *JrpcMethods) ExecuteUpdateAccountAuth(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.UpdateAccountAuth))
}

// ExecuteUpdateKey submits an UpdateKey transaction.
func (m *JrpcMethods) ExecuteUpdateKey(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.UpdateKey))
}

// ExecuteUpdateKeyPage submits an UpdateKeyPage transaction.
func (m *JrpcMethods) ExecuteUpdateKeyPage(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.UpdateKeyPage))
}

// ExecuteUpdateValidatorKey submits an UpdateValidatorKey transaction.
func (m *JrpcMethods) ExecuteUpdateValidatorKey(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.UpdateValidatorKey))
}

// ExecuteWriteData submits a WriteData transaction.
func (m *JrpcMethods) ExecuteWriteData(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.WriteData))
}

// ExecuteWriteDataTo submits a WriteDataTo transaction.
func (m *JrpcMethods) ExecuteWriteDataTo(ctx context.Context, params json.RawMessage) interface{} {
	return m.executeWith(ctx, params, new(protocol.WriteDataTo))
}

// Query queries an account or account chain by URL.
func (m *JrpcMethods) Query(_ context.Context, params json.RawMessage) interface{} {
	req := new(GeneralQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryUrl(req.Url, req.QueryOptions))
}

// QueryChain queries an account by ID.
func (m *JrpcMethods) QueryChain(_ context.Context, params json.RawMessage) interface{} {
	req := new(ChainIdQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryChain(req.ChainId))
}

// QueryData queries an entry on an account's data chain.
func (m *JrpcMethods) QueryData(_ context.Context, params json.RawMessage) interface{} {
	req := new(DataEntryQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryData(req.Url, req.EntryHash))
}

// QueryDataSet queries a range of entries on an account's data chain.
func (m *JrpcMethods) QueryDataSet(_ context.Context, params json.RawMessage) interface{} {
	req := new(DataEntrySetQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryDataSet(req.Url, req.QueryPagination, req.QueryOptions))
}

// QueryDirectory queries the directory entries of an account.
func (m *JrpcMethods) QueryDirectory(_ context.Context, params json.RawMessage) interface{} {
	req := new(DirectoryQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryDirectory(req.Url, req.QueryPagination, req.QueryOptions))
}

// QueryKeyPageIndex queries the location of a key within an account's key book(s).
func (m *JrpcMethods) QueryKeyPageIndex(_ context.Context, params json.RawMessage) interface{} {
	req := new(KeyPageIndexQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryKeyPageIndex(req.Url, req.Key))
}

// QueryMinorBlocks queries an account's minor blocks.
//
// WARNING: EXPERIMENTAL!
func (m *JrpcMethods) QueryMinorBlocks(_ context.Context, params json.RawMessage) interface{} {
	req := new(MinorBlocksQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryMinorBlocks(req.Url, req.QueryPagination, req.TxFetchMode, req.FilterSynthAnchorsOnlyBlocks))
}

// QueryTx queries a transaction by ID.
func (m *JrpcMethods) QueryTx(_ context.Context, params json.RawMessage) interface{} {
	req := new(TxnQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryTx(req.Txid, req.Wait, req.IgnorePending, req.QueryOptions))
}

// QueryTxHistory queries an account's transaction history.
func (m *JrpcMethods) QueryTxHistory(_ context.Context, params json.RawMessage) interface{} {
	req := new(TxHistoryQuery)
	err := m.parse(params, req)
	if err != nil {
		return err
	}

	return jrpcFormatResponse(m.querier.QueryTxHistory(req.Url, req.QueryPagination))
}
