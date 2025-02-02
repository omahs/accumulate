// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

//go:build !production
// +build !production

package testing

import (
	"github.com/AccumulateNetwork/jsonrpc2/v15"
	"gitlab.com/accumulatenetwork/accumulate/internal/database/smt/storage"
	"gitlab.com/accumulatenetwork/accumulate/internal/database/smt/storage/memory"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
)

func EnableDebugFeatures() {
	jsonrpc2.DebugMethodFunc = true
	errors.EnableLocationTracking()
	storage.EnableKeyNameTracking()
	memory.EnableLogWrites()
}

func DisableDebugFeatures() {
	errors.DisableLocationTracking()
	storage.DisableKeyNameTracking()
	memory.DisableLogWrites()
}
