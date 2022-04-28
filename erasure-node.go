// Package grasure is an Universal Erasure Coding Architecture in Go
//
// For usage and examples, see https://github.com/DurantVivado/Grasure
//
package grasure

import (
	"context"
	"github.com/DurantVivado/GrasureOL/codec"
	"github.com/DurantVivado/GrasureOL/xlog"
	"github.com/DurantVivado/reedsolomon"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
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

var prevStat NodeStat

//every several seconds, the node send heartbeats to the
//server to notice its survival
func (n *Node) sendHeartBeats(conn net.Conn, start time.Time) {
	defer conn.Close()
	timer := time.NewTimer(defaultHeartbeatDuration)
	for {
		select {
		case <-timer.C:
			//we use RPC
			cc := codec.NewJsonCodec(conn)
			h := &codec.Header{
				ServiceMethod: "Node.HeartBeat",
				Seq:           uint64(n.uid),
			}
			body := time.Since(start).String()
			err := cc.Write(h, body)
			if err != nil {
				xlog.Fatal(err)
			}
		case <-n.ctx.Done():
			xlog.Error(n.ctx.Err())
			return
		}
	}
}

// ConnectToCluster connects to the cluster every duration time,
//if successful, then launch the heartbeat RPC server
func (n *Node) ConnectToCluster(targetAddr, port string, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			conn, err := net.Dial("tcp", targetAddr+port)
			if err != nil {
				prevStat = n.stat
				n.stat = NetworkError
				xlog.Error(err)
				continue
			}
			n.stat = prevStat
			//send heart beat and lasting time
			n.sendHeartBeats(conn, time.Now())
			return
		case <-n.ctx.Done():
			return
		}
	}
}

func (n *Node) handleRequests(conn io.ReadWriteCloser) {
	var cc = codec.NewGobCodec(conn)
	h := &codec.Header{}
	err := cc.ReadHeader(h)
	if err != nil {
		xlog.Fatal(err)
	}
	switch h.ServiceMethod {
	case "Erasure.Read":
		req := &BlockReadRequest{}
		err = cc.ReadBody(req)
		if err != nil {
			xlog.Fatal(err)
		}
		//read local file
	default:

	}
}

func (n *Node) ListenWorkPort(ctx context.Context, port string) {
	l, err := net.Listen("tcp", port)
	if err != nil {
		return
	}
	xlog.Info("node listening on:", port)
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				xlog.Errorf(ctx.Err().Error())
				return
			default:
				xlog.Error(err)
				continue
			}
		}
		xlog.Infof("node:%s successfully connected", conn.RemoteAddr().String())
		go n.handleRequests(conn)
	}
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

//readFromNode read certain segment of data (fixed with offset and len in bytes)
//from the other node using RPC, return the number of bytes read and an error
func (n *Node) readFromNode(address string, offset, size uint64) ([]byte, error) {
	if size == 0 {
		return make([]byte, 0), nil
	}
	req := &BlockReadRequest{
		Address: address,
		Offset:  offset,
		Size:    size,
	}
	//send via RPC
	client, err := Dial("tcp", address, &Option{
		MagicNumber: READ_MAGIC_NUMBER,
		CodecType:   codec.JsonType,
	})
	if err != nil {
		return nil, err
	}
	reply := &BlockReadResponse{}
	if err := client.Call(n.ctx, "Erasure.Read", req, &reply); err != nil {
		return nil, err
	}
	xlog.Infoln("reply:", reply.Msg, reply.Data)
	return reply.Data, nil
}
