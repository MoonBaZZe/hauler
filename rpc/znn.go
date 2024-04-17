package rpc

import (
	common2 "github.com/MoonBaZZe/hauler/common"
	sdk_rpc_client "github.com/MoonBaZZe/znn-sdk-go/rpc_client"
	"github.com/MoonBaZZe/znn-sdk-go/zenon"
	"github.com/pkg/errors"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/protocol"
	"github.com/zenon-network/go-zenon/rpc/api"
	"github.com/zenon-network/go-zenon/rpc/api/subscribe"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
	"github.com/zenon-network/go-zenon/wallet"
)

type ZnnRpc struct {
	rpcClient *sdk_rpc_client.RpcClient
	Urls      common2.UrlsInfo

	momentumsSub  *server.ClientSubscription
	momentumsChan chan []subscribe.Momentum

	accountBlocksSub  *server.ClientSubscription
	accountBlocksChan chan []subscribe.AccountBlock
}

func NewZnnRpcClient(urls []string) (*ZnnRpc, error) {
	newUrls, err := common2.NewUrlsInfo(urls)
	if err != nil {
		return nil, err
	}
	var newZnnClient *sdk_rpc_client.RpcClient
	currentUrl := newUrls.GetCurrentUrl()
	for {
		newZnnClient, err = sdk_rpc_client.NewRpcClient(currentUrl)
		if err != nil {
			common2.GlobalLogger.Infof("Error when dialing %s, got: %s\n", currentUrl, err)
		} else {
			break
		}
		currentUrl = newUrls.NextUrl()
		if len(currentUrl) == 0 {
			return nil, errors.New("cannot connect to any urls to a znn node")
		}
	}

	newUrls.Clear()
	return &ZnnRpc{
		rpcClient: newZnnClient,
		Urls:      *newUrls,
	}, nil
}

/// Utils

func (r *ZnnRpc) Stop() error {
	close(r.momentumsChan)
	close(r.accountBlocksChan)
	if r.rpcClient == nil {
		return errors.New("znn rpc client is nil")
	}
	r.rpcClient.Stop()
	return nil
}

func (r *ZnnRpc) Reconnect() error {
	// todo?this should have a semaphore
	return nil
}

/// Subscriptions

func (r *ZnnRpc) SubscribeToMomentums() (*server.ClientSubscription, chan []subscribe.Momentum, error) {
	var err error
	r.momentumsSub, r.momentumsChan, err = r.rpcClient.SubscriberApi.ToMomentums()
	if err != nil {
		return nil, nil, err
	}
	return r.momentumsSub, r.momentumsChan, nil
}

func (r *ZnnRpc) SubscribeToAccountBlocks(address types.Address) (*server.ClientSubscription, chan []subscribe.AccountBlock, error) {
	var err error
	r.accountBlocksSub, r.accountBlocksChan, err = r.rpcClient.SubscriberApi.ToAccountBlocksByAddress(address)
	if err != nil {
		return nil, nil, err
	}
	return r.accountBlocksSub, r.accountBlocksChan, nil
}

/// Transactions

func (r *ZnnRpc) BroadcastTransaction(tx *nom.AccountBlock, keyPair *wallet.KeyPair) error {
	if err := zenon.CheckAndSetFields(r.rpcClient, tx, keyPair.Address, keyPair.Public); err != nil {
		return err
	}
	if err := zenon.SetDifficulty(r.rpcClient, tx); err != nil {
		return err
	}

	tx.Hash = tx.ComputeHash()
	tx.Signature = keyPair.Sign(tx.Hash.Bytes())

	return r.rpcClient.LedgerApi.PublishRawTransaction(tx)
}

/// RPC Calls

func (r *ZnnRpc) GetBridgeInfo() (*definition.BridgeInfoVariable, error) {
	return r.rpcClient.BridgeApi.GetBridgeInfo()
}

func (r *ZnnRpc) GetSyncInfo() (*protocol.SyncInfo, error) {
	return r.rpcClient.StatsApi.SyncInfo()
}

func (r *ZnnRpc) GetAccountBlockByHash(hash types.Hash) (*api.AccountBlock, error) {
	return r.rpcClient.LedgerApi.GetAccountBlockByHash(hash)
}

func (r *ZnnRpc) GetAccountBlocksByHeight(address types.Address, height, count uint64) (*api.AccountBlockList, error) {
	return r.rpcClient.LedgerApi.GetAccountBlocksByHeight(address, height, count)
}

func (r *ZnnRpc) GetMomentumsByHeight(height, count uint64) (*api.MomentumList, error) {
	return r.rpcClient.LedgerApi.GetMomentumsByHeight(height, count)
}

func (r *ZnnRpc) GetFrontierMomentum() (*api.Momentum, error) {
	return r.rpcClient.LedgerApi.GetFrontierMomentum()
}

func (r *ZnnRpc) GetSecurityInfo() (*definition.SecurityInfoVariable, error) {
	return r.rpcClient.BridgeApi.GetSecurityInfo()
}
