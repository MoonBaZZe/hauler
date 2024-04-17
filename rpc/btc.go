package rpc

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
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

func (b *BtcRpc) GetBestBlock() (*chainhash.Hash, int32, error) {
	return b.rpcClient.GetBestBlock()
}

func (b *BtcRpc) GetBlockCount() (int64, error) {
	return b.rpcClient.GetBlockCount()
}

func (b *BtcRpc) H() {
	//b.rpcClient.
}
