package db

import (
	"github.com/MoonBaZZe/hauler/common/block_header"
	zdb "github.com/zenon-network/go-zenon/common/db"
	"github.com/zenon-network/go-zenon/common/types"
)

type ZnnStorage interface {
	Storage() zdb.DB
	Snapshot() ZnnStorage
	SendSigInt()

	AddBlockHeader(block_header.BlockHeader) error
	GetBlockHeader(types.Hash) (*block_header.BlockHeader, error)

	GetLastUpdateHeight() (uint64, error)
	SetLastUpdateHeight(uint64) error
}
