package network

import (
	"encoding/base64"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/db/manager"
	"github.com/MoonBaZZe/hauler/rpc"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/pkg/errors"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
	"github.com/zenon-network/go-zenon/wallet"
	"go.uber.org/zap"

	"os"
)

type NetworksManager struct {
	znnNetwork *ZnnNetwork
	btcNetwork *BtcNetwork
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

func (m *NetworksManager) Init(config *common.Config, dbManager *manager.Manager, producerKeyPair *wallet.KeyPair, state *common.GlobalState) error {
	newRpcManager, err := rpc.NewRpcManager(config, m.stopChan)
	if err != nil {
		return err
	}

	newZnnNetwork, err := NewZnnNetwork(newRpcManager, dbManager, m, producerKeyPair, state, m.stopChan)
	if err != nil {
		return err
	}
	m.znnNetwork = newZnnNetwork

	//mergeMiningInfo, err := newRpcManager.Znn().GetMergeMiningInfo()
	//if err != nil {
	//	return err
	//}

	// todo remove after setting up a testnet
	mergeMiningInfo := &definition.MergeMiningInfoVariable{
		DecompressedTssECDSAPubKey: "BMAQx1M3LVXCuozDOqO5b9adj/PItYgwZFG/xTDBiZzTnQAT1qOPAkuPzu6yoewss9XbnTmZmb9JQNGXmkPYtK4=",
	}

	pubKey, err := base64.StdEncoding.DecodeString(mergeMiningInfo.DecompressedTssECDSAPubKey)
	if err != nil {
		return constants.ErrInvalidB64Decode
	}
	if len(pubKey) != constants.DecompressedECDSAPubKeyLength {
		return constants.ErrInvalidDecompressedECDSAPubKeyLength
	}

	net := &chaincfg.Params{
		PubKeyHashAddrID: 0x00,
	}
	tssAddress, err := btcutil.NewAddressPubKey(pubKey, net)
	if err != nil {
		return err
	}
	newBtcNetwork, err := NewBtcNetwork(newRpcManager, dbManager, m, state, m.stopChan, tssAddress)
	if err != nil {
		return err
	}
	m.btcNetwork = newBtcNetwork

	return nil
}

func (m *NetworksManager) Start() error {
	if err := m.znnNetwork.Start(); err != nil {
		return err
	}

	if err := m.btcNetwork.Start(); err != nil {
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
