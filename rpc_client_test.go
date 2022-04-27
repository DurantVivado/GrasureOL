package grasure

import (
	"fmt"
	"github.com/DurantVivado/GrasureOL/codec"
	"log"
	"sync"
	"testing"
	"time"
)

func TestClient_Call(t *testing.T) {
	newServer := NewServer()
	addr := ":8888"
	go newServer.Start("tcp", addr)

	client, err := Dial("tcp", addr, &Option{
		MagicNumber: 0x123456,
		CodecType: codec.GobType,
	})
	if err != nil{
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("req %d", i)
			var reply string
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}
