package grasure

import (
	"bufio"
	"context"
	"hash/crc32"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/DurantVivado/GrasureOL/xlog"
)

const (
	defaultInfoFilePath          = "cluster.info"
	defaultNodeFilePath          = "nodes.addr"
	defaultVirtualNum            = 3
	defaultReplicaNum            = 3
	defaultRedundancy            = None
	defaultHeartbeatDuration     = 3 * time.Second
	defaultNodeConnectExpireTime = 100 * time.Second
	defaultHeartbeatPort         = ":9999"
	defaultWritePort             = ":8888"
	defaultConnectionTimeout     = 10 * time.Second
	defaultHandleTimeout         = 0
	connected                    = "200 Connected to Grasure RPC"
	defaultRPCPath               = "/grasure"
	defaultDebugPath             = "/debug/grasureRPC"
	defaultWritePath             = "/grasure/write"
	defaultLogLevel              = InfoLevel
	defaultLogFile               = "grasure.log"
	defaultDiskFilePath          = "disks.path"
	defaultDataLayout            = Random
)

type LogLevel int

const (
	Disabled LogLevel = iota
	FatalLevel
	ErrorLevel
	InfoLevel
	DebugLevel
)

type Volume struct {
	Used  uint64 `json:"Used"`
	Free  uint64 `json:"Free"`
	Total uint64 `json:"Total"`
}

func (v *Volume) getUsage() (uint64, uint64) {
	return v.Total, v.Free
}

type Redundancy string

const (
	Erasure_RS  Redundancy = "Erasure_RS"
	Erasure_XOR Redundancy = "Erasure_XOR"
	Erasure_LRC Redundancy = "Erasure_LRC"
	Replication Redundancy = "Replication"
	None        Redundancy = "None"
)

type Mode string

const (
	InitMode      Mode = "InitMode"
	NormalMode    Mode = "NormalMode"
	DegradedMode  Mode = "DegradedMode"
	RecoveryMode  Mode = "RecoveryMode"
	ScaleMode     Mode = "ScaleMode"
	PowerSaveMode Mode = "PowerSaveMode"
)

type Output int

const (
	CONSOLE Output = iota
	LOGFILE
	NONE
)

type ClusterOption struct {
	verbose  bool
	override bool // 1:true, 0:false, default to 0
	output   Output
	logFile  string
}

var defaultClusterOption = &ClusterOption{
	verbose:  true,
	override: false,
	output:   CONSOLE,
	logFile:  defaultLogFile,
}

//Hash maps bytes to uint32
type Hash func(key []byte) uint32

var once sync.Once

//Cluster is an instance that is only created once at each node
type Cluster struct {
	//UUID is the global unique id of a cluster
	UUID int64 `json:"UUID"`

	//the mode on which the cluster operates
	mode Mode

	//localAddr the local addr the Cluster operates on
	localAddr []string

	//do the local node belongs to the Cluster? If no, we regard it as a Client.
	inCluster bool

	//is the local node the server node? Only server node gen UUID for other nodes.
	isServer bool

	//the server module of the current node
	server *Server

	//the client module of the current node
	client *Client

	//acl represents access control and database
	acl *ACL

	//consistentHash func
	hash Hash

	//localNode is the local node instance
	localNode *Node

	//nodeMap contains all the node using consistent hash algorithm: hash_num->*Node
	nodeMap sync.Map

	//nodeList is a list containing all nodes' UUID, with only the first usedNodeNum nodes used
	nodeList  []*Node
	nameNodes []*Node
	dataNodes []*Node

	//virtualList is a list containing all virtual node, each real node is mapped to virtualNum
	//virtualNode to ensure load balance
	virtualList []int64

	// virtualNum is the number of virtual nodes in consistent hash
	virtualNum int

	//usedNodeNum is the number of used nodes in given NodeFilePath
	usedNodeNum int

	//aliveNodeNum is the number of nodes connected to the server
	aliveNodeNum int32

	//redundancy is the redundancy policy adopted by the cluster
	redundancy Redundancy `json:"redundancy"`

	//pools are referred to as a specific group of erasure coding, in the form (codeType-K-M-BlockSize),
	//e.g., rs-3-2-1024 is one pool, but xor-3-2-1024 refers to another.string->*ErasurePool
	poolMap sync.Map

	//Used, Free, Total are capacity-related parameters
	volume *Volume

	//InfoFilePath is the path of a textual file storing basic information of the cluster
	InfoFilePath string

	//NodeFilePath is the path of file includes all nodes address, provided by user
	NodeFilePath string

	options *ClusterOption

	mu sync.Mutex

	ctx context.Context
}

