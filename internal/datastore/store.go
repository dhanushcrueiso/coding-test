package datastore

import (
	"fmt"
	"time"

	"github.com/dhanushcrueiso/coding-test/internal/store"
)

type repo struct {
	//mu    *sync.RWMutex
	Items *store.DataObj
}

func NewStore(items *store.DataObj) Repository {
	return &repo{
		//mu:    &sync.RWMutex{},
		Items: items,
	}
}

func (r *repo) Set(key string, item string, ttl *time.Duration) error {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()
	var expiresAt time.Time
	if ttl != nil && *ttl > 0 {
		expiresAt = time.Now().Add(*ttl)
	}

	fmt.Println(expiresAt)

	r.Items.Data.Data[key] = &store.Item{
		Type:      store.StringType,
		Value:     item,
		ExpiresAt: expiresAt,
	}
	fmt.Println(r.Items.Data.Data[key])
	return nil
}

func (r *repo) Get(key string) (interface{}, store.DataType, bool) {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	item, found := r.Items.Data.Data[key]
	if !found {
		return nil, 0, false
	}

	if item.IsExpired() {
		delete(r.Items.Data.Data, key)
		return nil, 0, false
	}
	return item.Value, item.Type, true
}

func (r *repo) Remove(key string) bool {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	_, exists := r.Items.Data.Data[key]
	if exists {
		delete(r.Items.Data.Data, key)
		return true
	}
	return false
}

func (r *repo) Update(key string, value string) bool {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	item, exists := r.Items.Data.Data[key]
	if !exists || item.IsExpired() {
		return false
	}

	if item.Type != store.StringType {
		return false
	}

	item.Value = value
	return true
}

func (r *repo) GetTTL(key string) (time.Duration, bool) {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	item, exists := r.Items.Data.Data[key]
	if !exists || item.IsExpired() {
		return 0, false
	}

	if item.ExpiresAt.IsZero() {
		return -1, true // -1 indicates no expiration
	}

	remaining := time.Until(item.ExpiresAt)
	if remaining < 0 {
		// Remove synchronously instead of in a goroutine
		delete(r.Items.Data.Data, key)
		return 0, false
	}

	return remaining, true
}

func (r *repo) SetTTL(key string, ttl time.Duration) bool {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	item, exists := r.Items.Data.Data[key]
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

func (r *repo) GetList(key string) ([]string, error) {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	item, exists := r.Items.Data.Data[key]
	if !exists || item.IsExpired() {
		return nil, fmt.Errorf("item not found or expired for key")
	}

	if item.Type != store.ListType {
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

func (r *repo) CreateList(key string, ttl time.Duration) bool {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	if _, exists := r.Items.Data.Data[key]; exists {
		return false
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	r.Items.Data.Data[key] = &store.Item{
		Type:      store.ListType,
		Value:     []string{},
		ExpiresAt: expiresAt,
	}

	return true
}

// Push adds a value to the end of a list
func (r *repo) Push(key string, value string) bool {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	item, exists := r.Items.Data.Data[key]
	if !exists || item.IsExpired() {
		// Create the list if it doesn't exist
		r.Items.Data.Data[key] = &store.Item{
			Type:      store.ListType,
			Value:     []string{value},
			ExpiresAt: time.Time{}, // No expiration by default
		}
		return true
	}

	if item.Type != store.ListType {
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
func (r *repo) Pop(key string) (string, bool) {
	r.Items.Mu.Lock()
	defer r.Items.Mu.Unlock()

	item, exists := r.Items.Data.Data[key]
	if !exists || item.IsExpired() {
		return "", false
	}

	if item.Type != store.ListType {
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
