package grasure

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestNode_HeartBeat(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	//defer cancel()
	c := NewCluster(ctx, 5, nil)
	go c.StartDFSServer(defaultWritePort)
	s := c.GetIPsFromRole(SERVER)
	if len(s) == 0 {
		t.Fatal("No such role as server")
	}
	// node := c.GetLocalNode()
	// if node == nil {
	// 	node = NewNode(ctx, 999, "127.0.0.1", TestNode)
	// }
	registry := fmt.Sprintf("http://%s%s%s", s[0], ":9999", defaultRegistryPath)
	var wg sync.WaitGroup
	wg.Add(1)
	c.heartbeatToServer(registry, ":9999", defaultHeartbeatDuration, &wg)
	wg.Done()
}
