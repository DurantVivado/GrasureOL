package grasure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DurantVivado/GrasureOL/codec"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)



func TestServer_ServeConn(t *testing.T) {
	addr := make(chan string)
	var foo Foo
	go startServer(addr, foo)


	// in fact, following code is like a simple geerpc client
	conn, err := net.Dial("tcp", <-addr)
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

func startServer(addr chan string, service interface{}) {
	if err := Register(service); err != nil {
		log.Fatal("register error:", err)
	}
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	Accept(l)
}

func TestServer_Register(t *testing.T) {
	addr := make(chan string)
	var foo Foo
	go startServer(addr, foo)
	client, _ := Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	ctx, _ := context.WithCancel(context.Background())
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(ctx,"Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}
