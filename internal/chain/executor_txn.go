package chain

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/indexing"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

// CheckTx implements ./abci.Chain
func (m *Executor) CheckTx(env *protocol.Envelope) (protocol.TransactionResult, *protocol.Error) {
	batch := m.DB.Begin(false)
	defer batch.Discard()

	st, executor, hasEnoughSigs, err := m.validate(batch, env)
	var notFound bool
	switch {
	case err == nil:
		// OK
	case errors.Is(err, storage.ErrNotFound):
		notFound = true
	default:
		return nil, &protocol.Error{Code: protocol.ErrorCodeCheckTxError, Message: err}
	}

	// Do not run transaction-specific validation for a synthetic transaction. A
	// synthetic transaction will be rejected by `m.validate` unless it is
	// signed by a BVN and can be proved to have been included in a DN block. If
	// `m.validate` succeeeds, we know the transaction came from a BVN, thus it
	// is safe and reasonable to allow the transaction to be delivered.
	//
	// This is important because if a synthetic transaction is rejected during
	// CheckTx, it does not get recorded. If the synthetic transaction is not
	// recorded, the BVN that sent it and the client that sent the original
	// transaction cannot verify that the synthetic transaction was received.
	if st.txType.IsSynthetic() {
		return new(protocol.EmptyResult), nil
	}

	// Synthetic transactions with a missing origin should still be recorded.
	// Thus this check only happens if the transaction is not synthetic.
	if notFound {
		return nil, &protocol.Error{Code: protocol.ErrorCodeNotFound, Message: err}
	}

	// Do not validate if we don't have enough signatures
	if !hasEnoughSigs {
		return new(protocol.EmptyResult), nil
	}

	result, err := executor.Validate(st, env)
	if err != nil {
		return nil, &protocol.Error{Code: protocol.ErrorCodeValidateTxnError, Message: err}
	}
	if result == nil {
		result = new(protocol.EmptyResult)
	}
	return result, nil
}

// DeliverTx implements ./abci.Chain
func (m *Executor) DeliverTx(env *protocol.Envelope) (protocol.TransactionResult, *protocol.Error) {
	// if txt.IsInternal() && tx.Transaction.Nonce != uint64(m.blockIndex) {
	// 	err := fmt.Errorf("nonce does not match block index, want %d, got %d", m.blockIndex, tx.Transaction.Nonce)
	// 	return nil, m.recordTransactionError(tx, nil, nil, &chainId, tx.GetTxHash(), &protocol.Error{Code: protocol.CodeInvalidTxnError, Message: err})
	// }

	// Set up the state manager and validate the signatures
	st, executor, hasEnoughSigs, err := m.validate(m.blockBatch, env)
	if err != nil {
		return nil, m.recordTransactionError(nil, env, false, &protocol.Error{Code: protocol.ErrorCodeCheckTxError, Message: fmt.Errorf("txn check failed : %v", err)})
	}

	if !hasEnoughSigs {
		// Write out changes to the nonce and credit balance
		err = st.Commit()
		if err != nil {
			return nil, m.recordTransactionError(st, env, true, &protocol.Error{Code: protocol.ErrorCodeRecordTxnError, Message: err})
		}

		status := &protocol.TransactionStatus{Pending: true}
		err = m.putTransaction(st, env, status, false)
		if err != nil {
			return nil, &protocol.Error{Code: protocol.ErrorCodeTxnStateError, Message: err}
		}
		return new(protocol.EmptyResult), nil
	}

	result, err := executor.Validate(st, env)
	if err != nil {
		return nil, m.recordTransactionError(st, env, false, &protocol.Error{Code: protocol.ErrorCodeInvalidTxnError, Message: fmt.Errorf("txn validation failed : %v", err)})
	}
	if result == nil {
		result = new(protocol.EmptyResult)
	}

	// Store pending state updates, queue state creates for synthetic transactions
	err = st.Commit()
	if err != nil {
		return nil, m.recordTransactionError(st, env, true, &protocol.Error{Code: protocol.ErrorCodeRecordTxnError, Message: err})
	}

	// Store the tx state
	status := &protocol.TransactionStatus{Delivered: true, Result: result}
	err = m.putTransaction(st, env, status, false)
	if err != nil {
		return nil, &protocol.Error{Code: protocol.ErrorCodeTxnStateError, Message: err}
	}

	m.blockState.Merge(&st.blockState)

	// Process synthetic transactions generated by the validator
	produced := st.blockState.ProducedTxns
	st.Reset()
	err = m.addSynthTxns(&st.stateCache, produced)
	if err != nil {
		return nil, &protocol.Error{Code: protocol.ErrorCodeSyntheticTxnError, Message: err}
	}
	err = st.Commit()
	if err != nil {
		return nil, m.recordTransactionError(st, env, true, &protocol.Error{Code: protocol.ErrorCodeRecordTxnError, Message: err})
	}

	m.blockState.Delivered++
	return result, nil
}

