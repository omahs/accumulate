// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package events

import (
	"time"

	"gitlab.com/accumulatenetwork/accumulate/internal/core"
)

type Event interface {
	isEvent()
}

func (DidCommitBlock) isEvent()    {}
func (DidSaveSnapshot) isEvent()   {}
func (WillChangeGlobals) isEvent() {}
func (FatalError) isEvent()        {}

type DidCommitBlock struct {
	Index uint64
	Time  time.Time
	Major uint64
}

type DidSaveSnapshot struct {
	MinorIndex uint64
}

type WillChangeGlobals struct {
	New, Old *core.GlobalValues
}

type FatalError struct {
	Err error
}

func (e FatalError) Error() string { return e.Err.Error() }
func (e FatalError) Unwrap() error { return e.Err }
