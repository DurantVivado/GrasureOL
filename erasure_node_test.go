package grasure

import (
	"fmt"
	"net"
	"testing"
)

func TestNode_ConnectToCluster(t *testing.T) {
	c := NewCluster(3, nil)
	s := c.GetIPsFromRole("Server")
	if len(s) == 0{
		t.Fatal("No such role as server")
	}
	targetAddr := s[0] + defaultPort
	conn, err := net.Dial("tcp",targetAddr)
	defer conn.Close()
	if err != nil{
		t.Fatal(err)
	}
	fmt.Println("Successfully connect to the cluster")
}