var c *Cluster

//NewCluster initializes a Cluster with customized ataShards, parityShards, usedNodeNum, replicateFactor and blockSize
func NewCluster(ctx context.Context, usedNodeNum int, hashfn Hash) *Cluster {
	//err = e.resetSystem()
	//if err != nil {
	//	return err
	//}
	//if !e.Quiet {
	//	fmt.Printf("System init!\n Erasure parameters: dataShards:%d, parityShards:%d,blocksize:%d,diskNum:%d\n",
	//		e.K, e.M, e.BlockSize, e.DiskNum)
	//}
	once.Do(func() {

		localAddr := getLocalAddr()
		if hashfn == nil {
			hashfn = crc32.ChecksumIEEE
		}
		c = &Cluster{
			UUID:         0,
			localAddr:    localAddr,
			inCluster:    false,
			isServer:     false,
			server:       NewServer(),
			client:       nil,
			volume:       &Volume{0, 0, 0},
			hash:         hashfn,
			usedNodeNum:  usedNodeNum,
			redundancy:   defaultRedundancy,
			InfoFilePath: defaultInfoFilePath,
			NodeFilePath: defaultNodeFilePath,
			nodeList:     make([]*Node, 0),
			dataNodes:    make([]*Node, 0),
			nameNodes:    make([]*Node, 0),
			virtualList:  make([]int64, 0),
			virtualNum:   defaultVirtualNum,
			options:      defaultClusterOption,
			ctx:          ctx,
		}
		//read the nodes infos via reading NodeFilePath
		c.ReadNodesAddr()
		s := c.GetIPsFromRole(SERVER)
		if len(s) == 0 {
			xlog.Fatal(errNoServerNodeInCluster)
		}
		n := c.GetIPsFromRole(NAME)
		if len(n) == 0 {
			xlog.Fatal(errNoNameNodeInCluster)
		}
		d := c.GetIPsFromRole(DATA)
		if len(d) == 0 {
			xlog.Fatal(errNoDataNodeInCluster)
		}

		for _, node := range c.nodeList {
			for _, addr := range localAddr {
				if node.addr == addr {
					c.localNode = node
					if node.isRole(SERVER) {
						c.isServer = true
						c.aliveNodeNum = 1
					}
					c.inCluster = true
					break
				}
			}

		}
		if c.localNode.isRole(DATA) {
			c.localNode.diskArrays = NewDiskArray(defaultDiskFilePath)
		}
		if !c.inCluster {
			xlog.Warn("You're operating a node out of the cluster, it will be treated as a client with access limitation.")
		}
		if c.usedNodeNum < 1 || c.usedNodeNum > len(c.nodeList) {
			xlog.Fatal(errInvalidUsedNodeNum)
		}
	})
	return c
}

func (c *Cluster) SetOuput(level LogLevel, filename string) {
	if filename != "" {
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		xlog.SetLevel(int(level), f)
	} else {
		xlog.SetLevel(int(level), os.Stdout)
	}

}

//AddNode adds a node into cluster using ConsistentHashAlgorithm
func (c *Cluster) AddNode(id int, node *Node) {
	for i := 0; i < c.virtualNum; i++ {
		hashVal := int64(c.hash([]byte(strconv.Itoa(i) + strconv.Itoa(id))))
		c.virtualList = append(c.virtualList, hashVal)
	}
	hashVal := int64(c.hash([]byte(strconv.Itoa(id))))
	c.nodeMap.Store(hashVal, node)
	sortInt64(c.virtualList)
}

//DelNode deletes a node from the map
func (c *Cluster) DelNode(id int) {
	for i := 0; i < c.virtualNum; i++ {
		hashVal := int64(c.hash([]byte(strconv.Itoa(i) + strconv.Itoa(id))))
		id := sort.Search(len(c.virtualList), func(i int) bool {
			return c.virtualList[i] >= hashVal
		})
		c.virtualList = append(c.virtualList[:id], c.virtualList[id+1:]...)
	}
	hashVal := int64(c.hash([]byte(strconv.Itoa(id))))
	c.nodeMap.Delete(hashVal)
}

