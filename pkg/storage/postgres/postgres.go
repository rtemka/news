package postgres

import (
	"context"
	"news/pkg/storage"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrNoRows = pgx.ErrNoRows

// Postgres выполняет CRUD операции с БД
type Postgres struct {
	db *pgxpool.Pool
}

// New выполняет подключение
// и возвращает объект для взаимодействия с БД
func New(connString string) (*Postgres, error) {

	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	return &Postgres{db: pool}, pool.Ping(context.Background())
}

// Close выполняет закрытие подключения к БД
func (p *Postgres) Close() {
	p.db.Close()
}

// Item находит по ссылке и возвращает rss-новость
func (p *Postgres) Item(ctx context.Context, link string) (storage.Item, error) {
	stmt := `
		SELECT
			n.id,
			n.title,
			n.description,
			n.pub_date,
			n.link
		FROM news as n
		WHERE n.link = $1;`

	var item storage.Item

	err := p.db.QueryRow(ctx, stmt, link).Scan(
		&item.Id, &item.Title, &item.Description,
		&item.PubDate, &item.Link)
	if err != nil {
		return item, err
	}

	return item, nil
}

// Items возвращает списком по крайней мере n rss-новостей
// отсортированных по дате публикации по убыванию
func (p *Postgres) Items(ctx context.Context, n int) ([]storage.Item, error) {
	stmt := `
		SELECT
			n.id,
			n.title,
			n.description,
			n.pub_date,
			n.link
		FROM news as n
		ORDER BY n.pub_date DESC
		LIMIT $1;`

	var items []storage.Item

	rows, err := p.db.Query(ctx, stmt, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var item storage.Item

		err := rows.Scan(&item.Id, &item.Title,
			&item.Description, &item.PubDate, &item.Link)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

// AddItems добавляет в БД слайс rss-новостей
func (p *Postgres) AddItems(ctx context.Context, items []storage.Item) error {
	return p.addItemsByBatch(ctx, items)
}

// addItemsByBatch вносит в БД слайс rss-новостей,
// используя *pgx.Batch
func (p *Postgres) addItemsByBatch(ctx context.Context, items []storage.Item) error {

	return p.db.BeginFunc(ctx, func(tx pgx.Tx) error {

		b := new(pgx.Batch) // создаем объект pgx.Batch

		stmt := `
		INSERT INTO news(title, description, pub_date, link)
		VALUES ($1, $2, $3, $4);`

		// добавляем все запросы в очередь
		for i := range items {
			b.Queue(stmt, items[i].Title, items[i].Description,
				items[i].PubDate, items[i].Link)
		}

		return tx.SendBatch(ctx, b).Close() // исполняем запросы и закрываем операцию

	})
}

// addItemsByCopy вносит в БД слайс rss-новостей,
// используя Posrgresql copy protocol
func (p *Postgres) addItemsByCopy(ctx context.Context, items []storage.Item) error {

	return p.db.BeginFunc(ctx, func(tx pgx.Tx) error {

		cf := pgx.CopyFromSlice(len(items), func(i int) ([]interface{}, error) {
			return []any{items[i].Title, items[i].Description, items[i].PubDate, items[i].Link}, nil
		}) // // функция копирования из слайса

		table := pgx.Identifier{"news"}                                       // имя таблицы
		columns := pgx.Identifier{"title", "description", "pub_date", "link"} // имена атрибутов

		_, err := tx.CopyFrom(ctx, table, columns, cf) // вносим данные в БД с помощью postgres COPY FROM
		if err != nil {
			return err
		}
		return nil

	})
}

// AddItem добавляет в БД rss-новость
func (p *Postgres) AddItem(ctx context.Context, item storage.Item) error {
	stmt := `
		INSERT INTO news(title, description, pub_date, link)
		VALUES ($1, $2, $3, $4);`

	return p.exec(ctx, stmt, item.Title, item.Description, item.PubDate, item.Link)
}

// DeleteItem удаляет из БД rss-новость
func (p *Postgres) DeleteItem(ctx context.Context, item storage.Item) error {
	stmt := `
		DELETE FROM news
		WHERE link = $1;`

	return p.exec(ctx, stmt, item.Link)
}

// UpdateItem обновляет в БД rss-новость
func (p *Postgres) UpdateItem(ctx context.Context, item storage.Item) error {
	stmt := `
		UPDATE news
		SET 
			title = $1,
			description = $2,
			pub_date = $3
			WHERE link = $4;`

	return p.exec(ctx, stmt, item.Title,
		item.Description, item.PubDate, item.Link)
}

// exec вспомогательная функция, выполняет
// *pgx.conn.Exec() в транзакции
func (p *Postgres) exec(ctx context.Context, sql string, args ...any) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = p.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
