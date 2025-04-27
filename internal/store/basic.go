package store

import (
	"fmt"
	"sync"
	"time"
)

type DataType int

const (
	StringType DataType = iota
	ListType
)

// Item represents a stored item with expiration
type Item struct {
	Type      DataType
	Value     interface{}
	ExpiresAt time.Time
}

type DataObj struct {
	Mu     *sync.RWMutex
	Data   *DataMap
	Timer  *time.Ticker
	StopCh chan bool
}

// IsExpired checks if an item is expired
func (i *Item) IsExpired() bool {
	if i.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(i.ExpiresAt)
}

type DataMap struct {
	Data map[string]*Item
}

// NewDataMap creates a new data map instance
func NewDataMap() *DataMap {
	return &DataMap{
		Data: make(map[string]*Item),
	}
}

func NewRedisMemoryStore() *DataObj {
	s := &DataObj{
		Mu:     &sync.RWMutex{},
		Data:   NewDataMap(),
		StopCh: make(chan bool),
	}

	s.Timer = time.NewTicker(time.Second * 30)
	go s.runCleanUp()

	return s
}

func (s *DataObj) runCleanUp() {
	for {
		select {
		case <-s.Timer.C:
			fmt.Println("cleaning expired items")
			s.cleanExpired()
		case <-s.StopCh:
			s.Timer.Stop()
			return
		}
	}
}

func (s *DataObj) cleanExpired() {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	now := time.Now()
	for k, v := range s.Data.Data {
		if !v.ExpiresAt.IsZero() && now.After(v.ExpiresAt) {
			fmt.Println("deleting the key", k, v)
			delete(s.Data.Data, k)
		}
	}
}

func (s *DataObj) Stop() {
	s.StopCh <- true
}
