// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package simulator

//lint:file-ignore ST1001 Don't care

import (
	"os"

	"github.com/stretchr/testify/require"
	"gitlab.com/accumulatenetwork/accumulate/internal/core"
	. "gitlab.com/accumulatenetwork/accumulate/internal/core/v1/block"
	"gitlab.com/accumulatenetwork/accumulate/internal/core/v1/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

func InitFromSnapshot(t TB, db database.Beginner, exec *Executor, filename string) {
	t.Helper()

	f, err := os.Open(filename)
	require.NoError(tb{t}, err)
	defer f.Close()
	batch := db.Begin(true)
	defer batch.Discard()
	require.NoError(tb{t}, exec.RestoreSnapshot(batch, f))
	require.NoError(tb{t}, batch.Commit())
}

func NormalizeEnvelope(t TB, envelope *protocol.Envelope) []*chain.Delivery {
	t.Helper()

	deliveries, err := core.NormalizeEnvelope(envelope)
	require.NoError(tb{t}, err)

	wrapped := make([]*chain.Delivery, len(deliveries))
	for i, d := range deliveries {
		wrapped[i] = &chain.Delivery{Delivery: *d}
	}
	return wrapped
}
