package mock

import (
	"sync"
	"time"
)

// KVStore mocks github.KVStore.
type KVStore struct {
	data        map[string][]byte
	reads       int
	updates     int
	m           sync.Mutex
	writeTokens chan struct{}
}

// NewKVStore creates new KVStore instance with given data
func NewKVStore(data map[string][]byte, writeTokens chan struct{}) *KVStore {
	return &KVStore{
		data:        data,
		writeTokens: writeTokens,
	}
}

// ReadKey returns data saved for given key.
func (s *KVStore) ReadKey(key []byte) ([]byte, error) {
	s.m.Lock()
	defer s.m.Unlock()

	s.reads++
	if s.data == nil {
		return nil, nil
	}

	return s.data[string(key)], nil
}

// UpdateKey stores given data under given key.
func (s *KVStore) UpdateKey(key []byte, data []byte) error {
	if s.writeTokens != nil {
		select {
		case <-s.writeTokens:
		case <-time.After(time.Second):
			panic("kvstore locked")
		}
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.updates++
	if s.data == nil {
		s.data = make(map[string][]byte)
	}
	s.data[string(key)] = data

	return nil
}

// Reads returns read call count.
func (s *KVStore) Reads() int {
	s.m.Lock()
	defer s.m.Unlock()

	return s.reads
}

// Updates returns update call count.
func (s *KVStore) Updates() int {
	s.m.Lock()
	defer s.m.Unlock()

	return s.updates
}
