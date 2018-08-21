package main

import (
	"flag"
	"github.com/forjoin92/depot/raftnode"
	"github.com/forjoin92/depot/service"
	"sync"
)

func main() {
	cluster := flag.String("cluster", "127.0.0.1:30403", "Comma separated cluster peers")
	id := flag.String("id", "127.0.0.1:30403", "Node id")
	//join := flag.Bool("join", false, "join a cluster")
	snapshotPath := flag.String("snapshotPath", "", "raft snapshot path")
	raftDBPath := flag.String("raftDBPath", "", "raft raftDB path")
	testAddr := flag.String("testAddr", "127.0.0.1", "test addr")
	testPort := flag.String("testPort", "8090", "test port")
	flag.Parse()

	node, err := raftnode.NewRaftNode(*id, *cluster, *snapshotPath, *raftDBPath)
	if err != nil {
		panic(err)
	}

	httpServer := service.NewHTTPServer(node, *testAddr, *testPort, false, false)

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		httpServer.Serve()
	}()
	wg.Wait()
}
