package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

func TestIdentityCreate(t *testing.T) {
	ic := IdentityCreate{}
	ic.SetName("ACME")
	kp := CreateKeyPair()
	kh := Bytes32(sha256.Sum256(kp.PubKey().Bytes()))
	ic.SetKeyHash(&kh)

	data, err := json.Marshal(&ic)
	if err != nil {
		t.Fatal(err)
	}

	ic2 := IdentityCreate{}
	err = json.Unmarshal(data, &ic2)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))

	data, err = ic.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	ic3 := IdentityCreate{}
	err = ic3.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}

	if ic.IdentityName != ic3.IdentityName {
		t.Fatalf("Unmarshalled identity doesn't match")
	}

	if bytes.Compare(ic.IdentityKeyHash[:], ic3.IdentityKeyHash[:]) != 0 {
		t.Fatalf("Unmarshalled key hash doesn't match")
	}
}
