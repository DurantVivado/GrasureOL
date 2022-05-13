package grasure

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/DurantVivado/GrasureOL/xlog"
)

type Context context.Context

func responseHeartbeat(httpAddr, addr string) error {
	// xlog.Infoln(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	buf := bytes.NewReader([]byte(c.getClusterInfo()))
	req, _ := http.NewRequest("HEARTBEAT", httpAddr, buf)
	req.Header.Set("X-Grasure-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		xlog.Fatalln("rpc server: heart beat err:", err)
		return err
	}
	return nil
}

//handleHeartbeat is a handler, after receiving heartbeat messages, the server mark certain node as alive
func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if int(c.aliveNodeNum) == c.usedNodeNum {
		c.mode = NormalMode
		xlog.Infoln("all nodes successfully connected")
		return
	}
	remoteAddr := strings.Split(r.RemoteAddr, ":")[0]
	//we set the node as alive
	c.nodeMap.Range(func(k, v interface{}) bool {
		node, _ := v.(*Node)
		ipAddr := strings.Split(remoteAddr, ":")[0]
		if ipAddr == node.addr && node.stat != HealthOK {
			c.aliveNodeNum++
			xlog.Infof("node:%s successfully connected", ipAddr)
			httpAddr := fmt.Sprintf("http://%s%s%s", ipAddr, defaultHeartbeatPort, defaultRPCPath)
			responseHeartbeat(httpAddr, ipAddr)
			node.stat = HealthOK
		}
		if int(c.aliveNodeNum) == c.usedNodeNum {
			// xlog.Infoln("all nodes successfully connected")
			return true
		}
		return false
	})
	// xlog.Infoln("recv heartbeat from :", remoteAddr)
}

//handleWrite is a handler that enables ServerNode and DataNode to receive files
func handleWrite(w http.ResponseWriter, r *http.Request) {
	xlog.Println(r.Header, r.Body)
}
