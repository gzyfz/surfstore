package surfstore

import (
	context "context"
	"sync"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type BlockStore struct {
	BlockMap map[string]*Block
	mtx      sync.Mutex
	UnimplementedBlockStoreServer
}

func (bs *BlockStore) GetBlock(ctx context.Context, blockHash *BlockHash) (*Block, error) {
	bs.mtx.Lock()
	//protect the content of unchangable when fetch it
	block := bs.BlockMap[blockHash.Hash]
	bs.mtx.Unlock()
	//now set the content back to changable
	return block, nil
}

func (bs *BlockStore) PutBlock(ctx context.Context, block *Block) (*Success, error) {
	hash := GetBlockHashString(block.BlockData)
	bs.mtx.Lock()
	//using map to store the blocks
	bs.BlockMap[hash] = block
	bs.mtx.Unlock()
	//could it fail somehow?
	return &Success{Flag: true}, nil
}

// Given a list of hashes “in”, returns a list containing the
// subset of in that are stored in the key-value store
func (bs *BlockStore) HasBlocks(ctx context.Context, blockHashesIn *BlockHashes) (*BlockHashes, error) {
	var res []string
	for _, hash := range blockHashesIn.Hashes {
		bs.mtx.Lock()
		_, exist := bs.BlockMap[hash]
		if exist {
			res = append(res, hash)
		}
		bs.mtx.Unlock()
	}
	return &BlockHashes{Hashes: res}, nil
}


// Return a list containing all blockHashes on this block server
func (bs *BlockStore) GetBlockHashes(ctx context.Context, _ *emptypb.Empty) (*BlockHashes, error) {
		bs.mtx.Lock()
		j := 0
		keys := make([]string,len(bs.BlockMap))
		for k := range(bs.BlockMap){
			keys[j] = k
			j++
		}
		bs.mtx.Unlock()
		return &BlockHashes{Hashes: keys},nil
	}

// This line guarantees all method for BlockStore are implemented
var _ BlockStoreInterface = new(BlockStore)

func NewBlockStore() *BlockStore {
	return &BlockStore{
		BlockMap: map[string]*Block{},
	}
}
