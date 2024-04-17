package app

import (
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/node"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/signal"
	"syscall"
)

type Manager struct {
	ctx    *cli.Context
	node   *node.Node
	logger *zap.Logger
}

func NewNodeManager(ctx *cli.Context) (*Manager, error) {
	// make config
	nodeConfig, err := MakeConfig()
	if err != nil {
		return nil, err
	}

	logger, err := common.CreateLogger()
	if err != nil {
		return nil, err
	}

	// make node
	newNode, err := node.NewNode(nodeConfig, logger)
	if err != nil {
		logger.Info("failed to create the node", zap.String("reason", err.Error()))
		return nil, err
	}

	return &Manager{
		ctx:    ctx,
		node:   newNode,
		logger: logger,
	}, nil
}

func (nodeManager *Manager) Start() error {
	// Start up the node
	nodeManager.logger.Info("starting hauler")
	if err := nodeManager.node.Start(); err != nil {
		nodeManager.logger.Fatal("failed to start node", zap.String("reason", err.Error()))
		os.Exit(1)
	} else {
		nodeManager.logger.Info("hauler successfully started")
	}

	signalFromOutside := false
	// Listening event closes the node
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		defer signal.Stop(c)
		<-c
		signalFromOutside = true
		nodeManager.logger.Info("Shutting down hauler from go func")

		go func() {
			if err := nodeManager.Stop(); err != nil {
				nodeManager.logger.Error(err.Error())
			}
		}()

	}()

	// Waiting for node to close
	nodeManager.node.Wait()
	nodeManager.logger.Info("signalFromOutside: ", zap.Bool("value: ", signalFromOutside))
	if signalFromOutside == false {
		if err := nodeManager.Stop(); err != nil {
			nodeManager.logger.Info(err.Error())
		}
	}

	return nil
}
func (nodeManager *Manager) Stop() error {
	nodeManager.logger.Warn("Stopping hauler ...")

	if err := nodeManager.SaveConfig(); err != nil {
		nodeManager.logger.Info("Failed to save config", zap.String("reason", err.Error()))
	}

	if err := nodeManager.node.Stop(); err != nil {
		nodeManager.logger.Info("Failed to stop node", zap.String("reason", err.Error()))
	} else {
		nodeManager.logger.Info("successfully stopped node")
	}
	return nil
}

func (nodeManager *Manager) SaveConfig() error {
	nodeManager.logger.Info("Write config to file")
	conf := nodeManager.node.GetConfig()
	if conf != nil {
		nodeManager.logger.Info("wrote config at the end")
		return common.WriteConfig(*conf)
	}
	return errors.New("Invalid config")
}
