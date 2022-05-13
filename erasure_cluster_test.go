package grasure

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/DurantVivado/GrasureOL/xlog"
)

var wg sync.WaitGroup

func TestCluster_Output1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	c.SetOuput(DebugLevel, "cluster.log")
	wg.Add(1)
	c.StartHeartbeatServer(":9999")
	wg.Wait()
}

func TestCluster_Output2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	c.SetOuput(DebugLevel, "")
	wg.Add(1)
	c.StartHeartbeatServer(":9999")
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
	c.StartHeartbeatServer(":9999")
	wg.Wait()

}

func TestCluster_ErasurePool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	c.SetOuput(DebugLevel, "")
	go c.StartHeartbeatServer(":9999")
	c.AddErasurePools(
		NewErasurePool(Erasure_RS, 2, 2, 4, 1024, nil, Random),
		NewErasurePool(Erasure_RS, 3, 2, 5, 4096, nil, LeftAsymmetric),
	)

}

func TestCluster_Write(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCluster(ctx, 5, nil)
	c.SetOuput(DebugLevel, "")
	go c.StartHeartbeatServer(defaultHeartbeatPort)
	c.AddErasurePools(
		NewErasurePool(Erasure_RS, 2, 2, 4, 1024, nil, Random),
		NewErasurePool(Erasure_RS, 3, 2, 5, 4096, nil, LeftAsymmetric),
	)
	f, err := os.Open("input/README.md")
	if err != nil {
		xlog.Fatal(err)
	}
	defer f.Close()
	encodePattern := fmt.Sprintf("%s-%d-%d-%d", Erasure_RS, 2, 2, 1024)
	p, ok := c.CheckIfExistPool(encodePattern)
	if !ok {
		c.AddErasurePools(NewErasurePool(Erasure_RS, 2, 2, 4, 1024, nil, Random))
	}
	go c.StartDFSServer(defaultWritePort)
	// for _, node := range p.nodes {
	httpAddr := fmt.Sprintf("http://%s%s%s", "192.168.58.136", defaultWritePort, defaultWritePath)
	p.Write(f, httpAddr)
	// }
}
