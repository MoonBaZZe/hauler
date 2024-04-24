package znn_storage

import (
	"encoding/binary"
	"errors"
	"github.com/MoonBaZZe/hauler/common/block_header"
	"github.com/syndtr/goleveldb/leveldb"
	zcommon "github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

func getBlockHeaderPrefix() []byte {
	return zcommon.JoinBytes(blockHeaderPrefix)
}

func getBlockHeaderKey(id types.Hash) []byte {
	return zcommon.JoinBytes(getBlockHeaderPrefix(), id.Bytes())
}

func (bh *blockHeadersStore) AddBlockHeader(blockHeader block_header.BlockHeader) error {
	if eventBytes, err := blockHeader.Serialize(); err != nil {
		bh.SendSigInt()
		return err
	} else {
		if err := bh.DB.Put(getBlockHeaderKey(blockHeader.Hash), eventBytes); err != nil {
			bh.SendSigInt()
			return err
		}
	}
	return nil
}

func (bh *blockHeadersStore) GetBlockHeader(id types.Hash) (*block_header.BlockHeader, error) {
	data, err := bh.DB.Get(getBlockHeaderKey(id))
	if errors.Is(err, leveldb.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		bh.SendSigInt()
		return nil, err
	}

	event, err := block_header.DeserializeBlockHeader(data)
	if err != nil {
		bh.SendSigInt()
		return nil, err
	}
	return event, nil
}

//func (es *blockHeadersStore) SetWrapRequestStatus(id types.Hash, status uint32) error {
//	if event, err := es.GetWrapRequestById(id); err != nil {
//		es.SendSigInt()
//		return err
//	} else {
//		if event == nil {
//			return leveldb.ErrNotFound
//		}
//		event.RedeemStatus = status
//		if eventBytes, err := event.Serialize(); err != nil {
//			es.SendSigInt()
//			return err
//		} else {
//			if err := es.DB.Put(getBlockHeaderKey(event.Id), eventBytes); err != nil {
//				es.SendSigInt()
//				return err
//			}
//		}
//	}
//	return nil
//}

//func (es *blockHeadersStore) GetUnsignedWrapRequests() ([]*events.WrapRequestZnn, error) {
//	iterator := es.DB.NewIterator(getBlockHeaderPrefix())
//	defer iterator.Release()
//	result := make([]*events.WrapRequestZnn, 0)
//
//	for {
//		if !iterator.Next() {
//			if iterator.Error() != nil {
//				es.SendSigInt()
//				return nil, iterator.Error()
//			}
//			break
//		}
//		if iterator.Value() == nil {
//			continue
//		}
//
//		event, err := events.DeserializeWrapEventZnn(iterator.Value())
//		if err != nil {
//			es.SendSigInt()
//			return nil, err
//		}
//		if len(event.Signature) > 0 {
//			continue
//		}
//
//		result = append(result, event)
//	}
//	return result, nil
//}

func getLastUpdateKey() []byte {
	return zcommon.JoinBytes(lastUpdatePrefix)
}

func (bh *blockHeadersStore) GetLastUpdateHeight() (uint64, error) {
	data, err := bh.DB.Get(getLastUpdateKey())
	if errors.Is(err, leveldb.ErrNotFound) {
		return 1, nil
	}

	if err != nil {
		bh.SendSigInt()
		return 0, err
	}

	return binary.LittleEndian.Uint64(data), nil
}

func (bh *blockHeadersStore) SetLastUpdateHeight(accBlHeight uint64) error {
	if _, err := bh.GetLastUpdateHeight(); err != nil {
		return err
	} else {
		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, accBlHeight)
		if err := bh.DB.Put(getLastUpdateKey(), bytes); err != nil {
			bh.SendSigInt()
			return err
		}
	}
	return nil
}
