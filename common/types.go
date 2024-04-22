package common

import (
	"github.com/MoonBaZZe/hauler/common/block_header"
	"github.com/btcsuite/btcd/wire"
	"github.com/zenon-network/go-zenon/common/types"
	"math/big"
)

func WireBlockHeaderToHaulerBlockHeader(header *wire.BlockHeader) *block_header.BlockHeader {
	return &block_header.BlockHeader{
		Version:    header.Version,
		PrevBlock:  types.HexToHashPanic(header.PrevBlock.String()),
		MerkleRoot: types.HexToHashPanic(header.MerkleRoot.String()),
		Timestamp:  uint32(header.Timestamp.Unix()),
		Bits:       header.Bits,
		Nonce:      header.Nonce,
		Height:     0,
		WorkSum:    big.NewInt(0),
		Hash:       types.HexToHashPanic(header.BlockHash().String()),
	}
}
