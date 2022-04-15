package grasure

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/DurantVivado/reedsolomon"
	"golang.org/x/sync/errgroup"
)

const (
	defaultInfoFilePath = "cluster.info"
	defaultNodeFilePath = "nodes.addr"
)
type Cluster struct{
	//the global unique id of a cluster
	uuid string

	//the nodes' info of a cluster, i.e. id->*Node
	nodes map[uint32]*Node

	//capacity related parameters
	Used    uint64
	Free    uint64
	Total   uint64

	// the number of data blocks in a stripe
	K int `json:"dataShards"`

	// the number of parity blocks in a stripe
	M int `json:"parityShards"`

	// the block size. default to 4KiB
	BlockSize int64 `json:"blockSize"`

	//FileMeta lists, indicating fileName, fileSize, fileHash, fileDist...
	FileMeta []*fileInfo `json:"fileLists"`

	//how many stripes are allowed to encode/decode concurrently
	ConStripes int `json:"-"`

	// the replication factor for config file
	ReplicateFactor int

	// the reedsolomon streaming encoder, for streaming access
	sEnc reedsolomon.StreamEncoder

	// the reedsolomon encoder, for block access
	enc reedsolomon.Encoder

	// the data stripe size, equal to k*bs
	dataStripeSize int64

	// the data plus parity stripe size, equal to (k+m)*bs
	allStripeSize int64

	// infoFile storage basic information of the cluster
	InfoFilePath string `json:"-"`

	//the path of file includes all nodes address, provided by user
	NodeFilePath string `json:"-"`

	//file map
	fileMap sync.Map


	// whether to override former files or directories, default to false
	Override bool `json:"-"`

	// errgroup pool
	errgroupPool sync.Pool

	// mutex
	mu sync.RWMutex

	//whether to mute outputs
	Quiet bool `json:"-"`
}


//NewCluster initializes a Cluster with customized ataShards, parityShards, usedNodeNum, replicateFactor and blockSize
func NewCluster(dataShards, parityShards, usedNodeNum, replicateFactor int, blockSize int64) (c *Cluster, err error) {
	
	//if !e.Quiet {
	//	fmt.Println("Warning: you are initializing a new erasure-coded system, which means the previous data will also be reset.")
	//}
	//if !assume {
	//	if ans, err := consultUserBeforeAction(); !ans && err == nil {
	//		return nil
	//	} else if err != nil {
	//		return err
	//	}
	//}
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
	//if e.DiskNum > len(e.diskInfos) {
	//	return errDiskNumTooLarge
	//}
	//if e.ConStripes < 1 {
	//	e.ConStripes = 1 //totally serialized
	//}
	//replicate the config files

	if replicateFactor < 1 {
		return nil, errInvalidReplicateFactor
	}
	//err = e.resetSystem()
	//if err != nil {
	//	return err
	//}
	//if !e.Quiet {
	//	fmt.Printf("System init!\n Erasure parameters: dataShards:%d, parityShards:%d,blocksize:%d,diskNum:%d\n",
	//		e.K, e.M, e.BlockSize, e.DiskNum)
	//}
	c = &Cluster{
		uuid:            genUUID(0),
		Used:            0,
		Free:            0,
		Total:           0,
		K:               dataShards,
		M:               parityShards,
		BlockSize:       blockSize,
		ConStripes:      100,
		ReplicateFactor: replicateFactor,
		dataStripeSize:  int64(dataShards) * blockSize,
		allStripeSize:   int64(dataShards+parityShards) * blockSize,
		InfoFilePath:    defaultInfoFilePath,
		NodeFilePath: 	 defaultNodeFilePath,
		Override:        false,
		errgroupPool:    sync.Pool{},
		mu:              sync.RWMutex{},
		Quiet:           false,
	}
	return c,nil
}

//read nodes
func (c* Cluster) ReadNodes() error{
	c.mu.RLock()
	defer c.mu.RUnlock()
	f, err := os.Open(c.NodeFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	c.nodes = make(map[uint32]*Node, 0)
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		path := string(line)
		if ok, err := pathExist(path); !ok && err == nil {
			return &diskError{path, "disk path not exist"}
		} else if err != nil {
			return err
		}
		metaPath := filepath.Join(path, "META")
		flag := false
		if ok, err := pathExist(metaPath); ok && err == nil {
			flag = true
		} else if err != nil {
			return err
		}
		diskInfo := &diskInfo{diskPath: string(line), available: true, ifMetaExist: flag}
		e.diskInfos = append(e.diskInfos, diskInfo)
	}
	return nil
}