func (m *Executor) processInternalDataTransaction(internalAccountPath string, wd *protocol.WriteData) error {
	dataAccountUrl := m.Network.NodeUrl(internalAccountPath)

	if wd == nil {
		return fmt.Errorf("no internal data transaction provided")
	}

	signer := m.Network.ValidatorPage(0)
	env := new(protocol.Envelope)
	env.Transaction = new(protocol.Transaction)
	env.Transaction.Header.Principal = m.Network.NodeUrl()
	env.Transaction.Body = wd
	env.Transaction.Header.Initiator = signer.AccountID32()
	env.Signatures = []protocol.Signature{&protocol.InternalSignature{Network: signer}}

	sw := protocol.SegWitDataEntry{}
	sw.Cause = *(*[32]byte)(env.GetTxHash())
	sw.EntryHash = *(*[32]byte)(wd.Entry.Hash())
	sw.EntryUrl = env.Transaction.Header.Principal
	env.Transaction.Body = &sw

	st, err := NewStateManager(m.blockBatch, m.Network.NodeUrl(internalAccountPath), env)
	if err != nil {
		return err
	}
	st.logger.L = m.logger

	da := new(protocol.DataAccount)
	va := m.blockBatch.Account(dataAccountUrl)
	err = va.GetStateAs(da)
	if err != nil {
		return err
	}

	st.UpdateData(da, wd.Entry.Hash(), &wd.Entry)

	status := &protocol.TransactionStatus{Delivered: true}
	err = m.blockBatch.Transaction(env.GetTxHash()).Put(env.Transaction, status, nil)
	if err != nil {
		return err
	}

	err = st.Commit()
	return err
}

// validate validates signatures, verifies they are authorized,
// updates the nonce, and charges the fee.
func (m *Executor) validate(batch *database.Batch, env *protocol.Envelope) (st *StateManager, executor TxExecutor, hasEnoughSigs bool, err error) {
	// Basic validation
	err = m.validateBasic(batch, env)
	if err != nil {
		return nil, nil, false, err
	}

	// Calculate the fee before modifying the transaction
	fee, err := protocol.ComputeTransactionFee(env)
	if err != nil {
		return nil, nil, false, err
	}

	// Load previous transaction state
	txt := env.Transaction.Type()
	txState, err := batch.Transaction(env.GetTxHash()).GetState()
	switch {
	case err == nil:
		// Populate the transaction from the database
		env.Transaction = txState
		txt = env.Transaction.Type()

	case !errors.Is(err, storage.ErrNotFound):
		return nil, nil, false, fmt.Errorf("an error occured while looking up the transaction: %v", err)
	}

	// If this fails, the code is wrong
	executor, ok := m.executors[txt]
	if !ok {
		return nil, nil, false, fmt.Errorf("internal error: unsupported TX type: %v", txt)
	}

	// Set up the state manager
	st, err = NewStateManager(batch, m.Network.NodeUrl(), env)
	if err != nil {
		return nil, nil, false, err
	}
	st.logger.L = m.logger

	// Validate the transaction
	switch {
	case txt.IsUser():
		hasEnoughSigs, err = m.validateUser(st, env, fee)

	case txt.IsSynthetic():
		hasEnoughSigs, err = true, m.validateSynthetic(batch, st, env)

	case txt.IsInternal():
		hasEnoughSigs, err = true, m.validateInternal(st, env)

	default:
		return nil, nil, false, fmt.Errorf("invalid transaction type %v", txt)
	}

	switch {
	case err == nil:
		return st, executor, hasEnoughSigs, err
	case errors.Is(err, storage.ErrNotFound):
		return st, executor, hasEnoughSigs, err
	default:
		return nil, nil, false, err
	}
}

