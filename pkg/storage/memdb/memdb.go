package memdb

import "news/pkg/storage"

type MemDB struct{}

func New() *MemDB {
	return &MemDB{}
}

var sampleItem = storage.Item{
	Id:          1,
	PubDate:     5555555,
	Title:       "sample item",
	Description: "sample discription",
	Link:        "https://test.com",
}

func (db *MemDB) Item(id int) (storage.Item, error) {
	return sampleItem, nil
}

func (db *MemDB) Items(id int) ([]storage.Item, error) {
	return []storage.Item{sampleItem, sampleItem}, nil
}

func (db *MemDB) AddItem(_ storage.Item) error {
	return nil
}

func (db *MemDB) DeleteItem(_ storage.Item) error {
	return nil
}

func (db *MemDB) UpdateItem(_ storage.Item) error {
	return nil
}

func (db *MemDB) Close() error {
	return nil
}
