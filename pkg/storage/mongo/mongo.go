package mongo

import (
	"context"
	"news/pkg/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Когда при выполнении операции не найдено
// ни одного документа
var ErrNoDocuments = mongo.ErrNoDocuments

// псевдоним для объекта хранения БД
type item = storage.Item

// Mongo структура для выполнения CRUD операций с БД
type Mongo struct {
	client *mongo.Client // клиент mongo
	// название текущей db,
	// переключается методом Database()
	database string
	// название текущей collection,
	// переключается методом Collection()
	collection string
}

// New подключается к БД, используя connstr, и возвращает
// объект для работы с БД
func New(connstr, database, collection string) (*Mongo, error) {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connstr))
	if err != nil {
		return nil, err
	}

	return &Mongo{
		client:     client,
		database:   database,
		collection: collection,
	}, client.Ping(context.Background(), nil)
}

// Database переключает имя базы данных mongodb
// в структуре *Mongo
func (m *Mongo) Database(database string) *Mongo {
	m.database = database
	return m
}

// Collection переключает имя коллекции в структуре *Mongo.
func (m *Mongo) Collection(collection string) *Mongo {
	m.collection = collection
	return m
}

// Close закрывает соединение с БД
func (m *Mongo) Close() error {
	return m.client.Disconnect(context.Background())
}

// AddItems добавляет в БД слайс rss-новостей,
// ингорирует те новости, что уже есть в БД
func (m *Mongo) AddItems(ctx context.Context, items []item) error {

	col := m.client.Database(m.database).Collection(m.collection)

	models := make([]mongo.WriteModel, len(items))

	for i := range items {
		filter := bson.D{bson.E{Key: "link", Value: items[i].Link}}
		// bson.MarshalValue()
		update := bson.D{bson.E{Key: "$setOnInsert", Value: items[i]}}

		models[i] = mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := col.BulkWrite(ctx, models, opts)

	return err
}

// Items возвращает списком по крайней мере n rss-новостей
// отсортированных по дате публикации по убыванию
func (m *Mongo) Items(ctx context.Context, n int) ([]item, error) {

	col := m.client.Database(m.database).Collection(m.collection)

	opts := options.Find().SetSort(bson.D{bson.E{Key: "pubDate", Value: -1}}).SetLimit(int64(n))
	cursor, err := col.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	var items []item

	return items, cursor.All(ctx, &items)
}

// AddItem добавляет в БД rss-новость, если новость уже
// есть в БД, то no-op
func (m *Mongo) AddItem(ctx context.Context, item item) error {

	col := m.client.Database(m.database).Collection(m.collection)
	filter := bson.D{bson.E{Key: "link", Value: item.Link}}
	opts := options.Update().SetUpsert(true)
	upd := bson.D{
		bson.E{
			Key: "$setOnInsert", Value: item},
	}

	_, err := col.UpdateOne(ctx, filter, upd, opts)

	return err
}

// Item находит по ссылке и возвращает rss-новость.
// Возвращает ошибку ErrNoDocuments в случае если документ не найден
func (m *Mongo) Item(ctx context.Context, link string) (item, error) {

	col := m.client.Database(m.database).Collection(m.collection)

	var item item

	return item, col.FindOne(ctx, bson.D{bson.E{Key: "link", Value: link}}).Decode(&item)
}

// DeleteItem удаляет из БД rss-новость
func (m *Mongo) DeleteItem(ctx context.Context, link string) error {
	col := m.client.Database(m.database).Collection(m.collection)
	_, err := col.DeleteOne(ctx, bson.D{bson.E{Key: "link", Value: link}})
	return err
}

// UpdateItem обновляет в БД rss-новость
func (m *Mongo) UpdateItem(ctx context.Context, item item) error {

	col := m.client.Database(m.database).Collection(m.collection)

	filter := bson.D{bson.E{Key: "link", Value: item.Link}}
	upd := bson.D{
		bson.E{
			Key: "$set", Value: item},
	}
	_, err := col.UpdateOne(ctx, filter, upd)

	return err
}
