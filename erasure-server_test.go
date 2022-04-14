package grasure

import (
	"encoding/json"
	"fmt"
	"github.com/DurantVivado/GrasureOL/codec"
	"log"
	"net"
	"testing"
	"time"
)


func TestServer_ServeConn(t *testing.T) {
	newServer := NewServer()
	addr := ":8888"
	go newServer.Start("tcp", addr)

	// in fact, following code is like a simple geerpc client
	conn, err := net.Dial("tcp", addr)
	if err != nil{
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)
	// send options
	err = json.NewEncoder(conn).Encode(DefaultOption)
	if err != nil{
		t.Fatal(err)
	}
	cc := codec.NewGobCodec(conn)
	// send request & receive response
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Erasure.encode",
			Seq:           uint64(i),
		}
		err = cc.Write(h, fmt.Sprintf("req %d", h.Seq))
		if err != nil{
			t.Fatal(err)
		}
		err = cc.ReadHeader(h)
		if err != nil{
			t.Fatal(err)
		}
		var reply string
		err = cc.ReadBody(&reply)
		if err != nil{
			t.Fatal(err)
		}
		log.Println("reply:", reply)
	}
}
