package grasure

import (
	"context"
	"fmt"
	"testing"
)

func TestNode_ConnectToCluster(t *testing.T) {
	ctx,_ := context.WithCancel(context.Background())
	//defer cancel()
	c := NewCluster(ctx, 2, nil)
	s := c.GetIPsFromRole("Server")
	if len(s) == 0{
		t.Fatal("No such role as server")
	}
	node := c.GetLocalNode()
	if node == nil{
		node = NewNode(ctx, 999, "127.0.0.1", TestNode)
	}
	node.ConnectToCluster(s[0], defaultPort,defaultHeartbeatDuration)
	fmt.Println("Successfully connect to the cluster")
}