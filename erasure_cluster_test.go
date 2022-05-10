package grasure

import (
	"context"
	"log"
	"sync"
	"testing"
)

func TestCluster_ReadNodeDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 2, nil)
	log.Println("TestCluster_ReadNodeDir OK")
}

func TestCluster_Registry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 2, nil)
	var wg sync.WaitGroup
	wg.Add(1)
	c.StartServer(":9999")
	wg.Wait()
}
