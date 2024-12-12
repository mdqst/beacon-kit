// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2024, Berachain Foundation. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package blockchain

import (
	"context"
	"sync"

	"github.com/berachain/beacon-kit/da/da"
	"github.com/berachain/beacon-kit/execution/deposit"
	"github.com/berachain/beacon-kit/log"
	"github.com/berachain/beacon-kit/node-api/backend"
	blockstore "github.com/berachain/beacon-kit/node-api/block_store"
	"github.com/berachain/beacon-kit/primitives/common"
	"github.com/berachain/beacon-kit/primitives/eip4844"
	"github.com/berachain/beacon-kit/primitives/math"
	"github.com/berachain/beacon-kit/primitives/transition"
)

// Service is the blockchain service.
type Service[
	AvailabilityStoreT AvailabilityStore[BeaconBlockBodyT],
	DepositStoreT backend.DepositStore[DepositT],
	ConsensusBlockT ConsensusBlock[BeaconBlockT],
	BeaconBlockT BeaconBlock[BeaconBlockT, BeaconBlockBodyT],
	BeaconBlockBodyT interface {
		BeaconBlockBody[ExecutionPayloadT]
		GetBlobKzgCommitments() eip4844.KZGCommitments[common.ExecutionHash]
		GetDeposits() []DepositT
	},
	BeaconStateT ReadOnlyBeaconState[
		BeaconStateT, ExecutionPayloadHeaderT,
	],
	BlockStoreT blockstore.BlockStore[BeaconBlockT],
	DepositT deposit.Deposit[DepositT, WithdrawalCredentialsT],
	WithdrawalCredentialsT any,
	ExecutionPayloadT ExecutionPayload,
	ExecutionPayloadHeaderT ExecutionPayloadHeader,
	GenesisT Genesis[DepositT, ExecutionPayloadHeaderT],
	ConsensusSidecarsT da.ConsensusSidecars[BlobSidecarsT],
	BlobSidecarsT BlobSidecars[BlobSidecarsT],

	PayloadAttributesT PayloadAttributes,
] struct {
	// homeDir is the directory for config and data"
	homeDir string
	// storageBackend represents the backend storage for beacon states and
	// associated sidecars.
	storageBackend StorageBackend[
		AvailabilityStoreT,
		BeaconStateT,
		DepositStoreT,
	]
	blobProcessor da.BlobProcessor[
		AvailabilityStoreT,
		ConsensusSidecarsT, BlobSidecarsT,
	]
	// store is the block store for the service.
	// TODO: Remove this and use the block store from the storage backend.
	blockStore BlockStoreT
	// depositStore is the deposit store that stores deposits.
	depositStore deposit.Store[DepositT]
	// depositContract is the contract interface for interacting with the
	// deposit contract.
	depositContract deposit.Contract[DepositT]
	// eth1FollowDistance is the follow distance for Ethereum 1.0 blocks.
	eth1FollowDistance math.U64
	// failedBlocksMu protects failedBlocks for concurrent access.
	failedBlocksMu sync.RWMutex
	// failedBlocks is a map of blocks that failed to be processed
	// and should be retried.
	failedBlocks map[math.U64]struct{}
	// logger is used for logging messages in the service.
	logger log.Logger
	// chainSpec holds the chain specifications.
	chainSpec common.ChainSpec
	// executionEngine is the execution engine responsible for processing
	//
	// execution payloads.
	executionEngine ExecutionEngine[PayloadAttributesT]
	// localBuilder is a local builder for constructing new beacon states.
	localBuilder LocalBuilder[BeaconStateT]
	// stateProcessor is the state processor for beacon blocks and states.
	stateProcessor StateProcessor[
		BeaconBlockT,
		BeaconStateT,
		*transition.Context,
		DepositT,
		ExecutionPayloadHeaderT,
	]
	// metrics is the metrics for the service.
	metrics *chainMetrics
	// optimisticPayloadBuilds is a flag used when the optimistic payload
	// builder is enabled.
	optimisticPayloadBuilds bool
	// forceStartupSyncOnce is used to force a sync of the startup head.
	forceStartupSyncOnce *sync.Once
}