func (m *Executor) validateBasic(batch *database.Batch, env *protocol.Envelope) error {
	// If the transaction is borked, the transaction type is probably invalid,
	// so check that first. "Invalid transaction type" is a more useful error
	// than "invalid signature" if the real error is the transaction got borked.
	txt := env.Transaction.Type()
	_, ok := m.executors[txt]
	if !ok && txt != protocol.TransactionTypeSignPending {
		return fmt.Errorf("unsupported TX type: %v", txt)
	}

	// All transactions must be signed
	if len(env.Signatures) == 0 {
		return fmt.Errorf("transaction is not signed")
	}

	switch {
	case len(env.TxHash) == 0:
		// TxHash must be specified for signature transactions
		if txt == protocol.TransactionTypeSignPending {
			return fmt.Errorf("invalid signature transaction: missing transaction hash")
		}

	case len(env.TxHash) != sha256.Size:
		// TxHash must either be empty or a SHA-256 hash
		return fmt.Errorf("invalid hash: not a SHA-256 hash")

	case !env.VerifyTxHash():
		// TxHash must match the transaction
		if txt != protocol.TransactionTypeSignPending {
			return fmt.Errorf("invalid hash: does not match transaction")
		}
	}

	// TxHash must either be empty or a SHA-256 hash
	if len(env.TxHash) != 0 && len(env.TxHash) != sha256.Size {
		return fmt.Errorf("invalid hash")
	}

	// TxHash must be specified for signature transactions
	if txt == protocol.TransactionTypeSignPending && len(env.TxHash) == 0 {
		return fmt.Errorf("invalid signature transaction: missing transaction hash")
	}

	// Verify the transaction's signatures
	txid := env.GetTxHash()
	for i, v := range env.Signatures {
		if !v.Verify(txid) {
			return fmt.Errorf("signature %d is invalid", i)
		}
	}

	// // TODO Do we need this check? It appears to be causing issues and we can't guarantee that the governor sends transactions only once.
	// // Check the envelope
	// _, err := batch.Transaction(env.EnvHash()).GetState()
	// switch {
	// case err == nil:
	// 	return fmt.Errorf("duplicate envelope")
	// case errors.Is(err, storage.ErrNotFound):
	// 	// OK
	// default:
	// 	return fmt.Errorf("error while checking envelope state: %v", err)
	// }

	// Check the transaction
	status, err := batch.Transaction(env.GetTxHash()).GetStatus()
	switch {
	case err == nil:
		// Signing an existing transaction
		if status.Delivered {
			return fmt.Errorf("transaction has already been delivered")
		}
	case errors.Is(err, storage.ErrNotFound):
		// Initiating a new transaction
		if txt == protocol.TransactionTypeSignPending {
			return fmt.Errorf("transaction not found")
		}

		initHash, err := env.Signatures[0].InitiatorHash()
		if err != nil {
			return fmt.Errorf("failed to initiate a transaction: %v", err)
		}

		if env.Transaction.Header.Initiator != *(*[32]byte)(initHash) {
			return fmt.Errorf("invalid initiator: hash does not match")
		}
	default:
		return fmt.Errorf("failed to load transaction status: %v", err)
	}

	return nil
}

