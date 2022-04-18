// Package grasure is an Universal Erasure Coding Architecture in Go
//
// For usage and examples, see https://github.com/DurantVivado/Grasure
//
package grasure

import (
	"github.com/DurantVivado/GrasureOL/xlog"
	"github.com/DurantVivado/reedsolomon"
	"net"
	"strings"
)

type NodeType int16

func getType(role string) int16{
	if strings.ToUpper(role) == "CLIENT"{
		return 1
	}else if strings.ToUpper(role) == "SERVER"{
		return 1<<1
	}else if strings.ToUpper(role) == "COMPUTING"{
		return 1<<2
	}else if strings.ToUpper(role) == "GATEWAY"{
		return 1<<3
	}else if strings.ToUpper(role) == "DATA"{
		return 1<<4
	}else if strings.ToUpper(role) == "NAME"{
		return 1<<5
	}else{
		xlog.Errorf("NodeType Undefined")
	}
	return 0
}

//A node may function as multiple types
//low  |  1	 |   1	 | 	   1 	| 	 1	 |  1  |  1  | high
//Type:Client, Server, Computing, Gateway, Data, Name
const (
	ClientNode = 1<<iota
	ServerNode
	ComputingNode
	GateWayNode
	DataNode
	NameNode
	//Add new kind of node here
)

type NodeHealth string

const (
	HealthOK NodeHealth = "HealthOK"
	CPUFailed NodeHealth = "CPUFailed"
	DiskFailed NodeHealth = "DiskFailed"
	NetworkError NodeHealth = "NetworkError"
)

type Node struct {

	uid int64

	addr string

	nodeType NodeType

	health NodeHealth

	conn *net.Conn

	//For storage node:
	//a meta structure indicates the blocks belong to which file and index.
	blockArr []*blockInfo

	//For brevity's concern, we omit the disk infos
	//the root directory of data blob
	dataRoot string
	//the root directory of meta file
	metaRoot string

	//For name node:
	//FileMeta lists, indicating fileName, fileSize, fileHash, fileDist...
	FileMeta []*fileInfo `json:"fileLists"`

	//For computing node:
	// the reedsolomon streaming encoder, for streaming access
	sEnc reedsolomon.StreamEncoder

	// the reedsolomon encoder, for block access
	enc reedsolomon.Encoder

	//how many stripes are allowed to encode/decode concurrently
	ConStripes int `json:"-"`

	// the replication factor for config file
	ReplicateFactor int


	// the data stripe size, equal to k*bs
	dataStripeSize int64

	// the data plus parity stripe size, equal to (k+m)*bs
	allStripeSize int64

	//// the disk number, only the first diskNum disks are used in diskPathFile
	//DiskNum int `json:"diskNum"`
	//
	//// diskInfo lists
	//diskInfos []*diskInfo
	//
	//// the path of file recording all disks path
	//DiskFilePath string `json:"-"`

}

func NewNode(id int64, addr string, nodeType NodeType) (*Node, error){
	if id < 1{
		return nil, errNodeIdLessThanOne
	}
	//initialize various nodes w.r.t types
	newnode := &Node{
		uid:          genUUID(id),
		addr:         addr,
		nodeType:     nodeType,
		health:       HealthOK,
	}
	return newnode, nil

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





//Options define the parameters for read and recover mode
type Options struct {

}



//global system-level variables
var (
	err error
)




//ReadDiskPath reads the disk paths from diskFilePath.
//There should be exactly ONE disk path at each line.
//
//This func can NOT be called concurrently.
//func (n *Node) ReadDiskPath() error {
//	e.mu.RLock()
//	defer e.mu.RUnlock()
//	f, err := os.Open(e.DiskFilePath)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	buf := bufio.NewReader(f)
//	e.diskInfos = make([]*diskInfo, 0)
//	for {
//		line, _, err := buf.ReadLine()
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			return err
//		}
//		path := string(line)
//		if ok, err := pathExist(path); !ok && err == nil {
//			return &diskError{path, "disk path not exist"}
//		} else if err != nil {
//			return err
//		}
//		metaPath := filepath.Join(path, "META")
//		flag := false
//		if ok, err := pathExist(metaPath); ok && err == nil {
//			flag = true
//		} else if err != nil {
//			return err
//		}
//		diskInfo := &diskInfo{diskPath: string(line), available: true, ifMetaExist: flag}
//		e.diskInfos = append(e.diskInfos, diskInfo)
//	}
//	return nil
//}
