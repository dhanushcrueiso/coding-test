package datastore

import (
	"time"

	"github.com/dhanushcrueiso/coding-test/internal/store"
)

type Repository interface {
	Set(key string, item string, time *time.Duration) error
	Get(key string) (interface{}, store.DataType, bool)
	Update(key string, value string) bool
	Remove(key string) bool
	GetTTL(key string) (time.Duration, bool)
	SetTTL(key string, ttl time.Duration) bool
	GetList(key string) ([]string, error)
	CreateList(key string, ttl time.Duration) bool
	Push(key string, value string) bool
	Pop(key string) (string, bool)
}
