package grasure

import (
	"bufio"
	"context"
	"github.com/DurantVivado/GrasureOL/xlog"
	"hash/crc32"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultInfoFilePath = "cluster.info"
	defaultNodeFilePath = "examples/nodes.addr"
	defaultVirtualNum = 3

	defaultPort = ":9999"
)

type Mode string

const (
	InitMode Mode = "Init"
	NormalMode Mode = "Normal"
	DegradedMode Mode = "Degraded"
	RecoveryMode Mode = "Recovery"
	ScaleMode Mode = "Scale"
	PowerSaveMode Mode = "PowerSave"
)

type ClusterOption struct{
	verbose bool
	override int // 1:true, 0:false, 2:ABA (ask before action)
}

var defaultClusterOption = &ClusterOption{
	verbose: true,
	override: 2,
}

//Hash maps bytes to uint32
type Hash func(key []byte) uint32

type Cluster struct{
	//uuid is the global unique id of a cluster
	uuid int64

	//the mode on which the cluster operates
	mode Mode

	//acl represents access control and database
	acl *ACL

	//consistentHash func
	hash Hash

	//nodeMap contains all the node using consistent hash algorithm
	nodeMap  map[int64]*Node

	//nodeList is a list containing all nodes' uuid, with only the first usedNodeNum nodes used
	nodeList []int64

	// virtualNum is the number of virtual nodes in consistent hash
	virtualNum int

	//usedNodeNum is the number of used nodes in given NodeFilePath
	usedNodeNum int

	//aliveNodeNum is the number of nodes connected to the server
	aliveNodeNum int32

	//pools are referred to as a specific group of erasure coding, in the form (codeType-K-M-BlockSize),
	//e.g., rs-3-2-1024 is one pool, but xor-3-2-1024 refers to another.
	pools []*ErasurePool

	//Used, Free, Total are capacity-related parameters
	Used    uint64
	Free    uint64
	Total   uint64

	//InfoFilePath is the path of a textual file storing basic information of the cluster
	InfoFilePath string `json:"-"`

	//NodeFilePath is the path of file includes all nodes address, provided by user
	NodeFilePath string `json:"-"`

	options *ClusterOption `json:"-"`

	mu sync.Mutex
}


//NewCluster initializes a Cluster with customized ataShards, parityShards, usedNodeNum, replicateFactor and blockSize
func NewCluster(usedNodeNum int, hashfn Hash) *Cluster {
	//err = e.resetSystem()
	//if err != nil {
	//	return err
	//}
	//if !e.Quiet {
	//	fmt.Printf("System init!\n Erasure parameters: dataShards:%d, parityShards:%d,blocksize:%d,diskNum:%d\n",
	//		e.K, e.M, e.BlockSize, e.DiskNum)
	//}
	if hashfn == nil{
		hashfn = crc32.ChecksumIEEE
	}
	c := &Cluster{
		uuid:            genUUID(0),
		Used:            0,
		Free:            0,
		Total:           0,
		hash: hashfn,
		usedNodeNum: usedNodeNum,
		InfoFilePath:    defaultInfoFilePath,
		NodeFilePath: 	 defaultNodeFilePath,
		nodeMap: make(map[int64]*Node),
		nodeList: make([]int64, 0),
		virtualNum: defaultVirtualNum,
		options: defaultClusterOption,
	}
	return c
}

//AddNode adds a node into cluster using ConsistentHashAlgorithm
func (c* Cluster) AddNode(uuid int64, node *Node){
	for i := 0; i < c.virtualNum; i++ {
		hashVal := int64(c.hash([]byte(strconv.Itoa(i) + strconv.FormatInt(uuid, 10))))
		c.nodeList = append(c.nodeList, hashVal)
	}
	c.nodeMap[uuid] = node
	sortInt64(c.nodeList)
}


//GetIPsFromRole returns IP address according to given role
func (c *Cluster) GetIPsFromRole(role string) (addrs []string){

	for _,node := range c.nodeMap{
		if node.isRole(role){
			addrs = append(addrs, node.addr)
		}
	}
	return
}

//GetNodesFromRole returns Node slice according to given role
func (c *Cluster) GetNodesFromRole(role string) (nodes []*Node){
	for _,node := range c.nodeMap{
		if node.isRole(role){
			nodes = append(nodes, node)
		}
	}
	return
}

//ReadNodesAddr reads the node information from file
func (c* Cluster) ReadNodesAddr(){
	c.mu.Lock()
	defer c.mu.Unlock()
	//parse the node_addr file
	f, err := os.Open(c.NodeFilePath)
	if err != nil {
		xlog.Error(err)
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	id := int64(1)

	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			xlog.Error(err)
		}
		lineStr := string(line)
		if len(lineStr) == 0 || strings.HasPrefix(lineStr, "//"){
			continue
		}
		lineArr := strings.Split(lineStr, " ")
		addr := lineArr[0]
		nodeTyp := int16(0)
		for _, role := range strings.Split(lineArr[1],","){
			nodeTyp = nodeTyp | getType(role)
		}

		node, err := NewNode(id, addr, NodeType(nodeTyp))
		if err != nil{
			xlog.Error(err)
		}
		if id <= int64(c.usedNodeNum) {
			c.AddNode(node.uid, node)
		}
		c.nodeList = append(c.nodeList, node.uid)
		id++
	}
}


//Run listens on certain port and connects to the ServeFunc for specified node
func (c *Cluster)checkNodes(ctx context.Context, port string){
	l, err := net.Listen("tcp", port)
	if err != nil{
		xlog.Fatal(err)
	}
	xlog.Info("Server listening on:", port)
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				xlog.Errorf(ctx.Err().Error())
				return
			default:
				continue
			}
		}
		remoteAddr := conn.RemoteAddr().String()
		for _,node := range c.nodeMap {
			ipAddr := strings.Split(remoteAddr, ":")[0]
			if ipAddr == node.addr {
				atomic.AddInt32(&c.aliveNodeNum, 1)
				xlog.Infof("node:%s successfully connected", conn.RemoteAddr().String())
			}
			if int(c.aliveNodeNum) == c.usedNodeNum {
				xlog.Infoln("all nodes successfully connected")
				conn.Close()
				return
			}
		}
	}
}
//ConnectNodes connect to all nodes and if successful, return nil
func(c *Cluster) ConnectNodes(port string, expireDuration time.Duration){
	//Every node must connect to other nodes to testify connection
	ctx, cancel := context.WithTimeout(context.Background(), expireDuration)
	defer cancel()
	c.aliveNodeNum ++
	c.checkNodes(ctx, port)
}
/*
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

*/