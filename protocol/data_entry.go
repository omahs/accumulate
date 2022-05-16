package protocol

import (
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
	"gitlab.com/accumulatenetwork/accumulate/internal/encoding/hash"
)

type DataEntryType uint64

type DataEntry interface {
	encoding.BinaryValue
	Type() DataEntryType
	Hash() []byte
	GetData() [][]byte
}

// ComputeEntryHash
// returns the entry hash given external id's and data associated with an entry
func ComputeEntryHash(data [][]byte) []byte {
	h := make(hash.Hasher, 0, len(data))
	for _, data := range data {
		h.AddBytes(data)
	}
	return h.MerkleHash()
}

const TransactionSizeMax = 10240
const SignatureSizeMax = 1024

func (e *AccumulateDataEntry) Hash() []byte {
	return ComputeEntryHash(e.Data)
}

func (e *AccumulateDataEntry) GetData() [][]byte {
	return e.Data
}

//CheckDataEntrySize is the marshaled size minus the implicit type header,
//returns error if there is too much or no data
func CheckDataEntrySize(e DataEntry) (int, error) {
	b, err := e.MarshalBinary()
	if err != nil {
		return 0, err
	}
	size := len(b)
	if size > TransactionSizeMax {
		return 0, fmt.Errorf("data amount exceeds %v byte entry limit", TransactionSizeMax)
	}
	if size <= 0 {
		return 0, fmt.Errorf("no data provided for WriteData")
	}
	return size, nil
}

//Cost will return the number of credits to be used for the data write
func DataEntryCost(e DataEntry) (uint64, error) {
	size, err := CheckDataEntrySize(e)
	if err != nil {
		return 0, err
	}
	return FeeData.AsUInt64() * uint64(size/256+1), nil
}
