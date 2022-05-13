package grasure

import (
	"io"
	"net/http"
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

	//the nodes that belongs to the pool
	nodes []*Node

	//the capacity
	volume *Volume

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
//When set dataNodes as nil by default, the pool uses the first nodeNum nodes,
//you can specify the nodes in the order of their ids (indexed from 0).
func NewErasurePool(redun Redundancy, dataShards, parityShards, nodeNum int, blockSize int64, dataNodes []int, layout Pattern) *ErasurePool {
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
		volume:       &Volume{0, 0, 0},
		options:      defaultErasurePoolOption,
	}
	if dataNodes == nil {
		//we adopt the first nodeNum of nodes in cluster
		erasurePool.nodes = c.dataNodes[:nodeNum]
		for i := 0; i < nodeNum; i++ {
			erasurePool.volume.Total += c.dataNodes[i].volume.Total
			erasurePool.volume.Free += c.dataNodes[i].volume.Free
			erasurePool.volume.Used += c.dataNodes[i].volume.Used
		}
	} else {
		for _, id := range dataNodes {
			erasurePool.nodes = append(erasurePool.nodes, c.dataNodes[id])
			erasurePool.volume.Total += c.dataNodes[id].volume.Total
			erasurePool.volume.Free += c.dataNodes[id].volume.Free
			erasurePool.volume.Used += c.dataNodes[id].volume.Used
		}
	}
	stripeNum := int(erasurePool.volume.Total / uint64(dataShards+parityShards) / uint64(blockSize))
	erasurePool.layout = NewLayout(layout, stripeNum, dataShards, parityShards, nodeNum)
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

//Write writes a file or byte flow into certain node.
func (ep *ErasurePool) Write(src io.ReadCloser, addr string) error {
	// xlog.Infoln(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("WRITE", addr, src)
	req.Header.Set("X-Grasure-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		xlog.Fatalln("rpc server: file transfer err:", err)
		return err
	}
	return nil

}
