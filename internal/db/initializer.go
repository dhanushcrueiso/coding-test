package db

import (
	"github.com/dhanushcrueiso/coding-test/internal/datastore"
	"github.com/dhanushcrueiso/coding-test/internal/store"
)

var DataSvc datastore.Service = nil

func InitDb() {
	storeMap := store.NewRedisMemoryStore()

	storeRepo := datastore.NewStore(storeMap)
	DataSvc = datastore.NewService(storeRepo)

}