func (m *Executor) validateSynthetic(batch *database.Batch, st *StateManager, env *protocol.Envelope) error {
	//placeholder for special validation rules for synthetic transactions.
	//need to verify the sender is a legit bvc validator also need the dbvc receipt
	//so if the transaction is a synth tx, then we need to verify the sender is a BVC validator and
	//not an impostor. Need to figure out how to do this. Right now we just assume the synth request
	//sender is legit.

	// TODO Get rid of this hack and actually check the nonce. But first we have
	// to implement transaction batching.
	v := st.RecordIndex(m.Network.NodeUrl(), "SeenSynth", env.GetTxHash())
	_, err := v.Get()
	if err == nil {
		return fmt.Errorf("duplicate synthetic transaction %X", env.GetTxHash())
	} else if errors.Is(err, storage.ErrNotFound) {
		err = v.Put([]byte{1})
		if err != nil {
			return err
		}
	} else {
		return err
	}

	for _, sig := range env.Signatures {
		switch sig := sig.(type) {
		case *protocol.SyntheticSignature:
			if !m.Network.NodeUrl().Equal(sig.DestinationNetwork) {
				return fmt.Errorf("destination network %v is not this network", sig.DestinationNetwork)
			}

		case *protocol.ReceiptSignature:
			// TODO We should add something so we know which subnet originated
			// the transaction. That way, the DN can also check receipts.
			if m.Network.Type != config.BlockValidator {
				// TODO Check receipts on the DN
				continue
			}

			// Load the anchor chain
			anchorChain, err := batch.Account(m.Network.AnchorPool()).ReadChain(protocol.AnchorChain(protocol.Directory))
			if err != nil {
				return fmt.Errorf("unable to load the DN intermediate anchor chain: %w", err)
			}

			// Is the result a valid DN anchor?
			_, err = anchorChain.HeightOf(sig.Result)
			switch {
			case err == nil:
				// OK
			case errors.Is(err, storage.ErrNotFound):
				return fmt.Errorf("invalid receipt: result is not a known DN anchor")
			default:
				return fmt.Errorf("unable to check if a DN anchor is valid: %w", err)
			}

		case *protocol.ED25519Signature, *protocol.LegacyED25519Signature:
			// TODO Check the key

		default:
			return fmt.Errorf("synthetic transaction do not support %T signatures", sig)
		}
	}

	// Does the origin exist?
	if st.Origin != nil {
		return nil
	}

	// Is it OK for the origin to be missing?
	switch st.txType {
	case protocol.TransactionTypeSyntheticCreateChain,
		protocol.TransactionTypeSyntheticDepositTokens,
		protocol.TransactionTypeSyntheticWriteData:
		// These transactions allow for a missing origin
		return nil

	default:
		return fmt.Errorf("origin %q %w", st.OriginUrl, storage.ErrNotFound)
	}
}

func (m *Executor) validateInternal(st *StateManager, env *protocol.Envelope) error {
	if st.Origin == nil {
		return fmt.Errorf("origin %q %w", st.OriginUrl, storage.ErrNotFound)
	}

	//placeholder for special validation rules for internal transactions.
	//need to verify that the transaction came from one of the node's governors.
	return nil
}

func (m *Executor) validateUser(st *StateManager, env *protocol.Envelope, fee protocol.Fee) (bool, error) {
	switch signer := st.Signator.(type) {
	case *protocol.LiteTokenAccount:
		return true, m.validateLiteSigner(st, env, signer, fee)

	case *protocol.KeyPage:
		return m.validatePageSigner(st, env, signer, fee)

	default:
		return false, fmt.Errorf("invalid signer: %v cannot sign transactions", signer.GetType())
	}
}

