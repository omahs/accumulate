package validator

import (
	"github.com/AccumulateNetwork/accumulated/types"
	"github.com/AccumulateNetwork/accumulated/types/api/transactions"
	pb "github.com/AccumulateNetwork/accumulated/types/proto"
	"github.com/AccumulateNetwork/accumulated/types/state"
	cfg "github.com/tendermint/tendermint/config"

	//dbm "github.com/tendermint/tm-db"
	"time"
)

// BVCLeader is a boilerplate
type BVCLeader struct {
	ValidatorContext

	mdroot [32]byte
}

func NewBVCLeader() *BVCLeader {
	v := BVCLeader{}
	v.SetInfo(types.ChainTypeDC[:], types.ChainSpecDC, pb.AccInstruction_State_Query)
	v.ValidatorContext.ValidatorInterface = &v
	return &v
}

func (v *BVCLeader) Initialize(config *cfg.Config, db *state.StateDB) error {
	return v.ValidatorContext.Initialize(config, db)
}

func (v *BVCLeader) Check(currentstate *state.StateEntry, submission *transactions.GenTransaction) error {
	return nil
}

func (v *BVCLeader) BeginBlock(height int64, time *time.Time) error {
	v.lastHeight = v.currentHeight
	v.lastTime = v.currentTime
	v.currentHeight = height
	v.currentTime = *time

	return nil
}

func (v *BVCLeader) Validate(currentstate *state.StateEntry, submission *transactions.GenTransaction) (*ResponseValidateTX, error) {
	//return persistent entry or error
	return nil, nil
}

func (v *BVCLeader) EndBlock(mdroot []byte) error {
	copy(v.mdroot[:], mdroot[:])
	return nil
}
