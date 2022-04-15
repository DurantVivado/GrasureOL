// Package grasure is an Universal Erasure Coding Architecture in Go
//
// For usage and examples, see https://github.com/DurantVivado/Grasure
//
package grasure

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

type NodeType string

const (
	ClientType NodeType = "Client"
	ServerType NodeType = "Server"
	GateWayType NodeType = "Gateway"
	Storage NodeType = "Storage"
)

type NodeHealth string

const (
	HealthOK NodeType = "Client"
	CPUFailed NodeType = "CPUFailed"
	DiskFailed NodeType = "DiskFailed"
	NetworkError NodeType = "NetworkError"
)

type Node struct {
	uid string
	addr string
	nodeType NodeType
	health NodeHealth

	// the disk number, only the first diskNum disks are used in diskPathFile
	DiskNum int `json:"diskNum"`

	// diskInfo lists
	diskInfos []*diskInfo

	// the path of file recording all disks path
	DiskFilePath string `json:"-"`
}



//diskInfo contains the disk-level information
//but not mechanical and electrical parameters
type diskInfo struct {
	//the disk path
	diskPath string

	//it's flag and when disk fails, it renders false.
	available bool

	//it tells how many blocks a disk holds
	numBlocks int

	//it's a disk with meta file?
	ifMetaExist bool

	//the capacity of a disk
	capacity int64

}

//fileInfo defines the file-level information,
//it's concurrently safe
type fileInfo struct {
	//file name
	FileName string `json:"fileName"`

	//file size
	FileSize int64 `json:"fileSize"`

	//hash value (SHA256 by default)
	Hash string `json:"fileHash"`

	//distribution forms a block->disk mapping
	Distribution [][]int `json:"fileDist"`

	//blockToOffset has the same row and column number as Distribution but points to the block offset relative to a disk.
	blockToOffset [][]int

	//block state, default to blkOK otherwise blkFail in case of bit-rot.
	blockInfos [][]*blockInfo

	//system-level file info
	// metaInfo     *os.fileInfo

	//loadBalancedScheme is the most load-balanced scheme derived by SGA algo
	loadBalancedScheme [][]int
}

type blockStat uint8
type blockInfo struct {
	bstat blockStat
}

//Options define the parameters for read and recover mode
type Options struct {
	//Degrade tells if degrade read is on
	Degrade bool
}



//global system-level variables
var (
	err error
)

//constant variables
const (
	blkOK         blockStat = 0
	blkFail       blockStat = 1
	tempFile                = "./test/file.temp"
	maxGoroutines           = 10240
)


//ReadDiskPath reads the disk paths from diskFilePath.
//There should be exactly ONE disk path at each line.
//
//This func can NOT be called concurrently.
func (n *Node) ReadDiskPath() error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	f, err := os.Open(e.DiskFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	e.diskInfos = make([]*diskInfo, 0)
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
s