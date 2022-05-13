package grasure

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DurantVivado/GrasureOL/xlog"
)

// Registry is a simple register center, provide following functions.
// add a server and receive heartbeat to keep it alive.
// returns all alive servers and delete dead servers sync simultaneously.
type Registry struct {
	timeout time.Duration
	mu      sync.Mutex // protect following
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultRegistryPath    = "/grasure/registry"
	defaultRegistryTimeout = time.Minute * 5
)

// NewRegistry create a registry instance with timeout setting
func NewRegistry(timeout time.Duration) *Registry {
	return &Registry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultRegister = NewRegistry(defaultRegistryTimeout)

func (r *Registry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now() // if exists, update start time to keep alive
	}
}

func (r *Registry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

// Runs at /_geerpc_/registry
func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		// keep it simple, server is in req.Header
		w.Header().Set("X-Grasure-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		// keep it simple, server is in req.Header
		addr := req.Header.Get("X-Grasure-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HandleHTTP registers an HTTP handler for Registry messages on registryPath
func (r *Registry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path:", registryPath)
}

func RegistryHandleHTTP() {
	DefaultRegister.HandleHTTP(defaultRegistryPath)
}

// Heartbeat send a heartbeat message every once in a while
// it's a helper function for a server to register or send heartbeat
func Heartbeat(ctx context.Context, registry, addr string, duration time.Duration) {
	if duration == 0 {
		// make sure there is enough time to send heart beat
		// before it's removed from registry
		duration = defaultRegistryTimeout - time.Duration(1)*time.Minute
	}
	var err error
	buf := bytes.NewReader([]byte(c.localNode.getState()))
	err = sendHeartbeat(registry, addr, buf)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			select {
			case <-t.C:
				err = sendHeartbeat(registry, addr, buf)
			case <-ctx.Done():
				// xlog.Error(ctx.Err())
				return
			}
		}
	}()
}

func sendHeartbeat(registry, addr string, body io.Reader) error {
	// xlog.Infoln(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("HEARTBEAT", registry, body)
	req.Header.Set("X-Grasure-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		xlog.Fatalln("rpc server: heart beat err:", err)
		return err
	}
	return nil
}
