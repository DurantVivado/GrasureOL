package grasure

import "sync"

type EncodeType string

const (
	ReedSolomon EncodeType = "ReedSolomon"
	XOR EncodeType = "XOR"
	LRC EncodeType = "LRC"
	Replica EncodeType = "Replica"
)

type erasurePoolOption struct {

}

var defaultErasurePoolOption = &erasurePoolOption{}

type ErasurePool struct{
	//the encodeType is how the files are encoded in this pool
	encodeType EncodeType

	// the number of data blocks in a stripe
	K int `json:"dataShards"`

	// the number of parity blocks in a stripe
	M int `json:"parityShards"`

	// the block size. default to 4KiB
	BlockSize int64 `json:"blockSize"`

	//the nodes that belongs to the pool, by ID
	nodeIds []int

	//file map
	fileMap sync.Map

	// whether to override former files or directories, default to false
	Override bool `json:"-"`

	// errgroup pool
	errgroupPool sync.Pool

	// mutex
	mu sync.RWMutex

	options *erasurePoolOption
}

func NewErasurePool(encodeType string, dataShards, parityShards, usedNodeNum, replicateFactor int, blockSize int64) (*ErasurePool, error){
	if dataShards <= 0 || parityShards <= 0 {
		return nil, errInvalidShardNumber
	}
	//The reedsolomon library only implements GF(2^8) and will be improved later
	if dataShards + parityShards > 256 {
		return nil, errMaxShardNum
	}
	if dataShards + parityShards > usedNodeNum {
		return nil, errNotEnoughNodeAvailable
	}
	if replicateFactor < 1 {
		return nil, errInvalidReplicateFactor
	}

	erasurePool := &ErasurePool{
		encodeType:   EncodeType(encodeType),
		K:            dataShards,
		M:            parityShards,
		BlockSize:    blockSize,
		fileMap:      sync.Map{},
		errgroupPool: sync.Pool{},
		mu:           sync.RWMutex{},
	}
	return erasurePool, nil
}