package db

import (
	"github.com/MoonBaZZe/hauler/common/block_header"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
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

type BtcStorage interface {
	Storage() zdb.DB
	Snapshot() BtcStorage
	SendSigInt()

	AddTransaction(*wire.MsgTx) error
	GetTransaction(chainhash.Hash) (*wire.MsgTx, error)
}
