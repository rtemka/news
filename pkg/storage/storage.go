package storage

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"
)

// Storage - контракт на работу с БД
type Storage interface {
	Item(ctx context.Context, link string) (Item, error) // Получить новость по ссылке
	Items(ctx context.Context, n int) ([]Item, error)    // Получить все новости списком
	AddItem(context.Context, Item) error                 // Добавить новость
	AddItems(context.Context, []Item) error              // Добавить новости списком
	DeleteItem(context.Context, Item) error              // Удалить новость
	UpdateItem(context.Context, Item) error              // Обновить новость
	Close() error                                        // закрыть БД
}

// Item - модель данных rss-новости
type Item struct {
	Id          int64  `json:"id" bson:"_id"`
	Title       string `json:"title" bson:"title"`
	PubDate     int64  `json:"pubDate" bson:"pubDate"`
	Description string `json:"description" bson:"description"`
	Link        string `json:"link" bson:"link"`
}

func (i Item) String() string {
	return fmt.Sprintf("Id: %d, Title: %s, Description: %s, Link: %s",
		i.Id, i.Title, i.Description, i.Link)
}

// ItemContainer - контейнер содержащий rss-новости.
// Используется для декодирования xml и транспортировки по каналу
type ItemContainer struct {
	Items []Item `xml:"channel>item"`
}

// xmlItem - копия Item, единственная польза
// от которой декодирование xml для Item.
// Боремся с проблемой конвертирования времени
// при десериализации
type xmlItem struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	PubDate     unix     `xml:"pubDate"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
}

func (xi *xmlItem) toItem() Item {
	return Item{
		Id:          0,
		Title:       xi.Title,
		PubDate:     int64(xi.PubDate),
		Description: xi.Description,
		Link:        xi.Link,
	}
}

func (i *Item) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var xi xmlItem
	err := d.DecodeElement(&xi, &start)
	if err != nil {
		return err
	}
	*i = xi.toItem()
	return nil
}

// для конвертирования из RFC1123Z
// 'Mon, 02 Jan 2006 15:04:05 -0700' в unix timestamp
type unix int64

func (t *unix) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	pt, err := time.Parse(time.RFC1123Z, s)
	if err != nil {
		return err
	}
	*t = unix(pt.Unix())

	return nil
}
