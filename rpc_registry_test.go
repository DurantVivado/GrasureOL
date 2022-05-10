package grasure

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	RegistryHandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func startRegistryServer(registryAddr string, wg *sync.WaitGroup) {
	var foo Foo
	l, _ := net.Listen("tcp", ":8888")
	server := NewServer()
	_ = server.Register(&foo)
	ctx, _ := context.WithCancel(context.Background())
	Heartbeat(ctx, registryAddr, "localhost:8888", 2*time.Second)
	wg.Done()
	server.Accept(l)
}

func rcall(registry string) {
	d := NewRegistryDiscovery(registry, 0)
	xc := NewXClient(d, RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fooLog(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func rbroadcast(registry string) {
	d := NewRegistryDiscovery(registry, 0)
	xc := NewXClient(d, RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		func(i int) {
			defer wg.Done()
			fooLog(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// expect 2 - 5 timeout
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			fooLog(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func TestRegistryHandleHTTP(t *testing.T) {
	registryAddr := "http://localhost:9999/grasure/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(2)
	startRegistryServer(registryAddr, &wg)
	go startRegistryServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	rcall(registryAddr)
	rbroadcast(registryAddr)
}
