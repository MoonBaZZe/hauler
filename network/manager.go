package network

import (
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/db/manager"
	"github.com/MoonBaZZe/hauler/rpc"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"os"
)

type NetworksManager struct {
	znnNetwork *ZnnNetwork
	stopChan   chan os.Signal
	logger     *zap.SugaredLogger
}

func NewNetworksManager(stopChan chan os.Signal) (*NetworksManager, error) {
	newLogger, errLogger := common.CreateSugarLogger()
	if errLogger != nil {
		return nil, errLogger
	}

	newNetworkManager := &NetworksManager{
		stopChan: stopChan,
		logger:   newLogger,
	}

	return newNetworkManager, nil
}

func (m *NetworksManager) Init(config *common.Config, dbManager *manager.Manager, state *common.GlobalState) error {
	newRpcManager, err := rpc.NewRpcManager(config.NoMEndpoints, m.stopChan)
	if err != nil {
		return err
	}

	newZnnNetwork, err := NewZnnNetwork(newRpcManager, dbManager, m, state, m.stopChan)
	if err != nil {
		return err
	}
	m.znnNetwork = newZnnNetwork

	// todo create btc network
	// todo are handlers needed?
	//node.BtcClient, err = rpc.NewBtcRpcClient(config.BtcConfig, nil)
	//if err != nil {
	//	return nil, err
	//}
	return nil
}

func (m *NetworksManager) Start() error {
	if err := m.znnNetwork.Start(); err != nil {
		return err
	}

	return nil
}

func (m *NetworksManager) Znn() *ZnnNetwork {
	if m.znnNetwork == nil {
		panic(errors.New("znn network not init"))
	}
	return m.znnNetwork
}
