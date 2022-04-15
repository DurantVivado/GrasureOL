package grasure

import (
	"encoding/json"
	"fmt"
	"github.com/DurantVivado/GrasureOL/codec"
	"github.com/DurantVivado/GrasureOL/xlog"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x7fffffff

type Option struct {
	MagicNumber int        // MagicNumber marks this's a geerpc request
	CodecType   codec.Type // client may choose different Codec to encode body
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}


//Server represents an RPC server.
type Server struct{
	
}

// request stores all information of a call
type request struct {
	h            *codec.Header // header of request
	argv, replyv reflect.Value // argv and replyv of request
}


func NewServer() *Server{
	return &Server{}
}

var DefaultServer = NewServer()

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func(s *Server) Accept(lis net.Listener){
	for{
		conn, err := lis.Accept()
		if err != nil{
			xlog.Errorln("rpc server: accept error:", err)
			return
		}
		go s.ServeConn(conn)
	}
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.

func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		xlog.Errorln("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		xlog.Errorln("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		xlog.Errorln("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	s.serveCodec(f(conn))
}

//invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct {}{}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
func (s *Server) serveCodec( cc codec.Codec){
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for{
		req, err := s.readRequest(cc)
		if err != nil{
			if req == nil{
				//if it turns out impossible to recover, then close the connection
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg)
	}
	_ = cc.Close()
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			xlog.Errorln("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		xlog.Errorln("rpc server: read argv err:", err)
	}
	return req, nil
}

func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	// TODO, should call registered rpc methods to get the right replyv
	// day 1, just print argv and send a hello message
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d: %s", req.h.Seq, req.h.ServiceMethod))
	s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}

func(s *Server) Start(network, addr string){
	if addr == ":0"{
		xlog.Errorln("Please don't set addr as :0, in this way an arbitrary port is picked")
		return
	}
	l, err := net.Listen(network, addr)
	if err != nil{
		xlog.Fatalln("server network error:", err)
	}
	xlog.Infoln("start rpc server on :", l.Addr())
	s.Accept(l)
}

