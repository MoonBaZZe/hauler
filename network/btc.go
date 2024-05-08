package network

import (
	"bytes"
	"context"
	"fmt"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/common/block_header"
	"github.com/MoonBaZZe/hauler/db/manager"
	"github.com/MoonBaZZe/hauler/rpc"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/consensys/gnark-crypto/accumulator/merkletree"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zenon-network/go-zenon/common/types"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"math"
	"math/rand"
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

// Todo: how to implement the logic so that only one / a few haulers update the headerChain and not everyone. Is the random sleep enough?

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
			if bestBlock.Hash.String() != headerChainInfo.Tip.String() {
				if bestBlock.PrevBlock.String() == headerChainInfo.Tip.String() {
					if errAdd := bN.rpcManager.Znn().AddBitcoinBlockHeader(bestBlock, bN.networkManager.Znn().GetProducerKeyPair()); errAdd != nil {
						bN.logger.Debugf("%v\n", errAdd)
						break
					}
					bN.logger.Debugf("Added block header %s\n", bestBlock.Hash.String())
				} else {
					verboseHash, err := chainhash.NewHashFromStr(headerChainInfo.Tip.String())
					if err != nil {
						bN.logger.Debugf("%v\n", err)
						break
					}
					verboseBlock, err := bN.rpcManager.Btc().GetBlockVerbose(verboseHash)
					if err != nil {
						bN.logger.Debugf("%v\n", err)
						break
					}
					currentCount, err := bN.rpcManager.Btc().GetBlockCount()
					if err != nil {
						bN.logger.Debugf("%v\n", err)
						break
					}
					for i := verboseBlock.Height + 1; i <= currentCount; i++ {
						blockHash, err := bN.rpcManager.Btc().GetBlockHash(i)
						if err != nil {
							bN.logger.Debugf("%v\n", err)
							break
						}
						block, err := bN.rpcManager.Btc().GetBlockHeader(blockHash)
						if err != nil {
							bN.logger.Debugf("%v\n", err)
							break
						}

						blockHeader := &block_header.BlockHeader{
							Version:    block.Version,
							PrevBlock:  types.HexToHashPanic(block.PrevBlock.String()),
							MerkleRoot: types.HexToHashPanic(block.MerkleRoot.String()),
							Timestamp:  uint32(block.Timestamp.Unix()),
							Bits:       block.Bits,
							Nonce:      block.Nonce,
						}
						if errAdd := bN.rpcManager.Znn().AddBitcoinBlockHeader(blockHeader, bN.networkManager.Znn().GetProducerKeyPair()); errAdd != nil {
							bN.logger.Debugf("%v\n", errAdd)
							break
						}
						bN.logger.Debugf("Added block header %s\n", block.BlockHash())
						time.Sleep(11 * time.Second)
					}
				}
			}
		}
		seconds := time.Duration(rand.Uint32()%20 + 5)
		time.Sleep(seconds * time.Second)
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
		shouldSleep = true
		bN.memPoolSemaphore.Release(1)
	}
}

