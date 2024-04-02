// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Berachain Foundation
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the "Software"), to deal in the Software without
// restriction, including without limitation the rights to use,
// copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following
// conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package beacon

import (
	enginetypes "github.com/berachain/beacon-kit/mod/execution/types"
	"github.com/berachain/beacon-kit/mod/primitives"
)

func (s *Store) UpdateLatestExecutionPayload(
	payload enginetypes.ExecutionPayload,
) error {
	return s.latestExecutionPayload.Set(s.ctx, payload)
}

func (s *Store) GetLatestExecutionPayload() (
	enginetypes.ExecutionPayload, error,
) {
	return s.latestExecutionPayload.Get(s.ctx)
}

// UpdateEth1BlockHash sets the Eth1 hash in the BeaconStore.
func (s *Store) UpdateEth1BlockHash(
	hash primitives.ExecutionHash,
) error {
	return s.eth1BlockHash.Set(s.ctx, hash)
}

// GetEth1Hash retrieves the Eth1 hash from the BeaconStore.
func (s *Store) GetEth1BlockHash() (primitives.ExecutionHash, error) {
	return s.eth1BlockHash.Get(s.ctx)
}
