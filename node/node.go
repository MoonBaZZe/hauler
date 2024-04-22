package node

import (
	"errors"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/db/manager"
	"github.com/MoonBaZZe/hauler/network"
	wallet2 "github.com/MoonBaZZe/znn-sdk-go/wallet"
	"github.com/prometheus/tsdb/fileutil"
	"github.com/zenon-network/go-zenon/wallet"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type Node struct {
	config *common.Config

	dbManager       *manager.Manager
	networksManager *network.NetworksManager
	state           *common.GlobalState

	producerKeyPair *wallet.KeyPair
	logger          *zap.SugaredLogger
	// Channel to wait for termination notifications
	stopChan chan os.Signal
	lock     sync.RWMutex
	// Prevents concurrent use of instance directory
	dataDirLock fileutil.Releaser
}

func NewNode(config *common.Config, logger *zap.Logger) (*Node, error) {
	var err error

	node := &Node{
		config:   config,
		state:    common.NewGlobalState(&config.GlobalState), // todo set default?
		logger:   logger.Sugar(),
		stopChan: make(chan os.Signal, 1),
	}

	// prepare node
	node.logger.Info("preparing node ... ")
	if err = node.openDataDir(); err != nil {
		return nil, err
	}

	if node.dbManager, err = manager.NewDbManager(node.stopChan); err != nil {
		return nil, err
	}

	node.networksManager, err = network.NewNetworksManager(node.stopChan)
	if err != nil {
		return nil, err
	}
	if errInit := node.networksManager.Init(config, node.dbManager, node.state); errInit != nil {
		return nil, errInit
	}

	// init btc.go rpc
	newKeyStore, err := wallet2.ReadKeyFile(config.ProducerKeyFileName, config.ProducerKeyFilePassphrase, path.Join(config.DataPath, config.ProducerKeyFileName))
	if err != nil {
		return nil, err
	}
	node.logger.Info("read producer")
	_, node.producerKeyPair, err = newKeyStore.DeriveForIndexPath(config.ProducerIndex)
	if err != nil {
		return nil, err
	}
	if len(newKeyStore.Entropy) == 0 {
		return nil, errors.New("entropy cannot be nil")
	}

	for node.networksManager.Znn().IsSynced() == false {
		node.logger.Info("node is syncing, will wait for it to finish")
		time.Sleep(15 * time.Second)
	}

	if err := common.WriteConfig(*node.config); err != nil {
		node.logger.Info(err.Error())
	}
	return node, nil
}

func (node *Node) Start() error {
	node.lock.Lock()
	defer node.lock.Unlock()

	frMom, frMomErr := node.networksManager.Znn().GetFrontierMomentum()
	if frMomErr != nil {
		return frMomErr
	} else if frMom == nil {
		return errors.New("frontier momentum is nil")
	} else {
		if errState := node.state.SetFrontierMomentum(frMom.Height); errState != nil {
			return errState
		}
	}

	if err := node.networksManager.Start(); err != nil {
		return err
	}

	//m, err := node.BtcClient.GetBlockCount()
	//if err != nil {
	//	return err
	//}
	//fmt.Println(m)

	//m, err := node.BtcClient.GetMemPool()
	//if err != nil {
	//	return err
	//}
	//for k, v := range m {
	//	fmt.Println(k, v)
	//}
	return nil
}

func (node *Node) Stop() error {
	defer close(node.stopChan)
	return nil
}

func (node *Node) Wait() {
	signalReceived := <-node.stopChan
	node.logger.Info("signal from wait: ", signalReceived)
}

func (node *Node) WatchBestBlock() {
	for {
		//bestBlock, height, err := node.BtcClient.GetBestBlock()
	}
}

func (node *Node) GetConfig() *common.Config {
	return node.config
}

func (node *Node) openDataDir() error {
	if node.config.DataPath == "" {
		return nil
	}

	if err := os.MkdirAll(node.config.DataPath, 0700); err != nil {
		return err
	}
	node.logger.Info("successfully ensured DataPath exists", zap.String("data-path", node.config.DataPath))

	// Lock the instance directory to prevent concurrent use by another instance as well as
	// accidental use of the instance directory as a database.
	if fileLock, _, err := fileutil.Flock(filepath.Join(node.config.DataPath, ".lock")); err != nil {
		node.logger.Info("unable to acquire file-lock", zap.String("reason", err.Error()))
		return convertFileLockError(err)
	} else {
		node.dataDirLock = fileLock
	}

	node.logger.Info("successfully locked dataDir")
	return nil
}