//GetIPsFromRole returns IP address according to given role
func (c *Cluster) GetIPsFromRole(role string) (addrs []string) {

	c.nodeMap.Range(func(k, v interface{}) bool {
		node, _ := v.(*Node)
		if node.isRole(role) {
			addrs = append(addrs, node.addr)
			return true
		}
		return false
	})
	return
}

//GetNodesFromRole returns Node slice according to given role
func (c *Cluster) GetNodesFromRole(role string) (nodes []*Node) {
	c.nodeMap.Range(func(k, v interface{}) bool {
		node, _ := v.(*Node)
		if node.isRole(role) {
			nodes = append(nodes, node)
			return true
		}
		return false
	})
	return
}

//GetLocalNode returns the local node if exists in the cluster
func (c *Cluster) GetLocalNode() (ret *Node) {
	c.nodeMap.Range(func(k, v interface{}) bool {
		node, _ := v.(*Node)
		for _, addr := range c.localAddr {
			if addr == node.addr {
				ret = node
				return true
			}
		}
		return false
	})

	return ret
}

//ReadNodesAddr reads the node information from file
func (c *Cluster) ReadNodesAddr() {
	c.mu.Lock()
	defer c.mu.Unlock()
	//parse the node_addr file
	f, err := os.Open(c.NodeFilePath)
	if err != nil {
		xlog.Error(err)
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	id := 1

	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			xlog.Error(err)
		}
		lineStr := string(line)
		if len(lineStr) == 0 || strings.HasPrefix(lineStr, "//") {
			continue
		}
		lineArr := strings.Split(lineStr, " ")
		addr := lineArr[0]
		nodeTyp := make([]NodeType, 0)
		for _, role := range strings.Split(lineArr[1], ",") {
			nodeTyp = append(nodeTyp, NodeType(role))
		}
		node := NewNode(c.ctx, id, addr, nodeTyp, defaultRedundancy)
		if id <= c.usedNodeNum {
			c.AddNode(id, node)
		}
		if node.isRole(DATA) {
			c.dataNodes = append(c.dataNodes, node)
		}
		if node.isRole(NAME) {
			c.nameNodes = append(c.nameNodes, node)
		}
		c.nodeList = append(c.nodeList, node)
		id++
	}
}

//StartServer starts the server to handle HTTP requests
func (c *Cluster) StartHeartbeatServer(port string) {
	c.server.HandleHTTP()
	c.SetNodeStatus(2 * defaultHeartbeatDuration)
	if err := http.ListenAndServe(port, c.server); err != nil {
		xlog.Fatal(err)
	}

}

//StartDFSServer starts the server to handle HTTP requests
func (c *Cluster) StartDFSServer(port string) {
	c.server.HandleHTTP()
	if err := http.ListenAndServe(port, c.server); err != nil {
		xlog.Fatal(err)
	}

}

//SetNodeStatus reset nodes' status every duration
func (c *Cluster) SetNodeStatus(duration time.Duration) {
	//set the nodes as unconnected every interval
	go func() {
		t := time.NewTicker(duration)
		for {
			select {
			case <-t.C:
				c.nodeMap.Range(func(k, v interface{}) bool {
					node, _ := v.(*Node)
					if node.addr == c.localNode.addr {
						return false
					}
					if node.stat == NetworkError {
						xlog.Errorf("node:%s disconnected!\n", node.addr)
						c.aliveNodeNum--
						return false
					}
					node.stat = NetworkError
					return true
				})
			case <-c.ctx.Done():
				// xlog.Error(c.ctx.Error())
				return
			}
		}
	}()
}

//heartbeat enables nodes to send periodic message to server to monitor the node's health
func (c *Cluster) heartbeatToServer(registry, port string, duration time.Duration, wg *sync.WaitGroup) {
	l, err := net.Listen("tcp", port)
	if err != nil {
		xlog.Fatal(err)
	}
	Heartbeat(c.ctx, registry, port, duration)
	wg.Done()
	c.server.Accept(l)
}

//AddErasurePool adds an erasure-pool to the cluster's poolMap
//whose key should be like "RS-10-4-1024k", namely policy.
func (c *Cluster) AddErasurePools(pool ...*ErasurePool) {
	for _, p := range pool {
		key := string(p.Redun) + "-" + toString(p.K) + "-" + toString(p.M) + "-" + toString(p.BlockSize)
		c.poolMap.Store(key, p)
	}
}

//CheckIfExistPool checks if there exists pool of certain pattern, e.g. "RS-4-2-4096"
func (c *Cluster) CheckIfExistPool(pattern string) (*ErasurePool, bool) {
	if v, ok := c.poolMap.Load(pattern); ok {
		return v.(*ErasurePool), true
	} else {
		return nil, false
	}
}

