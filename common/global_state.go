package common

import (
	"context"
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
}

func NewGlobalState(state *uint8) *GlobalState {
	return &GlobalState{
		state:                  state,
		frontierMomentumHeight: 0,
		stateSemaphore:         semaphore.NewWeighted(1),
		frontierMomSemaphore:   semaphore.NewWeighted(1),
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