func (m *Executor) validatePageSigner(st *StateManager, env *protocol.Envelope, page *protocol.KeyPage, fee protocol.Fee) (bool, error) {
	// Get the URL of the key page's book
	pageBook, _, ok := protocol.ParseKeyPageUrl(page.Url)
	if !ok {
		return false, fmt.Errorf("invalid key page URL: %v", page.Url)
	}

	// Verify that the key page is authorized to sign transactions for the book
	switch origin := st.Origin.(type) {
	case *protocol.KeyBook:
		if !origin.Url.Equal(pageBook) {
			return false, fmt.Errorf("%v is not authorized to sign transactions for %v", page.Url, origin.Url)
		}
	default:
		if !origin.Header().KeyBook.Equal(pageBook) {
			return false, fmt.Errorf("%v is not authorized to sign transactions for %v", page.Url, origin.Header().Url)
		}
	}

	// Verify that the key page is allowed to sign the transaction
	bit, ok := env.Transaction.Type().AllowedTransactionBit()
	if ok && page.TransactionBlacklist.IsSet(bit) {
		return false, fmt.Errorf("page %s is not authorized to sign %v", st.SignatorUrl, env.Transaction.Type())
	}

	var firstKeySpec *protocol.KeySpec
	for i, sig := range env.Signatures {
		ks := page.FindKey(sig.GetPublicKey())
		if ks == nil {
			return false, fmt.Errorf("no key spec matches signature %d", i)
		}
		if i == 0 {
			firstKeySpec = ks
		}
	}

	sigs, err := st.batch.Transaction(env.GetTxHash()).GetSignatures()
	switch {
	case err == nil:
		// Signing an existing transaction

	case errors.Is(err, storage.ErrNotFound):
		// Initiating a new transaction
		height, err := st.GetHeight(st.SignatorUrl)
		if err != nil {
			return false, err
		}

		initiator := env.Signatures[0]
		if height != initiator.GetSignerHeight() {
			return false, fmt.Errorf("invalid height: want %d, got %d", height, initiator.GetSignerHeight())
		}

		timestamp := initiator.GetTimestamp()
		if timestamp == 0 {
			return false, fmt.Errorf("invalid initiator: missing timestamp")
		}

		if firstKeySpec.Nonce >= timestamp {
			return false, fmt.Errorf("invalid nonce: have %d, received %d", firstKeySpec.Nonce, timestamp)
		}

		firstKeySpec.Nonce = timestamp

	default:
		return false, fmt.Errorf("failed to get signatures: %v", err)
	}

	if !page.DebitCredits(uint64(fee)) {
		return false, fmt.Errorf("insufficent credits for the transaction: %q has %v, cost is %d", page.Url, page.CreditBalance.String(), fee)
	}

	err = st.UpdateSignator(page)
	if err != nil {
		return false, err
	}

	// Add to the sig set to get the resulting count
	sigCount := sigs.Add(env.Signatures...)

	// Queue a write
	st.SignTransaction(env.GetTxHash(), env.Signatures...)

	// If the number of signatures is less than the threshold, the transaction is pending
	return sigCount >= int(page.Threshold), nil
}

func (m *Executor) validateLiteSigner(st *StateManager, env *protocol.Envelope, account *protocol.LiteTokenAccount, fee protocol.Fee) error {
	if !st.SignatorUrl.Equal(st.OriginUrl) {
		return fmt.Errorf("%v is not authorized to sign transactions for %v", st.SignatorUrl, st.OriginUrl)
	}

	urlKH, _, err := protocol.ParseLiteTokenAddress(st.OriginUrl)
	if err != nil {
		// This shouldn't happen because invalid URLs should never make it
		// into the database.
		return fmt.Errorf("invalid lite token URL: %v", err)
	}

	for i, sig := range env.Signatures {
		sigKH := sha256.Sum256(sig.GetPublicKey())
		if !bytes.Equal(urlKH, sigKH[:20]) {
			return fmt.Errorf("signature %d's public key does not match the origin record", i)
		}
	}

	// Don't bother with nonces for the faucet
	if st.txType != protocol.TransactionTypeAcmeFaucet {
		timestamp := env.Signatures[0].GetTimestamp()
		if timestamp == 0 {
			return fmt.Errorf("invalid initiator: missing timestamp")
		}

		if account.Nonce >= timestamp {
			return fmt.Errorf("invalid nonce: have %d, received %d", account.Nonce, timestamp)
		}

		account.Nonce = timestamp
	}

	if !account.DebitCredits(uint64(fee)) {
		return fmt.Errorf("insufficent credits for the transaction: %q has %v, cost is %d", account.Url, account.CreditBalance.String(), fee)
	}

	return st.UpdateSignator(account)
}

func (m *Executor) recordTransactionError(st *StateManager, env *protocol.Envelope, postCommit bool, failure *protocol.Error) *protocol.Error {
	status := &protocol.TransactionStatus{
		Delivered: true,
		Code:      uint64(failure.Code),
		Message:   failure.Error(),
	}
	err := m.putTransaction(st, env, status, postCommit)
	if err != nil {
		m.logError("Failed to store transaction", "txid", logging.AsHex(env.GetTxHash()), "origin", env.Transaction.Header.Principal, "error", err)
	}
	return failure
}

