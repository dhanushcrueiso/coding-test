package datastore

import (
	"time"

	"github.com/dhanushcrueiso/coding-test/internal/store"
)

type Service interface {
	Set(key string, item string, ttl *time.Duration) error
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

type dataSvc struct {
	repo Repository
}

func (d *dataSvc) Set(key string, item string, ttl *time.Duration) error {
	return d.repo.Set(key, item, ttl)
}

func (d *dataSvc) Get(key string) (interface{}, store.DataType, bool) {
	return d.repo.Get(key)
}

func (d *dataSvc) Update(key string, value string) bool {
	return d.repo.Update(key, value)
}

func (d *dataSvc) Remove(key string) bool {
	return d.repo.Remove(key)
}

func (d *dataSvc) GetTTL(key string) (time.Duration, bool) {
	return d.repo.GetTTL(key)
}
func (d *dataSvc) SetTTL(key string, ttl time.Duration) bool {
	return d.repo.SetTTL(key, ttl)
}
func (d *dataSvc) GetList(key string) ([]string, error) {
	return d.repo.GetList(key)
}
func (d *dataSvc) CreateList(key string, ttl time.Duration) bool {
	return d.repo.CreateList(key, ttl)
}

func (d *dataSvc) Push(key string, value string) bool {
	return d.repo.Push(key, value)
}

func (d *dataSvc) Pop(key string) (string, bool) {
	return d.repo.Pop(key)
}

func NewService(r Repository) Service {
	return &dataSvc{
		repo: r,
	}
}
