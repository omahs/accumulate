package validator

import (
	"crypto/sha256"
	"github.com/AccumulateNetwork/accumulated/types"
	"github.com/AccumulateNetwork/accumulated/types/proto"
	"github.com/AccumulateNetwork/accumulated/types/state"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"math/big"
	"testing"
)

func CreateFakeIdentityState(identitychainpath string, key ed25519.PrivKey) (*state.Object, []byte) {
	id, _, _ := types.ParseIdentityChainPath(identitychainpath)

	idhash := sha256.Sum256([]byte(id))

	so := state.Object{}
	ids := state.NewIdentityState(id)
	ids.SetKeyData(0, key.PubKey().Bytes())
	so.Entry, _ = ids.MarshalBinary()

	eh := sha256.Sum256(so.Entry)
	so.EntryHash = eh[:]
	//we intentionally don't set the so.StateHash & so.PrevStateHash
	return &so, idhash[:]
}

func CreateFakeTokenAccountState(identitychainpath string, t *testing.T) *state.Object {
	id, cp, _ := types.ParseIdentityChainPath(identitychainpath)

	tas := state.NewTokenAccountState([]byte(id), []byte(cp), nil)

	deposit := big.NewInt(5000)
	tas.AddBalance(deposit)

	so := state.Object{}
	so.Entry, _ = tas.MarshalBinary()
	eh := sha256.Sum256(so.Entry)
	so.EntryHash = eh[:]
	//we intentionally don't set the so.StateHash & so.PrevStateHash
	return &so
}

func CreateFakeTokenTransaction(t *testing.T, kp ed25519.PrivKey) *proto.Submission {

	inputamt := big.NewInt(5000)

	identityname := "RoadRunner"
	tokenchainname := "RoadRunner/ACME"

	outputs := make(map[string]*big.Int)
	outputs["WileECoyote/MyACMEToken"] = big.NewInt(5000)

	sub, err := types.CreateTokenTransaction(&identityname, &tokenchainname,
		inputamt, &outputs, nil, kp)
	if err != nil {
		t.Fatalf("Failed to make a token rpc call %v", err)
	}
	return sub
}

func TestTokenTransactionValidator_Check(t *testing.T) {
	kp := types.CreateKeyPair()
	identitychainpath := "RoadRunner/ACME"
	currentstate := StateEntry{}
	currentstate.ChainState = CreateFakeTokenAccountState(identitychainpath, t)
	var idhash []byte
	currentstate.IdentityState, idhash = CreateFakeIdentityState(identitychainpath, kp)

	chainhash := sha256.Sum256([]byte(identitychainpath))

	ttv := NewTokenTransactionValidator()

	faketx := CreateFakeTokenTransaction(t, kp)

	//need to simulate a state entry for chain and token
	err := ttv.Check(&currentstate, idhash, chainhash[:], 0, 0, faketx.Data)
	if err != nil {
		t.Fatalf("Error performing check %v", err)
	}
}
