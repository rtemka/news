package memdb

import (
	"context"
	"news/pkg/storage"
)

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

func (db *MemDB) Item(_ context.Context, _ string) (storage.Item, error) {
	return sampleItem, nil
}

func (db *MemDB) Items(_ context.Context, _ int) ([]storage.Item, error) {
	return []storage.Item{sampleItem, sampleItem}, nil
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
