package protocol

import (
	"encoding"

	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
)

// Fee is the unit cost of a transaction.
type Fee uint64

func (n Fee) AsUInt64() uint64 {
	return uint64(n)
}

const (
	// FeeFailedMaximum $0.01
	FeeFailedMaximum Fee = 100

	// FeeSignature $0.0001
	FeeSignature Fee = 1

	// FeeCreateIdentity $5.00 = 50000 credits @ 0.0001 / credit.
	FeeCreateIdentity Fee = 50000

	// FeeCreateAccount $0.25
	FeeCreateAccount Fee = 2500

	// FeeTransferTokens $0.03
	FeeTransferTokens Fee = 300

	// FeeTransferTokensExtra $0.01
	FeeTransferTokensExtra Fee = 100

	// FeeCreateToken $50.00
	FeeCreateToken Fee = 500000

	// FeeGeneralSmall $0.001
	FeeGeneralSmall Fee = 10

	// FeeCreateKeyPage $1.00
	FeeCreateKeyPage Fee = 10000

	// FeeCreateKeyPageExtra $0.01
	FeeCreateKeyPageExtra Fee = 100

	// FeeData $0.001 / 256 bytes
	FeeData Fee = 10

	// FeeScratchData $0.0001 / 256 bytes
	FeeScratchData Fee = 1

	// FeeUpdateAuth $0.03
	FeeUpdateAuth Fee = 300

	// FeeUpdateAuthExtra $0.01
	FeeUpdateAuthExtra Fee = 100

	// MinimumCreditPurchase $0.01
	MinimumCreditPurchase Fee = 100
)

func dataCount(obj encoding.BinaryMarshaler) (int, int, error) {
	// Check the transaction size (including signatures)
	data, err := obj.MarshalBinary()
	if err != nil {
		return 0, 0, errors.Wrap(errors.StatusInternalError, err)
	}

	// count the number of 256-byte chunks
	size := len(data)
	count := size / 256
	if size%256 != 0 {
		count++
	}

	return count, size, nil
}

func ComputeSignatureFee(sig Signature) (Fee, error) {
	// Check the transaction size
	count, size, err := dataCount(sig)
	if err != nil {
		return 0, errors.Wrap(errors.StatusUnknownError, err)
	}
	if size > SignatureSizeMax {
		return 0, errors.Format(errors.StatusBadRequest, "signature size exceeds %v byte entry limit", SignatureSizeMax)
	}

	// Base fee
	fee := FeeSignature

	// Charge extra for each 256B past the first
	fee += FeeSignature * Fee(count-1)

	// Charge extra for each layer of delegation
	for {
		del, ok := sig.(*DelegatedSignature)
		if !ok {
			break
		}
		fee += FeeSignature
		sig = del.Signature
	}

	return fee, nil
}

func ComputeTransactionFee(tx *Transaction) (Fee, error) {
	// Do not charge fees for the DN or BVNs
	if IsDnUrl(tx.Header.Principal) {
		return 0, nil
	}
	if _, ok := ParsePartitionUrl(tx.Header.Principal); ok {
		return 0, nil
	}

	// Don't charge for synthetic and internal transactions
	if !tx.Body.Type().IsUser() {
		return 0, nil
	}

	// Check the transaction size
	count, size, err := dataCount(tx)
	if err != nil {
		return 0, errors.Wrap(errors.StatusUnknownError, err)
	}
	if size > TransactionSizeMax {
		return 0, errors.Format(errors.StatusBadRequest, "transaction size exceeds %v byte entry limit", TransactionSizeMax)
	}

	var fee Fee
	switch body := tx.Body.(type) {
	case *CreateToken:
		fee = FeeCreateToken + FeeData*Fee(count-1)

	case *CreateIdentity:
		fee = FeeCreateIdentity + FeeData*Fee(count-1)

	case *CreateTokenAccount,
		*CreateDataAccount:
		fee = FeeCreateAccount + FeeData*Fee(count-1)

	case *SendTokens:
		fee = FeeTransferTokens + FeeTransferTokensExtra*Fee(len(body.To)-1) + FeeData*Fee(count-1)
	case *IssueTokens:
		fee = FeeTransferTokens + FeeTransferTokensExtra*Fee(len(body.To)-1) + FeeData*Fee(count-1)
	case *CreateLiteTokenAccount:
		fee = FeeTransferTokens + FeeData*Fee(count-1)

	case *CreateKeyBook:
		fee = FeeCreateKeyPage + FeeData*Fee(count-1)
	case *CreateKeyPage:
		fee = FeeCreateKeyPage + FeeCreateKeyPageExtra*Fee(len(body.Keys)-1) + FeeData*Fee(count-1)

	case *UpdateKeyPage:
		fee = FeeUpdateAuth + FeeUpdateAuthExtra*Fee(len(body.Operation)-1) + FeeData*Fee(count-1)
	case *UpdateAccountAuth:
		fee = FeeUpdateAuth + FeeUpdateAuthExtra*Fee(len(body.Operations)-1) + FeeData*Fee(count-1)
	case *UpdateKey:
		fee = FeeUpdateAuth + FeeData*Fee(count-1)

	case *BurnTokens,
		*LockAccount:
		fee = FeeGeneralSmall + FeeData*Fee(count-1)

	case *WriteData:
		fee = Fee(count)
		if body.Scratch {
			fee *= FeeScratchData
		} else {
			fee *= FeeData
		}
		if body.WriteToState {
			fee *= 2
		}

	case *WriteDataTo:
		fee = FeeData * Fee(count)

	case *AddCredits,
		*AcmeFaucet:
		fee = 0

	default:
		// All user transactions must have a defined fee amount, even if it's zero
		return 0, errors.Format(errors.StatusInternalError, "unknown transaction type %v", body.Type())
	}

	return fee, nil
}
