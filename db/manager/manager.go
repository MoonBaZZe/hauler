package manager

import (
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/db"
	"github.com/MoonBaZZe/hauler/db/btc_storage"

	"github.com/MoonBaZZe/hauler/db/znn_storage"
	zdb "github.com/zenon-network/go-zenon/common/db"
	"go.uber.org/zap"
	"os"
)

type Manager struct {
	znnStorage db.ZnnStorage
	btcStorage db.BtcStorage
	stopChan   chan os.Signal
	logger     *zap.SugaredLogger
}

func NewDbManager(stop chan os.Signal) (*Manager, error) {
	newZnnLdb, err := common.CreateOrOpenLevelDb(common.HeadChainName)
	if err != nil {
		return nil, err
	}
	newLogger, errLog := common.CreateSugarLogger()
	if errLog != nil {
		return nil, errLog
	}

	newDbManager := &Manager{
		znnStorage: znn_storage.NewZnnStorage(zdb.NewLevelDBWrapper(newZnnLdb), stop),
		btcStorage: btc_storage.NewBtcStorage(zdb.NewLevelDBWrapper(newZnnLdb), stop),
		stopChan:   stop,
		logger:     newLogger,
	}
	return newDbManager, nil
}

func (m *Manager) ZnnStorage() db.ZnnStorage {
	return m.znnStorage
}

func (m *Manager) BtcStorage() db.BtcStorage {
	return m.btcStorage
}
