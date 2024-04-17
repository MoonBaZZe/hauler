package block_header

import (
	zcommon "github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"google.golang.org/protobuf/proto"
	"math/big"
)

type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int32 `json:"version"`

	// Hash of the previous block header in the block chain.
	PrevBlock types.Hash `json:"prevBlock"`

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot types.Hash `json:"merkleRoot"`

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp uint32 `json:"timestamp"`

	// Difficulty target for the block.
	Bits uint32 `json:"bits"`

	// Nonce used to generate the block.
	Nonce uint32 `json:"nonce"`

	Height  uint32     `json:"height"`
	WorkSum *big.Int   `json:"workSum"`
	Hash    types.Hash `json:"hash"`
}

func (bh *BlockHeader) Proto() *BlockHeaderProto {
	return &BlockHeaderProto{
		Version:    bh.Version,
		PrevBlock:  bh.PrevBlock.Bytes(),
		MerkleRoot: bh.PrevBlock.Bytes(),
		Timestamp:  bh.Timestamp,
		Bits:       bh.Bits,
		Nonce:      bh.Nonce,
		Height:     bh.Height,
		WorkSum:    bh.WorkSum.Bytes(),
		Hash:       bh.Hash.Bytes(),
	}
}

func DeProtoBlockHeader(b *BlockHeaderProto) *BlockHeader {
	ev := &BlockHeader{
		Version:    b.Version,
		PrevBlock:  types.BytesToHashPanic(b.PrevBlock),
		MerkleRoot: types.BytesToHashPanic(b.MerkleRoot),
		Timestamp:  b.Timestamp,
		Bits:       b.Bits,
		Nonce:      b.Nonce,
		Height:     b.Height,
		WorkSum:    zcommon.BytesToBigInt(b.WorkSum),
		Hash:       types.BytesToHashPanic(b.Hash),
	}
	return ev
}

func (bh *BlockHeader) Serialize() ([]byte, error) {
	return proto.Marshal(bh.Proto())
}
func DeserializeBlockHeader(data []byte) (*BlockHeader, error) {
	ev := &BlockHeaderProto{}
	if err := proto.Unmarshal(data, ev); err != nil {
		return nil, err
	}
	return DeProtoBlockHeader(ev), nil
}
