package grasure

import (
	"context"
	"net/http"
	"strings"

	"github.com/DurantVivado/GrasureOL/xlog"
)

type Context context.Context

//handleHeartbeat is a handler, after receiving heartbeat messages, the server mark certain node as alive
func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()
	remoteAddr := strings.Split(r.RemoteAddr, ":")[0]
	//we set the node as alive
	for _, node := range c.nodeMap {
		ipAddr := strings.Split(remoteAddr, ":")[0]
		if ipAddr == node.addr && node.stat != HealthOK {
			c.aliveNodeNum++
			xlog.Infof("node:%s successfully connected", ipAddr)
			node.stat = HealthOK
		}
		if int(c.aliveNodeNum) == c.usedNodeNum {
			xlog.Infoln("all nodes successfully connected")
			return
		}

	}
	xlog.Infoln("recv heartbeat from :", remoteAddr)
}
