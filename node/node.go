package node

import (
	"errors"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/rpc"
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
	config          *common.Config
	producerKeyPair *wallet.KeyPair
	logger          *zap.SugaredLogger

	ZnnClient *rpc.ZnnRpc

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
		logger:   logger.Sugar(),
		stopChan: make(chan os.Signal, 1),
	}

	node.ZnnClient, err = rpc.NewZnnRpcClient(config.NoMEndpoints)
	if err != nil {
		return nil, err
	}

	// prepare node
	node.logger.Info("preparing node ... ")
	if err = node.openDataDir(); err != nil {
		return nil, err
	}

	// init btc rpc

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

	for node.ZnnClient.IsSynced() == false {
		node.logger.Info("node is syncing, will wait for it to finish")
		time.Sleep(15 * time.Second)
	}

	if err := common.WriteConfig(*node.config); err != nil {
		node.logger.Info(err.Error())
	}
	return node, nil
}

func (node *Node) Start() error {
	return nil
}

func (node *Node) Stop() error {
	return nil
}

func (node *Node) Wait() error {
	return nil
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
