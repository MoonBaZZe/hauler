package rpc

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

type BtcRpc struct {
	rpcClient *rpcclient.Client
}

func NewBtcRpcClient(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (*BtcRpc, error) {
	newBtcClient, err := rpcclient.New(config, ntfnHandlers)
	if err != nil {
		return nil, err
	}
	return &BtcRpc{
		rpcClient: newBtcClient,
	}, nil
}

func (b *BtcRpc) GetRawMemPool() ([]*chainhash.Hash, error) {
	return b.rpcClient.GetRawMempool()
}

func (b *BtcRpc) GetBestBlockHash() (*chainhash.Hash, error) {
	return b.rpcClient.GetBestBlockHash()
}

func (b *BtcRpc) GetBlockHeader(hash *chainhash.Hash) (*wire.BlockHeader, error) {
	return b.rpcClient.GetBlockHeader(hash)
}

func (b *BtcRpc) GetBlockCount() (int64, error) {
	return b.rpcClient.GetBlockCount()
}

func (b *BtcRpc) GetRawTransaction(hash *chainhash.Hash) (*btcutil.Tx, error) {
	return b.rpcClient.GetRawTransaction(hash)
}
