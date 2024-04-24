package network

import (
	"context"
	"fmt"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/db/manager"
	"github.com/MoonBaZZe/hauler/rpc"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"os"
	"syscall"
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
	// Use a semaphore for accessing it
	memPool          map[string]bool
	memPoolSemaphore *semaphore.Weighted
}

func NewBtcNetwork(rpcManager *rpc.Manager, dbManager *manager.Manager, networkManager *NetworksManager, state *common.GlobalState, stopChan chan os.Signal, tssAddress btcutil.Address) (*BtcNetwork, error) {
	newLogger, errLog := common.CreateSugarLogger()
	if errLog != nil {
		return nil, errLog
	}

	newBtcNetwork := &BtcNetwork{
		rpcManager:       rpcManager,
		dbManager:        dbManager,
		networkManager:   networkManager,
		state:            state,
		stopChan:         stopChan,
		logger:           newLogger,
		tssAddress:       tssAddress,
		memPool:          make(map[string]bool),
		memPoolSemaphore: semaphore.NewWeighted(1),
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

func (bN *BtcNetwork) UpdateTransactions() {
	for {
		time.Sleep(2 * time.Second)

		// Create a copy for the current mempool so we don't block the resource while downloading the transactions
		memPoolHashes := make([]string, 0)
		if err := bN.memPoolSemaphore.Acquire(context.Background(), 1); err != nil {
			bN.logger.Debugf("could not acquire mem pool semaphore\n")
			continue
		}
		for k, _ := range bN.memPool {
			memPoolHashes = append(memPoolHashes, k)
		}
		bN.memPoolSemaphore.Release(1)

		for _, txHash := range memPoolHashes {
			hash, err := chainhash.NewHashFromStr(txHash)
			if err != nil {
				bN.logger.Debugf("%v\n", err)
				continue
			}

			// Check locally, otherwise we ask it from the rpc endpoint
			if _, err := bN.dbManager.BtcStorage().GetTransaction(*hash); err != nil {
				if errors.Is(err, leveldb.ErrNotFound) {
					// we will download it
					tx, err := bN.rpcManager.Btc().GetRawTransaction(hash)
					if err != nil {
						bN.logger.Debugf("%v\n", err)
						continue
					}

					if err := bN.dbManager.BtcStorage().AddTransaction(tx.MsgTx()); err != nil {
						bN.logger.Debugf("%v\n", err)
						continue
					}
				} else {
					bN.logger.Info("sent SIGINT from here 5")
					bN.logger.Debugf("%v\n", err)
					bN.stopChan <- syscall.SIGINT
				}
			}
		}
	}
}

func (bN *BtcNetwork) UpdateMemPool() {
	shouldSleep := false
	for {
		if shouldSleep {
			time.Sleep(60 * time.Second)
		}

		rawMemPool, err := bN.rpcManager.Btc().GetRawMemPool()
		if err != nil {
			bN.logger.Debugf("bN.rpcManager.Btc().GetRawMemPool(): %s\n", err.Error())
			shouldSleep = true
			continue
		}
		if err := bN.memPoolSemaphore.Acquire(context.Background(), 1); err != nil {
			bN.logger.Debugf("could not acquire mem pool semaphore\n")
			shouldSleep = true
			continue
		}
		bN.memPool = make(map[string]bool)
		for _, tx := range rawMemPool {
			bN.memPool[tx.String()] = true
		}
		bN.memPoolSemaphore.Release(1)
		shouldSleep = true
	}
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
