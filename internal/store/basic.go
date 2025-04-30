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
	Mu     sync.RWMutex
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
			// Make a copy of the key string to prevent mutation
			keyToDelete := string([]byte(k))
			fmt.Printf("Key expired: '%s' (expires: %v, now: %v)\n", keyToDelete, v.ExpiresAt, now)
			keysToDelete = append(keysToDelete, keyToDelete)
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

func (s *DataObj) Set(key string, item string, ttl *time.Duration) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	var expiresAt time.Time
	if ttl != nil && *ttl > 0 {
		expiresAt = time.Now().Add(*ttl)
	}

	fmt.Println(expiresAt)

	s.Data.Data[key] = &Item{
		Type:      StringType,
		Value:     item,
		ExpiresAt: expiresAt,
	}
	fmt.Println(s.Data.Data[key])
	return nil
}

func (s *DataObj) Get(key string) (interface{}, DataType, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	item, found := s.Data.Data[key]
	if !found {
		return nil, 0, false
	}

	// if item.IsExpired() {
	// 	delete(s.Data.Data, key)
	// 	return nil, 0, false
	// }
	return item.Value, item.Type, true
}

func (s *DataObj) Remove(key string) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	_, exists := s.Data.Data[key]
	if exists {
		delete(s.Data.Data, key)
		return true
	}
	return false
}

func (s *DataObj) Update(key string, value string) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	item, exists := s.Data.Data[key]
	if !exists || item.IsExpired() {
		return false
	}

	if item.Type != StringType {
		return false
	}

	item.Value = value
	return true
}

func (s *DataObj) GetTTL(key string) (time.Duration, bool) {
	fmt.Println("Current keys in map:")
	for k := range s.Data.Data {
		fmt.Printf("  '%s' (bytes: %x)\n", k, []byte(k))
	}
	// Create a defensive copy of the key
	keyCopy := defensiveCopy(key)

	fmt.Printf("GetTTL called with key: '%s' (bytes: %x)\n", keyCopy, []byte(keyCopy))

	s.Mu.Lock()
	defer s.Mu.Unlock()

	// Debug: print all keys
	fmt.Println("Current keys in map:")
	for k := range s.Data.Data {
		fmt.Printf("  '%s' (bytes: %x)\n", k, []byte(k))
	}

	item, exists := s.Data.Data[keyCopy]
	if !exists {
		fmt.Printf("Key not found for TTL: '%s'\n", keyCopy)
		return 0, false
	}

	// if item.IsExpired() {
	// 	fmt.Printf("Key expired: '%s'\n", keyCopy)
	// 	delete(s.Data.Data, keyCopy)
	// 	return 0, false
	// }

	if item.ExpiresAt.IsZero() {
		fmt.Printf("Key has no expiration: '%s'\n", keyCopy)
		return -1, true // -1 indicates no expiration
	}

	remaining := time.Until(item.ExpiresAt)
	// if remaining < 0 {
	// 	fmt.Printf("Key has negative remaining time: '%s'\n", keyCopy)
	// 	delete(s.Data.Data, keyCopy)
	// 	return 0, false
	// }

	fmt.Printf("Successfully retrieved TTL for key '%s': %v\n", keyCopy, remaining)
	return remaining, true
}

func (s *DataObj) SetTTL(key string, ttl time.Duration) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	item, exists := s.Data.Data[key]
	if !exists || item.IsExpired() {
		return false
	}

	if ttl <= 0 {
		item.ExpiresAt = time.Time{} // No expiration
	} else {
		item.ExpiresAt = time.Now().Add(ttl)
	}

	return true
}

func (s *DataObj) GetList(key string) ([]string, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	item, exists := s.Data.Data[key]
	if !exists || item.IsExpired() {
		return nil, fmt.Errorf("item not found or expired for key")
	}

	if item.Type != ListType {
		return nil, fmt.Errorf("item type mismatch")
	}

	list, ok := item.Value.([]string)
	if !ok {
		return nil, fmt.Errorf("item type mismatch")
	}

	// Return a copy to prevent external modifications
	result := make([]string, len(list))
	copy(result, list)

	return result, nil
}

func (s *DataObj) CreateList(key string, ttl time.Duration) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if _, exists := s.Data.Data[key]; exists {
		return false
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	s.Data.Data[key] = &Item{
		Type:      ListType,
		Value:     []string{},
		ExpiresAt: expiresAt,
	}

	return true
}

// Push adds a value to the end of a list
func (s *DataObj) Push(key string, value string) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	item, exists := s.Data.Data[key]
	if !exists || item.IsExpired() {
		// Create the list if it doesn't exist
		s.Data.Data[key] = &Item{
			Type:      ListType,
			Value:     []string{value},
			ExpiresAt: time.Time{}, // No expiration by default
		}
		return true
	}

	if item.Type != ListType {
		return false
	}

	list, ok := item.Value.([]string)
	if !ok {
		return false
	}

	item.Value = append(list, value)
	return true
}

// Pop removes and returns the last value from a list
func (s *DataObj) Pop(key string) (string, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	item, exists := s.Data.Data[key]
	if !exists || item.IsExpired() {
		return "", false
	}

	if item.Type != ListType {
		return "", false
	}

	list, ok := item.Value.([]string)
	if !ok || len(list) == 0 {
		return "", false
	}

	lastIndex := len(list) - 1
	value := list[lastIndex]
	item.Value = list[:lastIndex]

	return value, true
}

func defensiveCopy(key string) string {
	return string([]byte(key))
}
