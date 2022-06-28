package memdb

import (
	"context"
	"news/pkg/storage"
)

type MemDB struct{}

func New() *MemDB {
	return &MemDB{}
}

var SampleItem = storage.Item{
	Id:          1,
	PubDate:     5555555,
	Title:       "sample item",
	Description: "sample discription",
	Link:        "https://test.com",
}

func (db *MemDB) Item(_ context.Context, _ string) (storage.Item, error) {
	return SampleItem, nil
}

func (db *MemDB) Items(_ context.Context, n int) ([]storage.Item, error) {
	items := make([]storage.Item, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, SampleItem)
	}
	return items, nil
}

func (db *MemDB) AddItem(_ context.Context, _ storage.Item) error {
	return nil
}

func (db *MemDB) AddItems(_ context.Context, _ []storage.Item) error {
	return nil
}

func (db *MemDB) DeleteItem(_ context.Context, _ storage.Item) error {
	return nil
}

func (db *MemDB) UpdateItem(_ context.Context, _ storage.Item) error {
	return nil
}

func (db *MemDB) Close() error {
	return nil
}
