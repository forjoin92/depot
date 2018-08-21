package service

import (
	"net/http"
	"github.com/forjoin92/depot/raftnode"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"encoding/json"
	"io"
	"net"
	"strings"
)

type HTTPServer struct {
	node        *raftnode.RaftNode
	tlsEnabled  bool
	tlsRequired bool
	router      http.Handler
	addr string
	port string
	listener net.Listener
}

func NewHTTPServer(node *raftnode.RaftNode, addr, port string, tlsEnabled bool, tlsRequired bool) *HTTPServer {
	router := httprouter.New()

	s := &HTTPServer{
		node:         node,
		tlsEnabled:  tlsEnabled,
		tlsRequired: tlsRequired,
		router:      router,
	}

	router.GET("/get/:key", s.GetValue)
	router.PUT("/set", s.Set)
	router.DELETE("/delete/:key", s.Delete)

	return s
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !s.tlsEnabled && s.tlsRequired {
		resp := fmt.Sprintf(`{"message": "TLS_REQUIRED", "https_port": %s}`, s.port)
		w.Header().Set("X-NSQ-Content-Type", "nsq; version=1.0")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(403)
		io.WriteString(w, resp)
		return
	}
	s.router.ServeHTTP(w, req)
}

func (s *HTTPServer) Serve() {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", s.addr, s.port))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: listening on %s", "http", s.addr)

	server := &http.Server{
		Handler:  s,
	}
	err = server.Serve(s.listener)
	// theres no direct way to detect this error because it is not exposed
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		fmt.Println(err)
	}

	fmt.Printf("%s: closing %s", "http", s.addr)
}

func (s *HTTPServer) GetValue(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "%s", s.node.Get(ps.ByName("key")))
}

func (s *HTTPServer) Set(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	kvs := make(map[string]string)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&kvs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	for k, v := range kvs {
		s.node.Set(k, v)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if err := s.node.Del(key); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
