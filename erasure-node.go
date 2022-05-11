// Package grasure is an Universal Erasure Coding Architecture in Go
//
// For usage and examples, see https://github.com/DurantVivado/Grasure
//
package grasure

import (
	"context"
	"strconv"
	"strings"

	"github.com/DurantVivado/GrasureOL/xlog"
	"github.com/DurantVivado/reedsolomon"
)

type NodeType int16

func getType(role string) int16 {
	if strings.ToUpper(role) == "CLIENT" {
		return 1
	} else if strings.ToUpper(role) == "SERVER" {
		return 1 << 1
	} else if strings.ToUpper(role) == "COMPUTING" {
		return 1 << 2
	} else if strings.ToUpper(role) == "GATEWAY" {
		return 1 << 3
	} else if strings.ToUpper(role) == "DATA" {
		return 1 << 4
	} else if strings.ToUpper(role) == "NAME" {
		return 1 << 5
	} else {
		xlog.Errorf("NodeType Undefined")
	}
	return 0
}

//A node may function as multiple types
//low  |  1	 |   1	 | 	   1 	| 	 1	 |  1  |  1  | high
//Type:Client, Server, Computing, Gateway, Data, Name
const (
	ClientNode = 1 << iota
	ServerNode
	ComputingNode
	GateWayNode
	DataNode
	NameNode
	//Add new kind of node here
	TestNode
)

type NodeStat string

const (
	NodeInit     NodeStat = "NodeInit"
	HealthOK     NodeStat = "HealthOK"
	CPUFailed    NodeStat = "CPUFailed"
	DiskFailed   NodeStat = "DiskFailed"
	NetworkError NodeStat = "NetworkError"
)

//Options define the parameters for read and recover mode
type Options struct {
}

type Node struct {
	uid int64

	addr string

	nodeType NodeType

	stat NodeStat

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

	ctx context.Context
	opt *Options
}

//StorageNode is a node that storages data and parity blocks
type StorageNode struct {
	Node

	//whether data is re-encoded in the node
	redundancy Redundancy

	// the used disk number, only the first diskNum of disks are used in diskPathFile
	usedDiskNum int `json:"diskNum"`

	// diskInfo lists
	diskInfos []*diskInfo

	// the path of file recording all disks' path
	// default to the project's root path
	diskFilePath string `json:"-"`
}

//Namenode stores the meta data of cluster. There can be multiple
//Namenode in the cluster.
type Namenode struct {
	Node
}

func NewNode(ctx context.Context, id int, addr string, nodeType NodeType) *Node {
	//initialize various nodes w.r.t types
	newnode := &Node{
		uid:      int64(c.hash([]byte(strconv.Itoa(id)))),
		addr:     addr,
		nodeType: nodeType,
		ctx:      ctx,
		stat:     NodeInit,
	}

	return newnode

}

func (n *Node) isRole(role string) bool {
	nT := int16(n.nodeType)
	return (nT & getType(role)) != 0
}

type BlockReadRequest struct {
	Address string
	Offset  uint64
	Size    uint64
}

type BlockReadResponse struct {
	Msg  string
	Data []byte
}
