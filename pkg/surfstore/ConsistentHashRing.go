package surfstore

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

type ConsistentHashRing struct {
	ServerMap  map[string]string // 一个从哈希后的服务器地址到原始服务器地址的映射
	SortedKeys []string          // 一个已排序的哈希服务器地址列表
}

func (c ConsistentHashRing) GetResponsibleServer(blockId string) string {
	//hashedBlockId := c.Hash(blockId)
	// 查找第一个大于等于哈希块 ID 的服务器地址的索引
	i := sort.Search(len(c.SortedKeys), func(j int) bool {
		return c.SortedKeys[j] >= blockId
	})
	// 如果没有找到这样的服务器地址，则选择第一个服务器
	if i == len(c.SortedKeys) {
		i = 0
	}
	// 返回负责处理该块 ID 的服务器地址
	return c.ServerMap[c.SortedKeys[i]]
}

func (c ConsistentHashRing) Hash(addr string) string {
	h := sha256.New()
	h.Write([]byte(addr))
	return hex.EncodeToString(h.Sum(nil))
}

func NewConsistentHashRing(serverAddrs []string) *ConsistentHashRing {
	c := &ConsistentHashRing{
		ServerMap: make(map[string]string),
	}
	// 遍历服务器地址，并将它们哈希后的值添加到哈希环中
	for _, addr := range serverAddrs {
		hashedAddr := c.Hash("blockstore"+addr)
		c.ServerMap[hashedAddr] = addr
		c.SortedKeys = append(c.SortedKeys, hashedAddr)
	}
	// 将哈希后的服务器地址按照字典序排序
	sort.Strings(c.SortedKeys)
	return c
}