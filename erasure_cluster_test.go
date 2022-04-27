package grasure

import (
	"context"
	"log"
	"testing"
)

func TestCluster_ReadNodeDir(t *testing.T){
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = NewCluster(ctx, 2, nil)
	log.Println("TestCluster_ReadNodeDir OK")
}

func TestCluster_ConnectNodes(t *testing.T) {
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()
	c := NewCluster(ctx, 2, nil)
	c.ConnectNodes(defaultPort, defaultNodeConnectExpireTime)
}
