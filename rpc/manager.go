package rpc

import (
	"github.com/MoonBaZZe/hauler/common"
	"os"
)

type Manager struct {
	znnClient *ZnnRpc
	btcClient *BtcRpc
	stopChan  chan os.Signal
}

func NewRpcManager(config *common.Config, stop chan os.Signal) (*Manager, error) {
	newZnnClient, err := NewZnnRpcClient(config.NoMEndpoints)
	if err != nil {
		return nil, err
	}

	// Todo: http will have no handlers
	newBtcClient, err := NewBtcRpcClient(config.BtcConfig, nil)
	if err != nil {
		return nil, err
	}

	return &Manager{
		znnClient: newZnnClient,
		btcClient: newBtcClient,
		stopChan:  stop,
	}, nil
}

func (m *Manager) Znn() *ZnnRpc {
	return m.znnClient
}

func (m *Manager) Btc() *BtcRpc {
	return m.btcClient
}
