package grpc

import (
	context "context"
	"errors"
	"fmt"
	"log"
	"news/pkg/storage"

	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

var internalError = fmt.Errorf("internal server error")

type stor = storage.Storage
type storItem = storage.Item

type API struct {
	logger  *log.Logger
	storage stor
	UnimplementedNewsServer
}

func New(storage stor, logger *log.Logger) *API {
	return &API{
		logger:  logger,
		storage: storage,
	}
}

func ofStorageItems(items ...storItem) []*Item {
	out := make([]*Item, 0, len(items))
	for i := range items {
		out = append(out, ofStorageItem(&items[i]))
	}
	return out
}

func (api *API) List(ctx context.Context, in *wrapperspb.Int64Value) (*Items, error) {

	items, err := api.storage.Items(ctx, int(in.Value))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, internalError
	}

	return &Items{Items: ofStorageItems(items...)}, nil
}

func ofStorageItem(si *storItem) *Item {
	return &Item{
		Id:      si.Id,
		Oid:     si.Oid[:],
		Title:   si.Title,
		PubTime: si.PubDate,
		Content: si.Description,
		Link:    si.Link,
	}
}

func toStorageItem(i *Item) *storItem {
	var b [12]byte
	copy(b[:], i.Oid)
	return &storItem{
		Id:          i.Id,
		Oid:         b,
		Title:       i.Title,
		PubDate:     i.PubTime,
		Description: i.Content,
		Link:        i.Link,
	}
}
