package network

import (
	"encoding/base64"
	"fmt"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/common/block_header"
	"github.com/MoonBaZZe/hauler/db"
	"github.com/MoonBaZZe/hauler/db/manager"
	"github.com/MoonBaZZe/hauler/rpc"
	"github.com/pkg/errors"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/protocol"
	"github.com/zenon-network/go-zenon/rpc/api"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
	"github.com/zenon-network/go-zenon/wallet"
	"go.uber.org/zap"
	"math/big"
	"os"
	"syscall"
	"time"
)

type ZnnNetwork struct {
	dbManager       *manager.Manager
	rpcManager      *rpc.Manager
	networkManager  *NetworksManager
	producerKeyPair *wallet.KeyPair
	state           *common.GlobalState
	stopChan        chan os.Signal
	logger          *zap.SugaredLogger
}

// CheckSecurityInfoInitialized this method should have the same checks as in go-zenon
func CheckSecurityInfoInitialized(securityInfo *definition.SecurityInfoVariable) error {
	if len(securityInfo.Guardians) < constants.MinGuardians {
		return errors.New("SecurityInfo not initialised")
	}
	return nil
}

func NewZnnNetwork(rpcManager *rpc.Manager, dbManager *manager.Manager, networkManager *NetworksManager, producerKeyPair *wallet.KeyPair, state *common.GlobalState, stopChan chan os.Signal) (*ZnnNetwork, error) {
	securityInfo, err := rpcManager.Znn().GetSecurityInfo()
	if err != nil {
		return nil, err
	} else if securityErr := CheckSecurityInfoInitialized(securityInfo); securityErr != nil {
		return nil, securityErr
	}

	newLogger, errLog := common.CreateSugarLogger()
	if errLog != nil {
		return nil, errLog
	}

	newZnnNetwork := &ZnnNetwork{
		rpcManager:      rpcManager,
		dbManager:       dbManager,
		networkManager:  networkManager,
		producerKeyPair: producerKeyPair,
		state:           state,
		stopChan:        stopChan,
		logger:          newLogger,
	}
	return newZnnNetwork, nil
}

/// Utils

func (rC *ZnnNetwork) Start() error {
	go rC.ListenForMomentumHeight()

	if err := rC.Sync(); err != nil {
		rC.logger.Debugf("error: %s", err.Error())
		return err
	}

	//go rC.ListenForEmbeddedBridgeAccountBlocks()
	return nil
}

func (rC *ZnnNetwork) Stop() error {
	return rC.ZnnRpc().Stop()
}
func (rC *ZnnNetwork) eventsStore() db.ZnnStorage {
	return rC.dbManager.ZnnStorage()
}
func (rC *ZnnNetwork) ZnnRpc() *rpc.ZnnRpc {
	return rC.rpcManager.Znn()
}