//ClusterStatus is delpoyed on certain port of the server to inform the user of
//real-time status of the cluster
type ClusterStatus struct {
	UUID  string
	mode  string
	total uint64
	used  uint64
	free  uint64
	nodes []*NodeStatus
}

type NodeStatus struct {
	uid    int
	addr   string
	role   string
	status string
}

const templ = `<h1>Grasure Report</h1>
<h2>UUID:{{}}</h2>
<table>
{{range .Items}}
<tr>
  <td><a href='{{.HTMLURL}}'>{{.Number}}</a></td>
  <td>{{.State}}</td>
  <td><a href='{{.User.HTMLURL}}'>{{.User.Login}}</a></td>
  <td><a href='{{.HTMLURL}}'>{{.Title}}</a></td>
</tr>
{{end}}
</table>`

var report = template.Must(template.New("grasure report").Parse(templ))

func (c *Cluster) getClusterInfo(wr io.Writer) {
	s := &ClusterStatus{
		UUID:  toString(c.UUID),
		mode:  string(c.mode),
		total: c.volume.Total,
		used:  c.volume.Used,
		free:  c.volume.Free,
		nodes: make([]*NodeStatus, c.usedNodeNum),
	}
	c.nodeMap.Range(func(key, value interface{}) bool {
		node, _ := value.(*Node)

		s.nodes = append(s.nodes, &NodeStatus{
			uid:    int(node.uid),
			addr:   node.addr,
			role:   strings.Join(node.getRole(), ","),
			status: string(node.stat),
		})
		return false
	})
	if err := report.Execute(wr, s); err != nil {
		xlog.Fatal(err)
	}
}

//reset the storage
// func (c *Cluster) reset() error {

// 	g := new(errgroup.Group)

// 	for _, path := range e.diskInfos[:e.DiskNum] {
// 		path := path
// 		files, err := os.ReadDir(path.diskPath)
// 		if err != nil {
// 			return err
// 		}
// 		if len(files) == 0 {
// 			continue
// 		}
// 		g.Go(func() error {
// 			for _, file := range files {
// 				err = os.RemoveAll(filepath.Join(path.diskPath, file.Name()))
// 				if err != nil {
// 					return err
// 				}

// 			}
// 			return nil
// 		})
// 	}
// 	if err := g.Wait(); err != nil {
// 		return err
// 	}
// 	return nil
// }

//reset the system including config and data
// func (c *Cluster) resetSystem() error {

// 	//in-memory meta reset
// 	e.FileMeta = make([]*fileInfo, 0)
// 	// for k := range e.fileMap {
// 	// 	delete(e.fileMap, k)
// 	// }
// 	e.fileMap.Range(func(key, value interface{}) bool {
// 		e.fileMap.Delete(key)
// 		return true
// 	})
// 	err = e.WriteConfig()
// 	if err != nil {
// 		return err
// 	}
// 	//delete the data blocks under all diskPath
// 	err = e.reset()
// 	if err != nil {
// 		return err
// 	}
// 	err = e.replicateConfig(e.ReplicateFactor)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

//ReadConfig reads the config file during system warm-up.
//
//Calling it before actions like encode and read is a good habit.
// func (c *Cluster) ReadConfig() error {
// 	e.mu.Lock()
// 	defer e.mu.Unlock()

// 	if ex, err := pathExist(e.ConfigFile); !ex && err == nil {
// 		// we try to recover the config file from the storage system
// 		// which renders the last chance to heal
// 		err = e.rebuildConfig()
// 		if err != nil {
// 			return errConfFileNotExist
// 		}
// 	} else if err != nil {
// 		return err
// 	}
// 	data, err := ioutil.ReadFile(e.ConfigFile)
// 	if err != nil {
// 		return err
// 	}
// 	err = json.Unmarshal(data, &e)
// 	if err != nil {
// 		//if json file is broken, we try to recover it

// 		err = e.rebuildConfig()
// 		if err != nil {
// 			return errConfFileNotExist
// 		}

// 		data, err := ioutil.ReadFile(e.ConfigFile)
// 		if err != nil {
// 			return err
// 		}
// 		err = json.Unmarshal(data, &e)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	//initialize the ReedSolomon Code
// 	e.enc, err = reedsolomon.New(e.K, e.M,
// 		reedsolomon.WithAutoGoroutines(int(e.BlockSize)),
// 		reedsolomon.WithCauchyMatrix(),
// 		reedsolomon.WithInversionCache(true),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	e.dataStripeSize = int64(e.K) * e.BlockSize
// 	e.allStripeSize = int64(e.K+e.M) * e.BlockSize

