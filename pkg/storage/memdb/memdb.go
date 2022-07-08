package memdb

import (
	"context"
	"news/pkg/storage"
)

// MemDB - заглушка настоящей БД
type MemDB struct{}

func New() *MemDB {
	return &MemDB{}
}

// SampleItem можно использовать для тестов
var SampleItem = storage.Item{
	Id:          1,
	Oid:         [12]byte{},
	Title:       "sample item",
	PubDate:     5555555,
	Description: "sample discription",
	Link:        "https://test.com",
}

// Item возвращает один экземпляр SampleItem
func (db *MemDB) Item(_ context.Context, _ string) (storage.Item, error) {
	return SampleItem, nil
}

// Items возвращает столько Item, сколько запрошено
func (db *MemDB) Items(_ context.Context, n int) ([]storage.Item, error) {
	items := make([]storage.Item, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, SampleItem)
	}
	return items, nil
}

// AddItem - no-op
func (db *MemDB) AddItem(_ context.Context, _ storage.Item) error {
	return nil
}

// AddItems - no-op
func (db *MemDB) AddItems(_ context.Context, _ []storage.Item) error {
	return nil
}

// DeleteItem - no-op
func (db *MemDB) DeleteItem(_ context.Context, _ storage.Item) error {
	return nil
}

// UpdateItem - no-op
func (db *MemDB) UpdateItem(_ context.Context, _ storage.Item) error {
	return nil
}

// Close - no-op
func (db *MemDB) Close() error {
	return nil
}
