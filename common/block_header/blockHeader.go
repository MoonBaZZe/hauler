package block_header

import (
	"github.com/zenon-network/go-zenon/common/types"
	"google.golang.org/protobuf/proto"
)

type BlockHeader struct {
	Hash   types.Hash
	Height int32
}

func (bh *BlockHeader) Proto() *BlockHeaderProto {
	return &BlockHeaderProto{
		Hash:   bh.Hash.Bytes(),
		Height: bh.Height,
	}
}

func DeProtoBlockHeader(b *BlockHeaderProto) *BlockHeader {
	ev := &BlockHeader{
		Hash:   types.BytesToHashPanic(b.Hash),
		Height: b.Height,
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
