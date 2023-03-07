package surfstore

import (
	context "context"
	sync "sync"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type MetaStore struct {
	FileMetaMap    map[string]*FileMetaData
	mtx            sync.Mutex
	BlockStoreAddrs    []string
	ConsistentHashRing *ConsistentHashRing
	UnimplementedMetaStoreServer
}


func (m *MetaStore) GetFileInfoMap(ctx context.Context, _ *emptypb.Empty,) (*FileInfoMap, error) {
	return &FileInfoMap{FileInfoMap: m.FileMetaMap}, nil
}

func (m *MetaStore) UpdateFile(ctx context.Context, fileMetaData *FileMetaData) (*Version, error) {
	fileName := fileMetaData.Filename
	version := fileMetaData.Version
	m.mtx.Lock()
	_, exist := m.FileMetaMap[fileName]
	if exist {
		if version == m.FileMetaMap[fileName].Version+1 {
			m.FileMetaMap[fileName] = fileMetaData
		} else {
			//if therer is a version problem, use v=-1 to tell the user
			//that update fails
			version = -1
		}

	} else {
		//if data doesn't exist at all, then create it
		m.FileMetaMap[fileName] = fileMetaData
	}
	m.mtx.Unlock()
	return &Version{Version: version}, nil
}




func (m *MetaStore) GetBlockStoreMap(ctx context.Context, blockHashesIn *BlockHashes) (*BlockStoreMap, error) {
	res := make(map[string]*BlockHashes)
	for _,hash := range blockHashesIn.Hashes{
		resServer := m.ConsistentHashRing.GetResponsibleServer(hash)
		_,ok := res[resServer]
		if ok{
			res[resServer].Hashes = append(res[resServer].Hashes,hash)
		}else{
			res[resServer] = &BlockHashes{}
			res[resServer].Hashes = append(res[resServer].Hashes,hash)
		}
	}
	return &BlockStoreMap{BlockStoreMap:res},nil
}

func (m *MetaStore) GetBlockStoreAddrs(ctx context.Context, _ *emptypb.Empty) (*BlockStoreAddrs, error) {
	return &BlockStoreAddrs{BlockStoreAddrs:m.BlockStoreAddrs},nil
}


// This line guarantees all method for MetaStore are implemented
var _ MetaStoreInterface = new(MetaStore)


func NewMetaStore(blockStoreAddrs []string) *MetaStore {
	return &MetaStore{
		FileMetaMap:    map[string]*FileMetaData{},
		BlockStoreAddrs:    blockStoreAddrs,
		ConsistentHashRing: NewConsistentHashRing(blockStoreAddrs),
	}
}