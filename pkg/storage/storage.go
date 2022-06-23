package storage

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
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

// для конвертирования из RFC1123Z, RFC1123...
// 'Mon, 02 Jan 2006 15:04:05 -0700' и подобных
// в unix timestamp
type unix int64

var layouts = []string{time.RFC1123Z, time.RFC1123,
	time.UnixDate, "02 Jan 2006 15:04:05 -0700", time.ANSIC,
	time.RFC850, time.RFC822, time.RFC822Z}

func (t *unix) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	var pt time.Time
	var err error

	for i := range layouts {
		pt, err = time.Parse(layouts[i], s)
		if err == nil {
			break
		}
	}
	*t = unix(pt.Unix())

	return err
}

type StreamWriter struct {
	log     *log.Logger
	storage Storage
}

func NewStreamWriter(log *log.Logger, storage Storage) *StreamWriter {
	return &StreamWriter{
		log:     log,
		storage: storage,
	}
}

func (sw *StreamWriter) WriteToStorage(ctx context.Context, in <-chan any) error {
	var c, i uint // обработанные контейнеры, обработанные новости
	var u uint    // неопознанные значения пришедшие в канал

	for v := range in {

		if container, ok := v.(*ItemContainer); ok {

			dbctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			err := sw.storage.AddItems(dbctx, container.Items)
			if err != nil {
				return err
			}

			c++
			i += uint(len(container.Items))

		} else {
			u++
		}

	}

	sw.log.Printf("totals: containers=%d, items=%d, unrecognized=%d", c, i, u)

	return nil
}
