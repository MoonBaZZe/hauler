package network

import (
	"fmt"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/db/manager"
	"github.com/MoonBaZZe/hauler/rpc"
	"github.com/btcsuite/btcd/btcutil"
	"go.uber.org/zap"
	"os"
	"time"
)

type BtcNetwork struct {
	dbManager      *manager.Manager
	rpcManager     *rpc.Manager
	networkManager *NetworksManager
	state          *common.GlobalState
	stopChan       chan os.Signal
	logger         *zap.SugaredLogger
	tssAddress     btcutil.Address
}

func NewBtcNetwork(rpcManager *rpc.Manager, dbManager *manager.Manager, networkManager *NetworksManager, state *common.GlobalState, stopChan chan os.Signal, tssAddress btcutil.Address) (*BtcNetwork, error) {
	newLogger, errLog := common.CreateSugarLogger()
	if errLog != nil {
		return nil, errLog
	}

	newBtcNetwork := &BtcNetwork{
		rpcManager:     rpcManager,
		dbManager:      dbManager,
		networkManager: networkManager,
		state:          state,
		stopChan:       stopChan,
		logger:         newLogger,
		tssAddress:     tssAddress,
	}
	return newBtcNetwork, nil
}

func (bN *BtcNetwork) Start() error {
	fmt.Println("btcNetwork start 0")
	// Set the initial best block
	bestBlockHash, err := bN.rpcManager.Btc().GetBestBlockHash()
	if err != nil {
		return err
	}
	fmt.Println("btcNetwork start 1")
	bestBlockHeaderRpc, err := bN.rpcManager.Btc().GetBlockHeader(bestBlockHash)
	if err != nil {
		return err
	}
	fmt.Println("btcNetwork start 2")
	if err := bN.state.SetBestBlockHeader(bestBlockHeaderRpc); err != nil {
		return err
	}

	fmt.Println("btcNetwork start 3")
	go bN.UpdateBestBlock()
	return nil
}

func (bN *BtcNetwork) UpdateBestBlock() {
	for {
		time.Sleep(5 * time.Second)
		bestBlockHash, err := bN.rpcManager.Btc().GetBestBlockHash()
		if err != nil {
			bN.logger.Debugf("rpcManager.Btc().GetBestBlockHash error: %s", err.Error())
			continue
		}
		fmt.Printf("bestBlock: %s", bestBlockHash.String())
		bestBlockState, err := bN.state.GetBestBlockHeader()
		if err != nil {
			bN.logger.Debugf("state.GetBestBlockHeader error: %s", err.Error())
			continue
		} else if bestBlockState == nil {
			bestBlockHeaderRpc, err := bN.rpcManager.Btc().GetBlockHeader(bestBlockHash)
			if err != nil {
				bN.logger.Debugf("rpcManager.Btc().GetBlockHeader error: %s", err.Error())
				continue
			}
			if err := bN.state.SetBestBlockHeader(bestBlockHeaderRpc); err != nil {
				bN.logger.Debugf("bN.state.SetBestBlockHeader(bestBlockHeaderRpc) error: %s", err.Error())
				continue
			}
		}

		if bestBlockHash.String() != bestBlockState.Hash.String() {
			bestBlockHeaderRpc, err := bN.rpcManager.Btc().GetBlockHeader(bestBlockHash)
			if err != nil {
				bN.logger.Debugf("rpcManager.Btc().GetBlockHeader error: %s", err.Error())
				continue
			}
			if err := bN.state.SetBestBlockHeader(bestBlockHeaderRpc); err != nil {
				bN.logger.Debugf("bN.state.SetBestBlockHeader(bestBlockHeaderRpc) error: %s", err.Error())
				continue
			}
		}
	}
}

// Todo set extra data
func (bN *BtcNetwork) GetCoinBaseTx(extraNonce uint64) (*btcutil.Tx, error) {
	bestBlock, err := bN.state.GetBestBlockHeader()
	if err != nil {
		return nil, err
	}

	script, err := common.StandardCoinbaseScript(bestBlock.Height+1, extraNonce)
	if err != nil {
		return nil, err
	}

	coinbaseTx, err := common.CreateCoinbaseTx(script, bestBlock.Height+1, bN.tssAddress)
	if err != nil {
		return nil, err
	}
	// todo add extra data
	return coinbaseTx, nil
}

//func (bN *BtcNetwork) Create(extraNonce uint64) (*btcutil.Tx, error) {}
