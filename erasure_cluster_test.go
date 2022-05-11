package grasure

import (
	"context"
	"log"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestCluster_Output1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	c.SetOuput(DebugLevel, "cluster.log")
	wg.Add(1)
	c.StartServer(":9999")
	wg.Wait()
}

func TestCluster_Output2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	c.SetOuput(DebugLevel, "")
	wg.Add(1)
	c.StartServer(":9999")
	wg.Wait()
}

func TestCluster_ReadNodeDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	log.Println("TestCluster_ReadNodeDir OK")
}

func TestCluster_Registry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	wg.Add(1)
	c.StartServer(":9999")
	wg.Wait()

}

func TestCluster_ErasurePool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	c.SetOuput(DebugLevel, "")
	go c.StartServer(":9999")
	c.AddErasurePools(
		NewErasurePool(RS.String(), 2, 2, 4, 1024, nil),
		NewErasurePool(RS.String(), 3, 2, 5, 4096, nil),
	)

}
