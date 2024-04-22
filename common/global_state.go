package common

import (
	"context"
	"github.com/MoonBaZZe/hauler/common/block_header"
	"github.com/btcsuite/btcd/wire"
	"golang.org/x/sync/semaphore"
)

const (
	LiveState uint8 = 0
)

type GlobalState struct {
	state                  *uint8
	stateSemaphore         *semaphore.Weighted
	frontierMomentumHeight uint64
	frontierMomSemaphore   *semaphore.Weighted
	bestBlock              *block_header.BlockHeader
	bestBlockSemaphore     *semaphore.Weighted
}

func NewGlobalState(state *uint8) *GlobalState {
	return &GlobalState{
		state:                  state,
		frontierMomentumHeight: 0,
		bestBlock:              nil,
		stateSemaphore:         semaphore.NewWeighted(1),
		frontierMomSemaphore:   semaphore.NewWeighted(1),
		bestBlockSemaphore:     semaphore.NewWeighted(1),
	}
}

func (gs *GlobalState) GetState() (uint8, error) {
	err := gs.stateSemaphore.Acquire(context.Background(), 1)
	if err != nil {
		return 0, err
	}
	defer gs.stateSemaphore.Release(1)
	return *gs.state, nil
}

func (gs *GlobalState) SetFrontierMomentum(frMom uint64) error {
	err := gs.frontierMomSemaphore.Acquire(context.Background(), 1)
	if err != nil {
		return err
	}
	gs.frontierMomentumHeight = frMom
	gs.frontierMomSemaphore.Release(1)
	return nil
}

func (gs *GlobalState) GetFrontierMomentum() (uint64, error) {
	err := gs.frontierMomSemaphore.Acquire(context.Background(), 1)
	if err != nil {
		return 0, err
	}
	defer gs.frontierMomSemaphore.Release(1)
	return gs.frontierMomentumHeight, nil
}

func (gs *GlobalState) SetBestBlockHeader(header *wire.BlockHeader) error {
	err := gs.bestBlockSemaphore.Acquire(context.Background(), 1)
	if err != nil {
		return err
	}
	gs.bestBlock = WireBlockHeaderToHaulerBlockHeader(header)
	gs.bestBlockSemaphore.Release(1)
	return nil
}

func (gs *GlobalState) GetBestBlockHeader() (*block_header.BlockHeader, error) {
	err := gs.bestBlockSemaphore.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}
	defer gs.bestBlockSemaphore.Release(1)
	return gs.bestBlock, nil
}
