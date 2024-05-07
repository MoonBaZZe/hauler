package network

import (
	"context"
	"fmt"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/db/manager"
	"github.com/MoonBaZZe/hauler/rpc"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
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
	// Set the initial best block
	bestBlockHash, err := bN.rpcManager.Btc().GetBestBlockHash()
	if err != nil {
		return err
	}

	bestBlockHeaderRpc, err := bN.rpcManager.Btc().GetBlockHeader(bestBlockHash)
	if err != nil {
		return err
	}

	if err := bN.state.SetBestBlockHeader(bestBlockHeaderRpc); err != nil {
		return err
	}

	//go bN.UpdateMemPool()
	go bN.UpdateBestBlock()
	go bN.UpdateTransactions()
	go bN.UpdateHeaderChain()
	return nil
}

func (bN *BtcNetwork) UpdateHeaderChain() {
	for {
		bestBlock, err := bN.state.GetBestBlockHeader()
		if err != nil {
			bN.logger.Debugf("%v\n", err)
			bN.stopChan <- syscall.SIGINT
		}

		headerChainInfo, errHeader := bN.rpcManager.Znn().GetHeaderChainInfo()
		if errHeader != nil {
			bN.logger.Debugf("%v\n", errHeader)
		} else {
			if bestBlock.Hash.String() == headerChainInfo.Tip.String() {

			} else if bestBlock.PrevBlock.String() == headerChainInfo.Tip.String() {
				if errAdd := bN.rpcManager.Znn().AddBitcoinBlockHeader(bestBlock, bN.networkManager.Znn().GetProducerKeyPair()); errAdd != nil {
					bN.logger.Debugf("%v\n", errAdd)
				}
			} else {
				// todo we need to add multiple blocks
			}
		}

		time.Sleep(10 * time.Second)
	}

}

func (bN *BtcNetwork) UpdateTransactions() {
	for {
		time.Sleep(15 * time.Second)

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
				fmt.Printf("storage get tx error: %s\n", err.Error())
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

// todo update mempool as long as we have miners connected
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
			bN.logger.Debugf("rpcManager.Btc().GetBestBlockHash error: %s\n", err.Error())
			continue
		}
		fmt.Printf("bestBlock: %s\n", bestBlockHash.String())
		bestBlockState, err := bN.state.GetBestBlockHeader()
		if err != nil {
			bN.logger.Debugf("state.GetBestBlockHeader error: %s\n", err.Error())
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

func (bN *BtcNetwork) GetCurrentTemplateTransactions(coinbaseTx *btcutil.Tx, msgTxs []*wire.MsgTx) ([]*btcutil.Tx, bool) {
	// todo implement logic to choose the most profitable transactions
	transactions := make([]*btcutil.Tx, 0)
	transactions = append(transactions, coinbaseTx)

	hasWitness := false
	for _, msgTx := range msgTxs {
		tx := btcutil.NewTx(msgTx)
		hasWitness = hasWitness || tx.HasWitness()
		transactions = append(transactions, tx)
	}

	return transactions, hasWitness
}

// todo
func (bN *BtcNetwork) GetMerkleRootAndZkProofs(txs []*btcutil.Tx, hasWitness bool) *chainhash.Hash {
	merkleTree := blockchain.BuildMerkleTreeStore(txs, hasWitness)
	merkleRoot := merkleTree[len(merkleTree)-1]

	leaves := make([]*chainhash.Hash, 0)
	// Append coinbase tx and its sibling
	leaves = append(leaves, merkleTree[0])
	leaves = append(leaves, merkleTree[1])

	// Start from the hash of leaves 2 and 3
	lenTxs := len(txs)
	index := lenTxs + 1

	for index < len(merkleTree)-1 {
		leaves = append(leaves, merkleTree[index])
		// Calculate the next index, assuming a binary tree, moving to the next level
		index += len(txs) / 2
		lenTxs /= 2
	}

	//mod := ecc.BN254.ScalarField()
	//modNbBytes := len(mod.Bytes())

	//hasher := hash.MIMC_BN254
	//hGo := hasher.New()
	//nrLeaves := 15
	//proofIndex := uint64(0)
	//var l []byte
	//depth := 4
	//
	//var buf bytes.Buffer
	//for i := 0; i < nrLeaves; i++ {
	//	leaf, err := rand.Int(rand.Reader, mod)
	//	assert.NoError(err)
	//	b := leaf.Bytes()
	//	if i == int(proofIndex) {
	//		l = b
	//		fmt.Printf("leaf len: %d\n", len(l))
	//		fmt.Printf("leaf: %s\n", leaf.String())
	//	}
	//	buf.Write(make([]byte, modNbBytes-len(b)))
	//	buf.Write(b)
	//}

	return merkleRoot
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

func (bN *BtcNetwork) CreateTemplate(version int32, prevBlock, merkleRoot chainhash.Hash, timestamp, bits uint32) *wire.BlockHeader {
	return &wire.BlockHeader{
		Version:    version,
		PrevBlock:  prevBlock,
		MerkleRoot: merkleRoot,
		Timestamp:  time.Unix(int64(timestamp), 0),
		Bits:       bits,
		//Nonce:      nonce,
	}
}

//func (bN *BtcNetwork) Create(extraNonce uint64) (*btcutil.Tx, error) {}
