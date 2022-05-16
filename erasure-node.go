// Package grasure is an Universal Erasure Coding Architecture in Go
//
// For usage and examples, see https://github.com/DurantVivado/Grasure
//
package grasure

import (
	"context"
	"fmt"
	"strconv"
	"sync"
)

type NodeType string

const (
	DATA      = "DATA"
	CLIENT    = "CLIENT"
	SERVER    = "SERVER"
	COMPUTING = "COMPUTING"
	GATEWAY   = "GATEWAY"
	NAME      = "NAME"
)

type NodeStat string

const (
	NodeInit     NodeStat = "NodeInit"
	HealthOK     NodeStat = "HealthOK"
	CPUFailed    NodeStat = "CPUFailed"
	DiskFailed   NodeStat = "DiskFailed"
	NetworkError NodeStat = "NetworkError"
)

type Node struct {
	//uid is the node's unique id in the cluster
	uid int64

	//addr is the IP address
	addr string

	//nodeType is the type of the node
	nodeType []NodeType

	//reduandancy specifies the node-level redundnacy policy to ensure data availability
	redun Redundancy

	//stat is the state of the node, every node should be informed of other nodes' state
	stat NodeStat

	//for storage nodes
	diskArrays *DiskArray

	//For name node:
	//FileMeta lists, indicating fileName, fileSize, fileHash, fileDist...
	FileMeta sync.Map

	// volume parametes
	volume *Volume

	// // the replication factor for config file
	// ReplicateFactor int

	ctx context.Context
}

func NewNode(ctx context.Context, id int, addr string, nodeType []NodeType, redun Redundancy) *Node {
	//initialize various nodes w.r.t types
	newnode := &Node{
		uid:      int64(c.hash([]byte(strconv.Itoa(id)))),
		addr:     addr,
		nodeType: nodeType,
		ctx:      ctx,
		stat:     NodeInit,
		redun:    redun,
		volume:   &Volume{0, 0, 0},
	}

	return newnode

}

func (n *Node) isRole(role string) bool {
	for _, r := range n.nodeType {
		if string(r) == role {
			return true
		}
	}
	return false
}

func (n *Node) getRole() []string {
	ret := []string{}
	for _, r := range n.nodeType {
		ret = append(ret, string(r))
	}
	return ret
}

func (n *Node) getState() string {
	return fmt.Sprintf("state:%d;total:%d;used:%d", n.stat, n.volume.Total, n.volume.Used)
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