func (rC *ZnnNetwork) Sync() error {
	rC.logger.Debug("In sync znn")
	if accountBlockHeight, err := rC.eventsStore().GetLastUpdateHeight(); err != nil {
		return err
	} else {
		rC.logger.Debugf("last account block update height: %d", accountBlockHeight)
		accountBlockList, errRpc := rC.ZnnRpc().GetAccountBlocksByHeight(types.MergeMiningContract, accountBlockHeight+1, 30)
		if errRpc != nil {
			return errRpc
		}
		for len(accountBlockList.List) > 0 {
			for _, accBlock := range accountBlockList.List {
				if accBlock.BlockType == nom.BlockTypeContractReceive {
					rC.logger.Debug("found receive block")
					hash := accBlock.Hash
					for {
						rC.logger.Debugf("confDetail is nil: %v for %s", accBlock.ConfirmationDetail == nil, hash.String())
						accBlock, errRpc = rC.ZnnRpc().GetAccountBlockByHash(hash)
						if errRpc != nil {
							rC.logger.Debug(err)
						} else if accBlock == nil {

						} else if accBlock.ConfirmationDetail != nil {
							break
						}
						time.Sleep(5 * time.Second)
						continue
					}

					if sendBlock, errRpc := rC.ZnnRpc().GetAccountBlockByHash(accBlock.FromBlockHash); err != nil {
						return errRpc
					} else if sendBlock == nil {
						return errors.Errorf("Send block %s for associated receive %s is non existent", accBlock.Hash.String(), accBlock.FromBlockHash.String())
					} else {
						var live bool
						frMomHeight, errFrMom := rC.state.GetFrontierMomentum()
						if errFrMom != nil {
							return errFrMom
						}
						if frMomHeight < accBlock.ConfirmationDetail.MomentumHeight {
							return errors.New(fmt.Sprintf("frMomHeight %d cannot be less than the height of the momentum %d in which was included the acc block we process", frMomHeight, accBlock.ConfirmationDetail.MomentumHeight))
						}
						// todo do we need confirmations??
						live = (frMomHeight - accBlock.ConfirmationDetail.MomentumHeight) < uint64(1)
						live = live && rC.IsSynced()
						if newErr := rC.InterpretSendBlockData(sendBlock, live, accBlock.Height); newErr != nil {
							return newErr
						}
					}
				}
			}
			accountBlockHeight += uint64(len(accountBlockList.List))
			accountBlockList, err = rC.ZnnRpc().GetAccountBlocksByHeight(types.MergeMiningContract, accountBlockHeight+1, 30)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// InterpretSendBlockData We assume that if the receive acc block was created then it was no error
func (rC *ZnnNetwork) InterpretSendBlockData(sendBlock *api.AccountBlock, live bool, receiveBlockHeight uint64) error {
	rC.logger.Debugf("InterpretSendBlockData, live: %v", live)
	methodSig := base64.StdEncoding.EncodeToString(sendBlock.Data[0:4])
	switch methodSig {
	case base64.StdEncoding.EncodeToString(definition.ABIMergeMining.Methods[definition.AddBitcoinBlockHeaderMethodName].Id()):
		rC.logger.Debug("found AddBitcoinBlockHeaderMethodName")
		param := new(definition.BlockHeaderVariable)
		if err := definition.ABIMergeMining.UnpackMethod(param, definition.AddBitcoinBlockHeaderMethodName, sendBlock.Data); err != nil {
			return constants.ErrUnpackError
		}
		hash := param.BlockHash()
		if err := param.Hash.SetBytes(hash.Bytes()); err != nil {
			// todo refactor?
			return constants.ErrForbiddenParam
		}
		if blockHeader, err := rC.ZnnRpc().GetBlockHeader(hash); err != nil {
			if err.Error() == constants.ErrDataNonExistent.Error() {
				rC.logger.Debug(constants.ErrDataNonExistent)
				return nil
			}
			return err
		} else if blockHeader == nil {
			rC.logger.Info("block header non existent")
			return nil
		} else {
			if err = rC.AddBlockHeader(blockHeader); err != nil {
				return err
			}
		}
	}

	return rC.eventsStore().SetLastUpdateHeight(receiveBlockHeight)
}

// Subscriptions

func (rC *ZnnNetwork) ListenForMomentumHeight() {
	rC.logger.Debug("func (rC *znnNetwork) ListenForMomentumHeight() {")
	momSub, momChan, err := rC.ZnnRpc().SubscribeToMomentums()
	if err != nil {
		rC.logger.Error(err)
		rC.stopChan <- syscall.SIGINT
		return
	}
	rC.logger.Debug("Successfully started to listen for momentums")
	for {
		select {
		case errSub := <-momSub.Err():
			if errSub != nil {
				rC.logger.Debugf("listen mom err: %s", errSub.Error())
				rC.stopChan <- syscall.SIGINT
				return
			}
		case momentums := <-momChan:
			for _, mom := range momentums {
				if frMom, errState := rC.state.GetFrontierMomentum(); errState != nil {
					rC.logger.Info("error when trying to get frontierMomentum from state")
					rC.logger.Error(errState)
				} else {
					if mom.Height > frMom {
						if errState = rC.state.SetFrontierMomentum(mom.Height); errState != nil {
							rC.logger.Error(errState)
							rC.logger.Info("error when trying to set frontier momentum")
						}
					}
				}
			}
		}
	}
}

func (rC *ZnnNetwork) GetProducerKeyPair() *wallet.KeyPair {
	return rC.producerKeyPair
}

/*
func (rC *ZnnNetwork) ListenForEmbeddedBridgeAccountBlocks() {
	rC.logger.Debug("ListenForEmbeddedBridgeAccountBlocks")
	accBlSub, accBlCh, err := rC.ZnnRpc().SubscribeToAccountBlocks(types.BridgeContract)
	if err != nil {
		rC.logger.Info("sub accBerr: ", err)
		rC.stopChan <- syscall.SIGINT
		return
	}
	rC.logger.Debug("Successfully started to listen for account blocks")
	for {
		select {
		case errSub := <-accBlSub.Err():
			if errSub != nil {
				rC.logger.Debugf("listen accB err: %s", errSub.Error())
				rC.stopChan <- syscall.SIGINT
				return
			}
		case accBlocks := <-accBlCh:
			// these accountBlocks are seen before being inserted in a momentum
			for _, accBlock := range accBlocks {
				if accBlock.BlockType != nom.BlockTypeContractReceive {
					continue
				}
				// we wait for the acc block to be inserted in a momentum
				for {
					time.Sleep(4 * time.Second)
					if receiveBlock, err := rC.ZnnRpc().GetAccountBlockByHash(accBlock.Hash); err != nil {
						rC.logger.Info("receive block non existent")
						continue
					} else if receiveBlock == nil {
						rC.logger.Info("receive block non existent")
						continue
					}
					break
				}

				rC.logger.Info("detected block type receive")
				if sendBlock, err := rC.ZnnRpc().GetAccountBlockByHash(accBlock.FromHash); err != nil {
					rC.logger.Error(err)
				} else if sendBlock == nil {
					rC.logger.Info("send block non existent")
				} else {
					rC.logger.Info("found send block")
					rC.logger.Infof("confirmationDetail is nil: %v", sendBlock.ConfirmationDetail == nil)
					if newErr := rC.InterpretSendBlockData(sendBlock, true, accBlock.Height); newErr != nil {
						rC.logger.Info(newErr)
						// Try one more time
						time.Sleep(3 * time.Second)
						if newErr = rC.InterpretSendBlockData(sendBlock, true, accBlock.Height); newErr != nil {
							rC.logger.Info(newErr)
						}
					}
				}
			}
		}
	}
}
*/

func (rC *ZnnNetwork) ZnnStore() db.ZnnStorage {
	return rC.dbManager.ZnnStorage()
}

func (rC *ZnnNetwork) GetFrontierMomentum() (*api.Momentum, error) {
	return rC.ZnnRpc().GetFrontierMomentum()
}

func (rC *ZnnNetwork) IsSynced() bool {
	syncInfo, err := rC.ZnnRpc().GetSyncInfo()
	if err != nil {
		common.GlobalLogger.Error(err)
		return false
	}
	return syncInfo.State == protocol.SyncDone
}

func (rC *ZnnNetwork) AddBlockHeader(blockHeader *definition.BlockHeaderVariable) error {
	header := block_header.BlockHeader{
		Version:    blockHeader.Version,
		PrevBlock:  blockHeader.PrevBlock,
		MerkleRoot: blockHeader.MerkleRoot,
		Timestamp:  blockHeader.Timestamp,
		Bits:       blockHeader.Bits,
		Nonce:      blockHeader.Nonce,
		Height:     int32(blockHeader.Height),
		Hash:       blockHeader.Hash,
	}

	if blockHeader.WorkSum != nil {
		header.WorkSum = big.NewInt(0).Set(blockHeader.WorkSum)
	} else {
		header.WorkSum = big.NewInt(0)
	}

	return rC.ZnnStore().AddBlockHeader(header)
}
