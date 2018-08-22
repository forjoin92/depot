package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type KvStore struct {
	sync.RWMutex
	kvStore map[string]string
}

func NewKVStore() *KvStore {
	return &KvStore{
		kvStore: make(map[string]string),
	}
}

func (s *KvStore) Get(key string) string {
	s.RLock()
	defer s.RUnlock()
	return s.kvStore[key]
}

func (s *KvStore) set(key, value string) {
	s.Lock()
	defer s.Unlock()
	s.kvStore[key] = value
}

func (s *KvStore) del(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.kvStore, key)
}

func (s *KvStore) apply(op Op) interface{} {
	switch op.Method {
	case "SET":
		s.set(op.Key, op.Value)
	case "DEL":
		s.del(op.Key)
	default:
		fmt.Printf("unknown op:%s\n", op.Method)
		return errors.New(fmt.Sprintf("unknown op:%s", op.Method))
	}
	return nil
}

type Op struct {
	Method string
	Key    string
	Value  string
}

func (kv *KvStore) Apply(log *raft.Log) interface{} {
	var op Op
	if err := json.Unmarshal(log.Data, &op); err != nil {
		panic(err)
	}

	return kv.apply(op)
}

func (kv *KvStore) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (s *KvStore) Restore(rc io.ReadCloser) error {
	return nil
}
