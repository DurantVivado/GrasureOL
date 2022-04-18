package grasure

import (
	"log"
	"testing"
)

func TestErasure_ReadNodeDir(t *testing.T){
	c := NewCluster(3, nil)
	err := c.ReadNodesAddr()
	if err != nil {
		t.Error(err)
	}
	log.Println("TestErasure_ReadNodeDir OK")
}