// 	e.errgroupPool.New = func() interface{} {
// 		return &errgroup.Group{}
// 	}
// 	//unzip the fileMap
// 	for _, f := range e.FileMeta {
// 		stripeNum := len(f.Distribution)
// 		f.blockToOffset = makeArr2DInt(stripeNum, e.K+e.M)
// 		f.blockInfos = make([][]*blockInfo, stripeNum)
// 		countSum := make([]int, e.DiskNum)
// 		for row := range f.Distribution {
// 			f.blockInfos[row] = make([]*blockInfo, e.K+e.M)
// 			for line := range f.Distribution[row] {
// 				diskId := f.Distribution[row][line]
// 				f.blockToOffset[row][line] = countSum[diskId]
// 				f.blockInfos[row][line] = &blockInfo{bstat: blkOK}
// 				countSum[diskId]++
// 			}
// 		}
// 		//update the numBlocks
// 		for i := range countSum {
// 			e.diskInfos[i].numBlocks += countSum[i]
// 		}
// 		e.fileMap.Store(f.FileName, f)
// 		// e.fileMap[f.FileName] = f

// 	}
// 	e.FileMeta = make([]*fileInfo, 0)
// 	// we
// 	//e.sEnc, err = reedsolomon.NewStreamC(e.K, e.M, conReads, conWrites)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	return nil
// }

//Replicate the config file into the system for k-fold
//it's NOT striped and encoded as a whole piece.
// func (c *Cluster) replicateConfig(k int) error {
// 	selectDisk := genRandomArr(e.DiskNum, 0)[:k]
// 	for _, i := range selectDisk {
// 		disk := e.diskInfos[i]
// 		disk.ifMetaExist = true
// 		replicaPath := filepath.Join(disk.diskPath, "META")
// 		_, err = copyFile(e.ConfigFile, replicaPath)
// 		if err != nil {
// 			log.Println(err.Error())
// 		}

// 	}
// 	return nil
// }

//WriteConfig writes the cluster parameters and file information list into config files and transfer to other nodes
//
//Calling it after actions like encode and read is a good habit.
// func (c *Cluster) WriteConfig() error {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()

// 	f, err := os.OpenFile(c.InfoFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	// we marsh filemap into fileLists
// 	// for _, v := range e.fileMap {
// 	// 	e.FileMeta = append(e.FileMeta, v)
// 	// }
// 	// e.fileMap.Range(func(k, v interface{}) bool {
// 	// 	e.FileMeta = append(e.FileMeta, v.(*fileInfo))
// 	// 	return true
// 	// })
// 	data, err := json.Marshal(c)
// 	// data, err := json.MarshalIndent(e, " ", "  ")
// 	if err != nil {
// 		return err
// 	}
// 	buf := bufio.NewWriter(f)
// 	_, err = buf.Write(data)
// 	if err != nil {
// 		return err
// 	}
// 	buf.Flush()
// 	// f.Sync()
// 	err = c.updateConfigReplica()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

//reconstruct the config file if possible
// func (c *Cluster) rebuildConfig() error {
// 	//we read file meta in the disk path and try to rebuild the config file
// 	for i := range e.diskInfos[:e.DiskNum] {
// 		disk := e.diskInfos[i]
// 		replicaPath := filepath.Join(disk.diskPath, "META")
// 		if ok, err := pathExist(replicaPath); !ok && err == nil {
// 			continue
// 		}
// 		_, err = copyFile(replicaPath, e.ConfigFile)
// 		if err != nil {
// 			return err
// 		}
// 		break
// 	}
// 	return nil
// }

//update the config file of all replica
// func (c *Cluster) updateConfigReplica() error {

// 	//we read file meta in the disk path and try to rebuild the config file
// 	if e.ReplicateFactor < 1 {
// 		return nil
// 	}
// 	for i := range e.diskInfos[:e.DiskNum] {
// 		disk := e.diskInfos[i]
// 		replicaPath := filepath.Join(disk.diskPath, "META")
// 		if ok, err := pathExist(replicaPath); !ok && err == nil {
// 			continue
// 		}
// 		_, err = copyFile(e.ConfigFile, replicaPath)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
