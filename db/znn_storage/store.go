package znn_storage

import (
	"github.com/MoonBaZZe/hauler/db"
	zdb "github.com/zenon-network/go-zenon/common/db"
	"os"
	"syscall"
)

func getStorageIterator() []byte {
	return blockHeaderPrefix
}

// blockHeadersStore address is the contract address for the events
type blockHeadersStore struct {
	zdb.DB
	stopChan chan os.Signal
}

func (bh *blockHeadersStore) Storage() zdb.DB {
	return zdb.DisableNotFound(bh.DB.Subset(getStorageIterator()))
}
func (bh *blockHeadersStore) Snapshot() db.ZnnStorage {
	return NewZnnStorage(bh.DB.Snapshot(), bh.stopChan)
}

func (bh *blockHeadersStore) SendSigInt() {
	bh.stopChan <- syscall.SIGINT
}

func NewZnnStorage(db zdb.DB, stopChan chan os.Signal) db.ZnnStorage {
	if db == nil {
		panic("account store can't operate with nil db")
	}
	return &blockHeadersStore{
		DB:       db,
		stopChan: stopChan,
	}
}
