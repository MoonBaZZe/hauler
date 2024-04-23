package rpc

import (
	"github.com/btcsuite/btcd/btcjson"
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

func (b *BtcRpc) GetMemPool() (map[string]btcjson.GetRawMempoolVerboseResult, error) {
	return b.rpcClient.GetRawMempoolVerbose()
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
