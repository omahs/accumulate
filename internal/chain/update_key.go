package chain

import (
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type UpdateKey struct{}

func (UpdateKey) Type() protocol.TransactionType {
	return protocol.TransactionTypeUpdateKey
}

func (UpdateKey) validate(st *StateManager, tx *Delivery) (*protocol.UpdateKey, *protocol.KeyPage, *protocol.KeyBook, error) {
	body, ok := tx.Transaction.Body.(*protocol.UpdateKey)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid payload: want %T, got %T", new(protocol.UpdateKey), tx.Transaction.Body)
	}
	if len(body.NewKeyHash) != 32 {
		return nil, nil, nil, errors.New(errors.StatusBadRequest, "key hash is not a valid length for a SHA-256 hash")
	}

	page, ok := st.Origin.(*protocol.KeyPage)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid principal: want account type %v, got %v", protocol.AccountTypeKeyPage, st.Origin.Type())
	}

	bookUrl, _, ok := protocol.ParseKeyPageUrl(st.OriginUrl)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid principal: page url is invalid: %s", page.Url)
	}

	var book *protocol.KeyBook
	err := st.LoadUrlAs(bookUrl, &book)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid key book: %v", err)
	}

	if book.BookType == protocol.BookTypeValidator {
		return nil, nil, nil, fmt.Errorf("UpdateKey cannot be used to modify the validator key book")
	}

	return body, page, book, nil
}

func (UpdateKey) Validate(st *StateManager, tx *Delivery) (protocol.TransactionResult, error) {
	_, _, _, err := UpdateKey{}.validate(st, tx)
	return nil, err
}

func (UpdateKey) Execute(st *StateManager, tx *Delivery) (protocol.TransactionResult, error) {
	body, page, book, err := UpdateKey{}.validate(st, tx)
	if err != nil {
		return nil, err
	}

	// Do not update the key page version. Do not reset LastUsedOn.

	// Find the first signature
	txObj := st.batch.Transaction(tx.Transaction.GetHash())
	status, err := txObj.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("load transaction status: %w", err)
	}

	var initiator protocol.Signature
outer:
	for _, signer := range status.Signers {
		sigs, err := database.GetSignaturesForSigner(txObj, signer)
		if err != nil {
			return nil, fmt.Errorf("load signatures for %v: %w", signer.GetUrl(), err)
		}

		for _, sig := range sigs {
			if protocol.SignatureDidInitiate(sig, tx.Transaction.Header.Initiator[:]) {
				initiator = sig
				break outer
			}
		}
	}

	control := true

	for control {
		fmt.Println(initiator.Type())
		switch initiator.(type) {
		case *protocol.DelegatedSignature:
			initiator = initiator.(*protocol.DelegatedSignature).Signature
		case protocol.KeySignature:
			fmt.Println("keysignature")
			control = false
		default:
			fmt.Println("default")
			control = false
		}
	}
	var keysig protocol.KeySignature
	var ok bool
	if initiator.Type() == protocol.SignatureTypeSet {
		set := initiator.(*protocol.SignatureSet).Signatures
		fmt.Println(set, len(set))
		keysig, ok = set[0].(protocol.KeySignature)
	} else {
		keysig, ok = initiator.(protocol.KeySignature)
		if !ok {
			return nil, fmt.Errorf("signature is not a key signature")

		}
	}
	err = updateKey(page, book,
		&protocol.KeySpecParams{KeyHash: keysig.GetPublicKeyHash()},
		&protocol.KeySpecParams{KeyHash: body.NewKeyHash}, true)
	if err != nil {
		return nil, err
	}

	// Store the update, but do not change the page version
	err = st.Update(page)
	if err != nil {
		return nil, fmt.Errorf("failed to update %v: %v", page.GetUrl(), err)
	}
	return nil, nil
}

func updateKey(page *protocol.KeyPage, book *protocol.KeyBook, old, new *protocol.KeySpecParams, preserveDelegate bool) error {
	if new.IsEmpty() {
		return fmt.Errorf("cannot add an empty entry")
	}

	if new.Delegate != nil {
		if new.Delegate.ParentOf(page.Url) {
			return fmt.Errorf("self-delegation is not allowed")
		}

		if err := verifyIsNotPage(&book.AccountAuth, new.Delegate); err != nil {
			return errors.Format(errors.StatusUnknownError, "invalid delegate %v: %w", new.Delegate, err)
		}
	}

	// Find the old entry
	oldPos, entry, found := findKeyPageEntry(page, old)
	if !found {
		return fmt.Errorf("entry to be updated not found on the key page")
	}

	// Check for an existing key with same delegate
	newPos, _, found := findKeyPageEntry(page, new)
	if found && oldPos != newPos {
		return fmt.Errorf("cannot have duplicate entries on key page")
	}

	// Update the entry
	entry.PublicKeyHash = new.KeyHash

	if new.Delegate != nil || !preserveDelegate {
		entry.Delegate = new.Delegate
	}

	// Relocate the entry
	page.RemoveKeySpecAt(oldPos)
	page.AddKeySpec(entry)
	return nil
}
