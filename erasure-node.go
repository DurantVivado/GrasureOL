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
	ClientNode NodeType = 1 << iota
	ServerNode
	ComputingNode
	GateWayNode
	DataNode
	NameNode
	//Add new kind of node here
	TestNode
)

type NodeStat int

const (
	NodeInit NodeStat = iota
	HealthOK
	CPUFailed
	DiskFailed
	NetworkError
)

type Node struct {
	//uid is the node's unique id in the cluster
	uid int64

	//addr is the IP address
	addr string

	//nodeType is the type of the node
	nodeType NodeType

	//reduandancy specifies the node-level redundnacy policy to ensure data availability
	redun Redundancy

	//stat is the state of the node, every node should be informed of other nodes' state
	stat NodeStat

	//for storage nodes
	diskArrays *DiskArray

	// //For name node:
	// //FileMeta lists, indicating fileName, fileSize, fileHash, fileDist...
	// FileMeta []*fileInfo `json:"fileLists"`

	// //how many stripes are allowed to encode/decode concurrently
	// ConStripes int `json:"-"`

	// // the replication factor for config file
	// ReplicateFactor int

	ctx context.Context
}

//Namenode stores the meta data of cluster. There can be multiple
//Namenode in the cluster.
type Namenode struct {
	Node
}

func NewNode(ctx context.Context, id int, addr string, nodeType NodeType, redun Redundancy) *Node {
	//initialize various nodes w.r.t types
	newnode := &Node{
		uid:      int64(c.hash([]byte(strconv.Itoa(id)))),
		addr:     addr,
		nodeType: nodeType,
		ctx:      ctx,
		stat:     NodeInit,
		redun:    redun,
	}
	if newnode.isRole("DATA") {
		newnode.diskArrays = NewDiskArray(defaultDiskFilePath)
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
