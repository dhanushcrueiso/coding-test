package db

import (
	"CODING-TEST/internal/datastore"
	"CODING-TEST/internal/store"
)

var DataSvc datastore.Service = nil

func InitDb() {
	storeMap := store.NewRedisMemoryStore()

	storeRepo := datastore.NewStore(storeMap)
	DataSvc = datastore.NewService(storeRepo)

}
