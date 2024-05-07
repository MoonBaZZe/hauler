package btc_storage

import (
	"bytes"
	"errors"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/syndtr/goleveldb/leveldb"
	zcommon "github.com/zenon-network/go-zenon/common"
)

func getTransactionPrefix() []byte {
	return zcommon.JoinBytes(transactionPrefix)
}

func getTransactionKey(id chainhash.Hash) []byte {
	return zcommon.JoinBytes(getTransactionPrefix(), id.CloneBytes())
}

func (bh *btcTransactionStore) AddTransaction(msgTx *wire.MsgTx) error {
	txBuf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
	if err := msgTx.Serialize(txBuf); err != nil {
		bh.SendSigInt()
		return err
	} else {
		if err := bh.DB.Put(getTransactionKey(msgTx.TxHash()), txBuf.Bytes()); err != nil {
			bh.SendSigInt()
			return err
		}
	}
	return nil
}

func (bh *btcTransactionStore) GetTransaction(hash chainhash.Hash) (*wire.MsgTx, error) {
	data, err := bh.DB.Get(getTransactionKey(hash))
	if errors.Is(err, leveldb.ErrNotFound) {
		return nil, err
	}
	if err != nil {
		bh.SendSigInt()
		return nil, err
	}

	// Deserialize the transaction.
	var msgTx *wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return msgTx, nil
}