//reset the storage
func (c *Cluster) reset() error {

	g := new(errgroup.Group)

	for _, path := range e.diskInfos[:e.DiskNum] {
		path := path
		files, err := os.ReadDir(path.diskPath)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			continue
		}
		g.Go(func() error {
			for _, file := range files {
				err = os.RemoveAll(filepath.Join(path.diskPath, file.Name()))
				if err != nil {
					return err
				}

			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

//reset the system including config and data
func (c *Cluster) resetSystem() error {

	//in-memory meta reset
	e.FileMeta = make([]*fileInfo, 0)
	// for k := range e.fileMap {
	// 	delete(e.fileMap, k)
	// }
	e.fileMap.Range(func(key, value interface{}) bool {
		e.fileMap.Delete(key)
		return true
	})
	err = e.WriteConfig()
	if err != nil {
		return err
	}
	//delete the data blocks under all diskPath
	err = e.reset()
	if err != nil {
		return err
	}
	err = e.replicateConfig(e.ReplicateFactor)
	if err != nil {
		return err
	}
	return nil
}

//ReadConfig reads the config file during system warm-up.
//
//Calling it before actions like encode and read is a good habit.
func (c *Cluster) ReadConfig() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ex, err := pathExist(e.ConfigFile); !ex && err == nil {
		// we try to recover the config file from the storage system
		// which renders the last chance to heal
		err = e.rebuildConfig()
		if err != nil {
			return errConfFileNotExist
		}
	} else if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(e.ConfigFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &e)
	if err != nil {
		//if json file is broken, we try to recover it

		err = e.rebuildConfig()
		if err != nil {
			return errConfFileNotExist
		}

		data, err := ioutil.ReadFile(e.ConfigFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &e)
		if err != nil {
			return err
		}
	}
	//initialize the ReedSolomon Code
	e.enc, err = reedsolomon.New(e.K, e.M,
		reedsolomon.WithAutoGoroutines(int(e.BlockSize)),
		reedsolomon.WithCauchyMatrix(),
		reedsolomon.WithInversionCache(true),
	)
	if err != nil {
		return err
	}
	e.dataStripeSize = int64(e.K) * e.BlockSize
	e.allStripeSize = int64(e.K+e.M) * e.BlockSize

	e.errgroupPool.New = func() interface{} {
		return &errgroup.Group{}
	}
	//unzip the fileMap
	for _, f := range e.FileMeta {
		stripeNum := len(f.Distribution)
		f.blockToOffset = makeArr2DInt(stripeNum, e.K+e.M)
		f.blockInfos = make([][]*blockInfo, stripeNum)
		countSum := make([]int, e.DiskNum)
		for row := range f.Distribution {
			f.blockInfos[row] = make([]*blockInfo, e.K+e.M)
			for line := range f.Distribution[row] {
				diskId := f.Distribution[row][line]
				f.blockToOffset[row][line] = countSum[diskId]
				f.blockInfos[row][line] = &blockInfo{bstat: blkOK}
				countSum[diskId]++
			}
		}
		//update the numBlocks
		for i := range countSum {
			e.diskInfos[i].numBlocks += countSum[i]
		}
		e.fileMap.Store(f.FileName, f)
		// e.fileMap[f.FileName] = f

	}
	e.FileMeta = make([]*fileInfo, 0)
	// we
	//e.sEnc, err = reedsolomon.NewStreamC(e.K, e.M, conReads, conWrites)
	// if err != nil {
	// 	return err
	// }

	return nil
}

//Replicate the config file into the system for k-fold
//it's NOT striped and encoded as a whole piece.
func (c *Cluster) replicateConfig(k int) error {
	selectDisk := genRandomArr(e.DiskNum, 0)[:k]
	for _, i := range selectDisk {
		disk := e.diskInfos[i]
		disk.ifMetaExist = true
		replicaPath := filepath.Join(disk.diskPath, "META")
		_, err = copyFile(e.ConfigFile, replicaPath)
		if err != nil {
			log.Println(err.Error())
		}

	}
	return nil
}

//WriteConfig writes the erasure parameters and file information list into config files.
//
//Calling it after actions like encode and read is a good habit.
func (c *Cluster) WriteConfig() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	f, err := os.OpenFile(e.ConfigFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// we marsh filemap into fileLists
	// for _, v := range e.fileMap {
	// 	e.FileMeta = append(e.FileMeta, v)
	// }
	e.fileMap.Range(func(k, v interface{}) bool {
		e.FileMeta = append(e.FileMeta, v.(*fileInfo))
		return true
	})
	data, err := json.Marshal(e)
	// data, err := json.MarshalIndent(e, " ", "  ")
	if err != nil {
		return err
	}
	buf := bufio.NewWriter(f)
	_, err = buf.Write(data)
	if err != nil {
		return err
	}
	buf.Flush()
	// f.Sync()
	err = e.updateConfigReplica()
	if err != nil {
		return err
	}
	return nil
}

//reconstruct the config file if possible
func (c *Cluster) rebuildConfig() error {
	//we read file meta in the disk path and try to rebuild the config file
	for i := range e.diskInfos[:e.DiskNum] {
		disk := e.diskInfos[i]
		replicaPath := filepath.Join(disk.diskPath, "META")
		if ok, err := pathExist(replicaPath); !ok && err == nil {
			continue
		}
		_, err = copyFile(replicaPath, e.ConfigFile)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

//update the config file of all replica
func (c *Cluster) updateConfigReplica() error {

	//we read file meta in the disk path and try to rebuild the config file
	if e.ReplicateFactor < 1 {
		return nil
	}
	for i := range e.diskInfos[:e.DiskNum] {
		disk := e.diskInfos[i]
		replicaPath := filepath.Join(disk.diskPath, "META")
		if ok, err := pathExist(replicaPath); !ok && err == nil {
			continue
		}
		_, err = copyFile(e.ConfigFile, replicaPath)
		if err != nil {
			return err
		}
	}
	return nil
}

