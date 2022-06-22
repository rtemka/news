package postgres

import (
	"context"
	"fmt"
	"news/pkg/storage"
	"os"
	"path/filepath"
	"testing"
)

var tdb *Postgres // тестовая БД

func restoreTestDB(testdb *Postgres) error {

	b, err := os.ReadFile(filepath.Join("testdata", "testdb.sql"))
	if err != nil {
		return err
	}

	return tdb.exec(context.Background(), string(b))
}

func TestMain(m *testing.M) {

	connstr := os.Getenv("POSTGRES_TEST_DB_URL")
	if connstr == "" {
		fmt.Fprintln(os.Stderr, "environment variable POSTGRES_TEST_DB_URL must be set")
		os.Exit(1)
	}

	var err error
	tdb, err = New(connstr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := restoreTestDB(tdb); err != nil {
		tdb.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer tdb.Close()

	os.Exit(m.Run())
}

func TestPostgres(t *testing.T) {

	t.Run("AddItems()", func(t *testing.T) {
		wantItems := []storage.Item{testItem1, testItem2, testItem3, testItem4}

		err := tdb.AddItems(context.Background(), wantItems)
		if err != nil {
			t.Fatalf("Postgres.AddItems() error = %v", err)
		}

		gotItems, err := tdb.Items(context.Background(), len(wantItems))
		if err != nil {
			t.Fatalf("Postgres.Items() error = %v", err)
		}

		if len(gotItems) != len(wantItems) {
			t.Fatalf("Postgres.Items() got items = %d, want = %d", len(gotItems), len(wantItems))
		}

		for i := range wantItems {
			if gotItems[i] != wantItems[i] {
				t.Fatalf("Postgres.AddItems() got = %v, want = %v", gotItems[i], wantItems[i])
			}
		}
	})

	t.Run("Item()", func(t *testing.T) {
		want := testItem1

		got, err := tdb.Item(context.Background(), want.Link)
		if err != nil {
			t.Fatalf("Postgres.Item() error = %v", err)
		}

		if got != want {
			t.Fatalf("Postgres.Item() got = %v, want = %v", got, want)
		}
	})

	t.Run("Items()", func(t *testing.T) {
		want := testItem1

		limit := 1

		v, err := tdb.Items(context.Background(), limit)
		if err != nil {
			t.Fatalf("Postgres.Items() error = %v", err)
		}

		if len(v) != limit {
			t.Fatalf("Postgres.Items() got items = %d, want = %d", len(v), limit)
		}

		got := v[0]

		if got != want {
			t.Fatalf("Postgres.Items() got = %v, want = %v", got, want)
		}
	})

	t.Run("UpdateItem()", func(t *testing.T) {

		testItem1.Link = testItem4.Link
		testItem1.Id = testItem4.Id

		err := tdb.UpdateItem(context.Background(), testItem1)
		if err != nil {
			t.Fatalf("Postgres.UpdateItem() error = %v", err)
		}

		got, err := tdb.Item(context.Background(), testItem1.Link)
		if err != nil {
			t.Fatalf("Postgres.Item() error = %v", err)
		}

		if got != testItem1 {
			t.Fatalf("Postgres.UpdateItem() got = %v, want = %v", got, testItem1)
		}
	})

	t.Run("DeleteItem()", func(t *testing.T) {

		err := tdb.DeleteItem(context.Background(), testItem1)
		if err != nil {
			t.Fatalf("Postgres.DeleteItem() error = %v", err)
		}

		got, err := tdb.Item(context.Background(), testItem1.Link)
		if err != nil && err != ErrNoRows {
			t.Fatalf("Postgres.Item() error = %v", err)
		}

		if got != (storage.Item{}) {
			t.Fatalf("Postgres.DeleteItem() got = %v, want nothing", got)
		}
	})
}

var testItem1 = storage.Item{
	Id:          1,
	Title:       "Заголовок 1",
	Description: "Описание 1",
	PubDate:     1655806394,
	Link:        "https://test.com/14987527",
}

var testItem2 = storage.Item{
	Id:          2,
	Title:       "Заголовок 2",
	Description: "Описание 2",
	PubDate:     1655806393,
	Link:        "https://test.com/14987528",
}

var testItem3 = storage.Item{
	Id:          3,
	Title:       "Заголовок 3",
	Description: "Описание 3",
	PubDate:     1655806392,
	Link:        "https://test.com/14987529",
}

var testItem4 = storage.Item{
	Id:          4,
	Title:       "Заголовок 4",
	Description: "Описание 4",
	PubDate:     1655806391,
	Link:        "https://test.com/149875210",
}
