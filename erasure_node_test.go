package grasure

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNode_ConnectToCluster(t *testing.T) {
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()
	c := NewCluster(ctx, 3, nil)
	c.ReadNodesAddr()
	s := c.GetIPsFromRole("Server")
	if len(s) == 0{
		t.Fatal("No such role as server")
	}
	node := c.GetLocalNode()
	if node == nil{
		node, _ = NewNode(-9999, "127.0.0.1", TestNode)
	}
	node.ConnectToCluster(s[0], defaultPort,3*time.Second)
	fmt.Println("Successfully connect to the cluster")
}