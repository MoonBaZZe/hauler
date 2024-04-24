package btc_storage

import (
	"github.com/MoonBaZZe/hauler/db"
	zdb "github.com/zenon-network/go-zenon/common/db"
	"os"
	"syscall"
)

func getStorageIterator() []byte {
	return transactionPrefix
}

type btcTransactionStore struct {
	zdb.DB
	stopChan chan os.Signal
}

func (bh *btcTransactionStore) Storage() zdb.DB {
	return zdb.DisableNotFound(bh.DB.Subset(getStorageIterator()))
}
func (bh *btcTransactionStore) Snapshot() db.BtcStorage {
	return NewBtcStorage(bh.DB.Snapshot(), bh.stopChan)
}

func (bh *btcTransactionStore) SendSigInt() {
	bh.stopChan <- syscall.SIGINT
}

func NewBtcStorage(db zdb.DB, stopChan chan os.Signal) db.BtcStorage {
	if db == nil {
		panic("account store can't operate with nil db")
	}
	return &btcTransactionStore{
		DB:       db,
		stopChan: stopChan,
	}
}
