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
	fmt.Printf("Starting cleanup at %v, checking %d keys\n", now, len(s.Data.Data))

	// Create a separate list of keys to delete to avoid map modification during iteration
	keysToDelete := []string{}

	for k, v := range s.Data.Data {
		if !v.ExpiresAt.IsZero() && now.After(v.ExpiresAt) {
			fmt.Printf("Key expired: '%s' (expires: %v, now: %v)\n", k, v.ExpiresAt, now)
			keysToDelete = append(keysToDelete, k)
		}
	}

	// Now perform the actual deletion
	for _, k := range keysToDelete {
		fmt.Printf("Deleting key: '%s'\n", k)
		delete(s.Data.Data, k)

		// Verify deletion
		if _, stillExists := s.Data.Data[k]; stillExists {
			fmt.Printf("ERROR: Key '%s' still exists after deletion!\n", k)
		} else {
			fmt.Printf("Successfully deleted key: '%s'\n", k)
		}
	}

	fmt.Printf("Cleanup complete: %d keys deleted, %d keys remaining\n",
		len(keysToDelete), len(s.Data.Data))
}

func (s *DataObj) Stop() {
	s.StopCh <- true
}
