package db

import (
	zdb "github.com/zenon-network/go-zenon/common/db"
)

type ZnnStorage interface {
	Storage() zdb.DB
	Snapshot() ZnnStorage
	SendSigInt()

	GetLastUpdateHeight() (uint64, error)
	SetLastUpdateHeight(uint64) error
}
