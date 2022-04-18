package consistentHash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

//Hash maps bytes to uint32
type Hash func(key []byte) uint32

//ConsistentHash contains all hashed keys
type ConsistentHash struct {
	hash     Hash
	replicas int
	keys     []int          //sorted
	hashMap  map[int]string //store nodes
}

//NewConsistentHash creates a Map instance
// hash func defaults to crc32
func NewConsistentHash(replicas int, fn Hash) *ConsistentHash {
	m := &ConsistentHash{hash: fn, replicas: replicas, hashMap: make(map[int]string)}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

//AddNode adds a node into hash ring
func (ch *ConsistentHash) AddNode(keys ...string) {
	for _, key := range keys {
		for i := 0; i < ch.replicas; i++ {
			hashVal := int(ch.hash([]byte(strconv.Itoa(i) + key)))
			ch.keys = append(ch.keys, hashVal)
			ch.hashMap[hashVal] = key
		}
	}
	sort.Ints(ch.keys)
}

//GetNode gets a node according to the key
func (ch *ConsistentHash) GetNode(key string) string {
	if len(ch.keys) == 0 {
		return ""
	}
	hashVal := int(ch.hash([]byte(key)))
	//Binary Search for suitable node
	i := sort.Search(len(ch.keys), func(i int) bool {
		return ch.keys[i] >= hashVal
	})
	return ch.hashMap[ch.keys[i%len(ch.keys)]]
}

func (ch *ConsistentHash) DelNode(keys ...string) {
	for _, key := range keys {
		for i := 0; i < ch.replicas; i++ {
			hashVal := int(ch.hash([]byte(strconv.Itoa(i) + key)))
			id := sort.Search(len(ch.keys), func(i int) bool {
				return ch.keys[i] >= hashVal
			})
			ch.keys = append(ch.keys[:id], ch.keys[id+1:]...)
			delete(ch.hashMap, hashVal)
		}
	}
}
