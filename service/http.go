package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/forjoin92/depot/raftnode"
	"github.com/julienschmidt/httprouter"
)

type HTTPServer struct {
	node        *raftnode.RaftNode
	tlsEnabled  bool
	tlsRequired bool
	router      http.Handler
	addr        string
	port        string
	listener    net.Listener
}

func NewHTTPServer(node *raftnode.RaftNode, addr, port string, tlsEnabled bool, tlsRequired bool) (*HTTPServer, error) {
	// check node is empty or not
	if node == nil {
		return nil, fmt.Errorf("Node should not be empty")
	}

	// set default http addr and port
	if addr == "" {
		addr = "127.0.0.1"
	}
	if port == "" {
		port = "90" + strings.Split(string(node.ID()), ":")[1][3:]
	}

	router := httprouter.New()

	s := &HTTPServer{
		node:        node,
		tlsEnabled:  tlsEnabled,
		tlsRequired: tlsRequired,
		router:      router,
		addr:        addr,
		port:        port,
	}

	router.GET("/getKV/:key", s.getKV)
	router.PUT("/setKV", s.setKV)
	router.DELETE("/deleteKV/:key", s.deleteKV)
	router.POST("/addNode", s.addNode)
	router.DELETE("/removeNode", s.removeNode)

	return s, nil
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
	fmt.Printf("%s: listening on %s:%s\n", "http", s.addr, s.port)

	server := &http.Server{
		Handler: s,
	}
	err = server.Serve(s.listener)
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		fmt.Println(err)
	}

	fmt.Printf("%s: closing %s:%s\n", "http", s.addr, s.port)
}

// 获取keyvalue
func (s *HTTPServer) getKV(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	value := s.node.GetKV(ps.ByName("key"))
	if value == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", value)
}

// 设置keyvalue
func (s *HTTPServer) setKV(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	kvs := make(map[string]string)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&kvs); err != nil {
		log.Printf("Failed to read on PUT (%v)\n", err)
		http.Error(w, "Failed on PUT", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	for k, v := range kvs {
		if err := s.node.SetKV(k, v); err != nil {
			log.Printf("Failed to set (%v)\n", err)
			http.Error(w, "Failed on PUT", http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// 删除keyvalue
func (s *HTTPServer) deleteKV(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if err := s.node.DeleteKV(key); err != nil {
		log.Printf("Failed to delete (%v)\n", err)
		http.Error(w, "Failed on POST", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// 增加raft集群节点
func (s *HTTPServer) addNode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read on POST (%v)\n", err)
		http.Error(w, "Failed on POST", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if s.node.IsLeader() {
		// 接收节点为leader，直接增加节点
		err = s.node.AddNode(string(id))
	} else {
		// 接收点不是leader，转发到leader节点
		leader := s.node.Leader()
		raftAddr := strings.Split(string(leader), ":")
		raftPort, err := strconv.Atoi(raftAddr[1])
		if err != nil {
			log.Printf("Failed to get raft api port (%v)\n", err)
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}
		apiPort := 9000 + raftPort%100
		url := fmt.Sprintf("http://%s:%d/addNode", raftAddr[0], apiPort)
		log.Println("转发ip:", url)
		_, err = http.Post(url, "application/json", bytes.NewReader(id))
	}
	if err != nil {
		log.Printf("Failed to add node (%v)\n", err)
		http.Error(w, "Failed on POST", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// 移除raft集群节点
func (s *HTTPServer) removeNode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read on POST (%v)\n", err)
		http.Error(w, "Failed on POST", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := s.node.RemoveNode(string(id)); err != nil {
		log.Printf("Failed to remove node (%v)\n", err)
		http.Error(w, "Failed on POST", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
