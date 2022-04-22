package grasure

import (
	"log"
	"testing"
	"time"
)

func TestCluster_ReadNodeDir(t *testing.T){
	c := NewCluster(3, nil)
	c.ReadNodesAddr()
	log.Println("TestCluster_ReadNodeDir OK")
}

func TestCluster_ConnectNodes(t *testing.T) {
	c := NewCluster(3, nil)
	c.ReadNodesAddr()
	c.ConnectNodes(":9999", 100*time.Second)
}
