package rpc

import (
	"os"
)

type Manager struct {
	znnClient *ZnnRpc
	stopChan  chan os.Signal
}

func NewRpcManager(urls []string, stop chan os.Signal) (*Manager, error) {
	newZnnClient, err := NewZnnRpcClient(urls)
	if err != nil {
		return nil, err
	}

	return &Manager{
		znnClient: newZnnClient,
		stopChan:  stop,
	}, nil
}

func (m *Manager) Znn() *ZnnRpc {
	return m.znnClient
}
