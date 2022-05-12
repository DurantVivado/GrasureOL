package grasure

import (
	"sync"

	"github.com/DurantVivado/GrasureOL/encoder"
	"github.com/DurantVivado/GrasureOL/xlog"
	"github.com/DurantVivado/reedsolomon"
)

type erasurePoolOption struct {
	// whether to override former files or directories, default to false
	override bool
}

var defaultErasurePoolOption = &erasurePoolOption{
	override: false,
}

type ErasurePool struct {
	//the redun is how the files are encoded in this pool
	Redun Redundancy `json:"redundancy"`

	// the number of data blocks in a stripe
	K int `json:"dataShards"`

	// the number of parity blocks in a stripe
	M int `json:"parityShards"`

	//the used node number for the pool
	NodeNum int `json:"nodeNum"`

	// the block size. default to 4KiB
	BlockSize int64 `json:"blockSize"`

	//the nodes that belongs to the pool, by ID
	nodes []*Node

	//the encoder handles encoding and decoding of files
	enc encoder.Encoder

	//file map, it stores on meta server
	layout *Layout

	// errgroup pool
	errgroupPool sync.Pool

	// mutex
	mu sync.RWMutex

	options *erasurePoolOption
}

//NewErasurePool news an erasurePool with designated dataShards, parityShards, nodeNum and blockSize,
//When set nodeList as nil by default, the pool uses the first nodeNum nodes,
//you can specify the nodes in the order of their ids (indexed from 0).
func NewErasurePool(redun Redundancy, dataShards, parityShards, nodeNum int, blockSize int64, nodeList []int) *ErasurePool {
	if dataShards <= 0 || parityShards <= 0 {
		xlog.Fatal(errInvalidShardNumber)
	}
	//The reedsolomon library only implements GF(2^8) and will be improved later
	if dataShards+parityShards > 256 {
		xlog.Fatal(errMaxShardNum)
	}
	if dataShards+parityShards > nodeNum {
		xlog.Fatal(errNotEnoughNodeAvailable)
	}

	erasurePool := &ErasurePool{
		Redun:        redun,
		K:            dataShards,
		M:            parityShards,
		BlockSize:    blockSize,
		errgroupPool: sync.Pool{},
		mu:           sync.RWMutex{},
		NodeNum:      nodeNum,
		options:      defaultErasurePoolOption,
	}
	if nodeList == nil {
		//default
		erasurePool.nodes = c.nodeList[:nodeNum]
	} else {
		for _, id := range nodeList {
			erasurePool.nodes = append(erasurePool.nodes, c.nodeList[id])
		}
	}
	switch redun {
	case Erasure_RS:
		enc, err := reedsolomon.New(dataShards, parityShards,
			reedsolomon.WithAutoGoroutines(int(blockSize)),
			reedsolomon.WithCauchyMatrix(),
			reedsolomon.WithInversionCache(true))
		if err != nil {
			xlog.Fatal(err)
		}
		erasurePool.enc = enc
	default:
		xlog.Fatal(errNoRedunSelected)
	}
	return erasurePool
}