func (bN *BtcNetwork) UpdateBestBlock() {
	for {
		time.Sleep(7 * time.Second)
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
		} else if bestBlockState == nil || (bestBlockHash.String() != bestBlockState.Hash.String()) {
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

func (bN *BtcNetwork) GetCurrentTemplateTransactions(coinbaseTx *btcutil.Tx) ([]*btcutil.Tx, bool, error) {
	// todo implement logic to choose the most profitable transactions

	// Create a copy
	if err := bN.memPoolSemaphore.Acquire(context.Background(), 1); err != nil {
		bN.logger.Debugf("could not acquire mem pool semaphore\n")
		return nil, false, err
	}
	tempMemPool := make(map[string]bool)
	for k, _ := range bN.memPool {
		tempMemPool[k] = true
	}
	bN.memPoolSemaphore.Release(1)

	msgTxs := make([]*wire.MsgTx, 0)
	count := 0
	for k, _ := range tempMemPool {
		hash, errHash := chainhash.NewHashFromStr(k)
		if errHash != nil {
			bN.logger.Debugf("could not create hash\n")
			continue
		}
		if tx, err := bN.dbManager.BtcStorage().GetTransaction(*hash); err != nil {
			bN.logger.Debugf("%v\n", err)
			continue
		} else {
			msgTxs = append(msgTxs, tx)
		}
		// todo remove
		count++
		if count > 10 {
			break
		}
	}

	transactions := make([]*btcutil.Tx, 0)
	transactions = append(transactions, coinbaseTx)

	hasWitness := false
	for _, msgTx := range msgTxs {
		tx := btcutil.NewTx(msgTx)
		hasWitness = hasWitness || tx.HasWitness()
		transactions = append(transactions, tx)
	}

	return transactions, hasWitness, nil
}

func (bN *BtcNetwork) GetMerkleRootAndZkProofs(txs []*btcutil.Tx, hasWitness bool) *chainhash.Hash {
	merkleTree := blockchain.BuildMerkleTreeStore(txs, hasWitness)
	merkleRoot := merkleTree[len(merkleTree)-1]

	leaves := make([]*chainhash.Hash, 0)
	// Append coinbase tx
	index := 0
	leaves = append(leaves, merkleTree[index])

	// Select only the required hashes needed to prove that the coinbase tx belongs to the root
	nextPoT := common.NextPowerOfTwo(len(txs))
	arraySize := nextPoT*2 - 1
	for index < arraySize-1 {
		if index%2 == 0 {
			leaves = append(leaves, merkleTree[index+1])
		} else {
			leaves = append(leaves, merkleTree[index-1])
		}
		index += nextPoT
		nextPoT /= 2
	}

	mod := ecc.BN254.ScalarField()
	modNbBytes := len(mod.Bytes())
	hasher := hash.MIMC_BN254
	hGo := hasher.New()
	nrLeaves := len(leaves)
	proofIndex := uint64(0)
	var l []byte
	depth := uint(math.Log2(float64(nrLeaves))) + 1

	var buf bytes.Buffer
	for i := 0; i < nrLeaves; i++ {
		b := leaves[i].CloneBytes()
		if i == int(proofIndex) {
			l = b
			fmt.Printf("leaf len: %d\n", len(l))
			fmt.Printf("leaf: %s\n", leaves[i].String())
		}
		buf.Write(make([]byte, modNbBytes-len(b)))
		buf.Write(b)
	}

	// Create proof
	merkleRootProof, proofPath, numLeaves, err := merkletree.BuildReaderProof(&buf, hGo, modNbBytes, proofIndex)
	if err != nil {
		//t.Fatal("error creating Merkle Proof")
	}
	// Check proof
	fmt.Printf("len(merkleRoot): %d\n", len(merkleRootProof))
	fmt.Printf("len(proofPath): %d\n", len(proofPath))
	//for _, p := range proofPath {
	//	fmt.Printf("len(proof): %d\n", len(p))
	//}

	verified := merkletree.VerifyProof(hGo, merkleRootProof, proofPath, proofIndex, numLeaves)
	if !verified {
		//t.Fatal("The created Merkle Proof is not valid")
	}

	//var mtCircuit zcommon.MTCircuit
	//var witness zcommon.MTCircuit
	//mtCircuit.ProofElements = make([]frontend.Variable, depth)
	//witness.ProofElements = make([]frontend.Variable, depth)
	//// skip elm 0 (in proofPath) since it's the leaf hash and we calculate it ourselves
	//for i := 0; i < depth; i++ {
	//	witness.ProofElements[i] = proofPath[i+1]
	//}
	//witness.ProofIndex = proofIndex
	//witness.Root = merkleRoot
	//witness.Leaf = proofPath[0]
	//fmt.Printf("Before prover succeeded\n")
	//
	//assert.ProverSucceeded(&mtCircuit, &witness, test.WithCurves(ecc.BN254))

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