// NewService creates a new validator service.
func NewService[
	AvailabilityStoreT AvailabilityStore[BeaconBlockBodyT],
	DepositStoreT backend.DepositStore[DepositT],
	ConsensusBlockT ConsensusBlock[BeaconBlockT],
	BeaconBlockT BeaconBlock[BeaconBlockT, BeaconBlockBodyT],
	BeaconBlockBodyT interface {
		BeaconBlockBody[ExecutionPayloadT]
		GetBlobKzgCommitments() eip4844.KZGCommitments[common.ExecutionHash]
		GetDeposits() []DepositT
	},
	BeaconStateT ReadOnlyBeaconState[
		BeaconStateT, ExecutionPayloadHeaderT,
	],
	BlockStoreT blockstore.BlockStore[BeaconBlockT],
	DepositT deposit.Deposit[DepositT, WithdrawalCredentialsT],
	WithdrawalCredentialsT any,
	ExecutionPayloadT ExecutionPayload,
	ExecutionPayloadHeaderT ExecutionPayloadHeader,
	GenesisT Genesis[DepositT, ExecutionPayloadHeaderT],
	PayloadAttributesT PayloadAttributes,
	ConsensusSidecarsT da.ConsensusSidecars[BlobSidecarsT],
	BlobSidecarsT BlobSidecars[BlobSidecarsT],
](
	homeDir string,
	storageBackend StorageBackend[
		AvailabilityStoreT,
		BeaconStateT,
		DepositStoreT,
	],
	blobProcessor da.BlobProcessor[
		AvailabilityStoreT,
		ConsensusSidecarsT, BlobSidecarsT,
	],
	blockStore BlockStoreT,
	depositStore deposit.Store[DepositT],
	depositContract deposit.Contract[DepositT],
	eth1FollowDistance math.U64,
	logger log.Logger,
	chainSpec common.ChainSpec,
	executionEngine ExecutionEngine[PayloadAttributesT],
	localBuilder LocalBuilder[BeaconStateT],
	stateProcessor StateProcessor[
		BeaconBlockT,
		BeaconStateT,
		*transition.Context,
		DepositT,
		ExecutionPayloadHeaderT,
	],
	telemetrySink TelemetrySink,
	optimisticPayloadBuilds bool,
) *Service[
	AvailabilityStoreT, DepositStoreT,
	ConsensusBlockT, BeaconBlockT, BeaconBlockBodyT,
	BeaconStateT, BlockStoreT, DepositT, WithdrawalCredentialsT,
	ExecutionPayloadT, ExecutionPayloadHeaderT, GenesisT,
	ConsensusSidecarsT, BlobSidecarsT, PayloadAttributesT,
] {
	return &Service[
		AvailabilityStoreT, DepositStoreT,
		ConsensusBlockT, BeaconBlockT, BeaconBlockBodyT,
		BeaconStateT, BlockStoreT, DepositT, WithdrawalCredentialsT,
		ExecutionPayloadT, ExecutionPayloadHeaderT,
		GenesisT, ConsensusSidecarsT, BlobSidecarsT, PayloadAttributesT,
	]{
		homeDir:                 homeDir,
		storageBackend:          storageBackend,
		blobProcessor:           blobProcessor,
		blockStore:              blockStore,
		depositStore:            depositStore,
		depositContract:         depositContract,
		eth1FollowDistance:      eth1FollowDistance,
		failedBlocks:            make(map[math.Slot]struct{}),
		logger:                  logger,
		chainSpec:               chainSpec,
		executionEngine:         executionEngine,
		localBuilder:            localBuilder,
		stateProcessor:          stateProcessor,
		metrics:                 newChainMetrics(telemetrySink),
		optimisticPayloadBuilds: optimisticPayloadBuilds,
		forceStartupSyncOnce:    new(sync.Once),
	}
}

// Name returns the name of the service.
func (s *Service[
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
]) Name() string {
	return "blockchain"
}

func (s *Service[
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
]) Start(ctx context.Context) error {
	// Catchup deposits for failed blocks.
	go s.depositCatchupFetcher(ctx)

	return nil
}

func (s *Service[
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
]) Stop() error {
	return nil
}