func (m *Executor) putTransaction(st *StateManager, env *protocol.Envelope, status *protocol.TransactionStatus, postCommit bool) (err error) {
	// Don't add internal transactions to chains. Internal transactions are
	// exclusively used for communication between the governor and the state
	// machine.
	txt := env.Transaction.Type()
	if txt.IsInternal() {
		return nil
	}

	// When the transaction is synthetic, send a receipt back to its origin
	if txt.IsSynthetic() && st != nil && NeedsReceipt(txt) { // recordTransactionError can pass in a nil state manager
		receipt, sourceUrl := CreateReceipt(env, status, m.Network.NodeUrl())
		st.logger.Debug("Submitting synth receipt for", st.OriginUrl, " to ", sourceUrl)
		st.Submit(sourceUrl, receipt)
	}

	// Store against the transaction hash
	err = m.blockBatch.Transaction(env.GetTxHash()).Put(env.Transaction, status, env.Signatures)
	if err != nil {
		return fmt.Errorf("failed to store transaction: %v", err)
	}

	// Store against the envelope hash
	err = m.blockBatch.Transaction(env.EnvHash()).Put(env.Transaction, status, env.Signatures)
	if err != nil {
		return fmt.Errorf("failed to store transaction: %v", err)
	}

	if st == nil {
		return nil // Check failed
	}

	if postCommit {
		return nil // Failure after commit
	}

	// Update the account's list of pending transactions
	pending := indexing.PendingTransactions(m.blockBatch, st.OriginUrl)
	if status.Pending {
		err := pending.Add(st.txHash)
		if err != nil {
			return fmt.Errorf("failed to add transaction to the pending list: %v", err)
		}
	} else if status.Delivered {
		err := pending.Remove(st.txHash)
		if err != nil {
			return fmt.Errorf("failed to remove transaction to the pending list: %v", err)
		}
	}

	block := new(BlockState)
	defer func() {
		if err == nil {
			m.blockState.Merge(block)
		}
	}()

	// Add the transaction to the origin's main chain, unless it's pending
	if !status.Pending {
		err = addChainEntry(block, m.blockBatch, st.OriginUrl, protocol.MainChain, protocol.ChainTypeTransaction, env.GetTxHash(), 0, 0)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("failed to add signature to pending chain: %v", err)
		}
	}

	// Add the envelope to the origin's pending chain
	err = addChainEntry(block, m.blockBatch, st.OriginUrl, protocol.SignatureChain, protocol.ChainTypeTransaction, env.EnvHash(), 0, 0)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return fmt.Errorf("failed to add signature to pending chain: %v", err)
	}

	// The signator of a synthetic transaction is the subnet ADI (acc://dn or
	// acc://bvn-{id}). Do not add transactions to the subnet's pending chain
	// and do not charge fees to the subnet's ADI.
	if txt.IsSynthetic() {
		return nil
	}

	// If the origin and signator are different, add the envelope to the signator's pending chain
	if !st.OriginUrl.Equal(st.SignatorUrl) {
		err = addChainEntry(block, m.blockBatch, st.SignatorUrl, protocol.SignatureChain, protocol.ChainTypeTransaction, env.EnvHash(), 0, 0)
		if err != nil {
			return fmt.Errorf("failed to add signature to pending chain: %v", err)
		}
	}

	if status.Code == protocol.ErrorCodeOK.GetEnumValue() {
		return nil
	}

	// Update the nonce and charge the failure transaction fee. Reload the
	// signator to ensure we don't push any invalid changes. Use the database
	// directly, since the state manager won't be committed.
	sigRecord := m.blockBatch.Account(st.SignatorUrl)
	err = sigRecord.GetStateAs(st.Signator)
	if err != nil {
		return err
	}

	// TODO Make sure this is the initiator
	sig := env.Signatures[0]
	err = st.Signator.SetNonce(sig.GetPublicKey(), sig.GetTimestamp())
	if err != nil {
		return err
	}

	fee, err := protocol.ComputeTransactionFee(env)
	if err != nil || fee > protocol.FeeFailedMaximum {
		fee = protocol.FeeFailedMaximum
	}
	st.Signator.DebitCredits(uint64(fee))

	return sigRecord.PutState(st.Signator)
}
