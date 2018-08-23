package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type KvStore struct {
	sync.RWMutex
	KvStore map[string]string
}

func (kv *KvStore) Apply(log *raft.Log) interface{} {
	var op Op
	if err := json.Unmarshal(log.Data, &op); err != nil {
		panic(err)
	}
	return kv.apply(op)
}

func (kv *KvStore) Snapshot() (raft.FSMSnapshot, error) {
	snapshot := *kv
	return &snapshot, nil
}

func (kv *KvStore) Restore(inp io.ReadCloser) error {
	bSizeBuf := make([]byte, 2)
	if _, err := inp.Read(bSizeBuf); err != nil {
		panic(err)
	}
	bSize := int(binary.LittleEndian.Uint16(bSizeBuf))
	buf := make([]byte, bSize)
	if _, err := inp.Read(buf); err != nil && err != io.EOF {
		panic("snapshot decode error:" + err.Error())
	}
	fmt.Println("buf:", string(buf))
	kvs := make(map[string]string)
	if err := json.Unmarshal(buf, &kvs); err != nil {
		panic(err)
	}
	for k, v := range kvs {
		kv.KvStore[k] = v
	}
	return nil
}

func (kv *KvStore) Persist(sink raft.SnapshotSink) error {
	data, err := json.Marshal(kv.KvStore)
	if err != nil {
		return err
	}
	bSize := uint16(len(data))
	buf := make([]byte, bSize+2)
	binary.LittleEndian.PutUint16(buf[:2], bSize)
	copy(buf[2:], data)
	if _, err = sink.Write(buf); err != nil {
		return err
	}
	return nil
}

func (kv *KvStore) Release() {
}

func NewKVStore() *KvStore {
	return &KvStore{
		KvStore: make(map[string]string),
	}
}

func (s *KvStore) Get(key string) string {
	s.RLock()
	defer s.RUnlock()
	return s.KvStore[key]
}

func (s *KvStore) set(key, value string) {
	s.Lock()
	defer s.Unlock()
	s.KvStore[key] = value
}

func (s *KvStore) del(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.KvStore, key)
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
