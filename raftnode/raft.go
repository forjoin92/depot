package raftnode

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forjoin92/depot/store"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

type RaftNode struct {
	id    string
	peers []string

	snapshotPath string
	raftDBPath   string

	kvs  *store.KvStore
	raft *raft.Raft
}

func NewRaftNode(id string, cluster string, snapshotPath string, raftDBPath string) (*RaftNode, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(id)

	addr, err := net.ResolveTCPAddr("tcp", id)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to resolve TCP address: %s", err.Error()))
	}

	transport, err := raft.NewTCPTransport(id, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create TCP transport: %s", err.Error()))
	}

	snapshot, err := raft.NewFileSnapshotStore(snapshotPath, 3, os.Stderr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create file snapshot store: %s", err.Error()))
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDBPath, "raft.db"))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create log store and stable store: %s", err.Error()))
	}

	kvs := store.NewKVStore()

	r, err := raft.NewRaft(config, kvs, logStore, logStore, snapshot, transport)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create Raft system: %s", err.Error()))
	}

	clusters := strings.Split(cluster, ",")
	servers := make([]raft.Server, len(clusters))
	for i, server := range clusters {
		servers[i] = raft.Server{
			ID:      raft.ServerID(server),
			Address: raft.ServerAddress(server),
			//Address: transport.LocalAddr(),
		}
	}

	configuration := raft.Configuration{
		Servers: servers,
	}

	r.BootstrapCluster(configuration)

	node := &RaftNode{
		id:           id,
		peers:        clusters,
		snapshotPath: snapshotPath,
		raftDBPath:   raftDBPath,
		kvs:          kvs,
		raft:         r,
	}

	return node, nil
}

func (node *RaftNode) Get(key string) string {
	return node.kvs.Get(key)
}

func (node *RaftNode) Set(key, value string) error {
	if node.raft.State() != raft.Leader {
		return errors.New("Not the leader, all changes to the system must go through the leader.")
	}

	op := &store.Op{
		Method: "SET",
		Key:    key,
		Value:  value,
	}

	cmd, err := json.Marshal(op)
	if err != nil {
		return err
	}

	// Apply is used to issue a command to the FSM in a highly consistent manner.
	// This returns a future that ca be used to wait on the application.
	// This must be run on the leader or it will fail.
	return node.raft.Apply(cmd, 10*time.Second).Error()
}

func (node *RaftNode) Del(key string) error {
	if node.raft.State() != raft.Leader {
		return errors.New("Not the leader, all changes to the system must go through the leader.")
	}

	op := &store.Op{
		Method: "DEL",
		Key:    key,
	}

	cmd, err := json.Marshal(op)
	if err != nil {
		return err
	}

	// Apply is used to issue a command to the FSM in a highly consistent manner.
	// This returns a future that ca be used to wait on the application.
	// This must be run on the leader or it will fail.
	return node.raft.Apply(cmd, 10*time.Second).Error()
}
