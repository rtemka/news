package mongo

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var tdb *Mongo // тестовая БД

const (
	testDbName       = "newstest"
	testDbCollection = "news"
)

var testItem1 = item{
	Oid:         primitive.NewObjectID(),
	Title:       "test title1",
	PubDate:     1656510227,
	Description: "test descrip1",
	Link:        "https://test1.com",
}

var testItem2 = item{
	Oid:         primitive.NewObjectID(),
	Title:       "test title2",
	PubDate:     1656510228,
	Description: "test descrip2",
	Link:        "https://test2.com",
}

var testData = []any{testItem1, testItem2}

// восстанавилвает состояние тестовой БД
func restoreTestDB(db *Mongo) error {
	err := db.client.Database(db.database).Drop(context.Background())
	if err != nil {
		return err
	}
	col := db.client.Database(db.database).Collection(db.collection)
	_, err = col.InsertMany(context.Background(), testData)
	return err
}

func TestMain(m *testing.M) {

	connstr, ok := os.LookupEnv("MONGO_TEST_DB_URL")
	if !ok {
		fmt.Fprintln(os.Stderr, "environment variable MONGO_TEST_DB_URL must be set")
		os.Exit(1)
	}
	var err error
	tdb, err = New(connstr, testDbName, testDbCollection)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer func() {
		_ = tdb.Close()
	}()

	if err := restoreTestDB(tdb); err != nil {
		_ = tdb.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestMongo(t *testing.T) {

	t.Run("DeleteItem()", func(t *testing.T) {

		want := testItem2

		err := tdb.DeleteItem(context.Background(), want.Link)
		if err != nil {
			t.Fatalf("Mongo.DeleteItem() error = %v", err)
		}

		got, err := tdb.Item(context.Background(), want.Link)
		if err != nil && err != ErrNoDocuments {
			t.Fatalf("Mongo.Item() error = %v", err)
		}

		if got != (item{}) {
			t.Errorf("Mongo.Items() got = %v, want = nothing", got)
		}
	})

	t.Run("AddItem()", func(t *testing.T) {

		want := testItem2

		err := tdb.AddItem(context.Background(), want)
		if err != nil {
			t.Fatalf("Mongo.AddItem() error = %v", err)
		}

		got, err := tdb.Item(context.Background(), want.Link)
		if err != nil {
			t.Fatalf("Mongo.Item() error = %v", err)
		}

		if got != want {
			t.Errorf("Mongo.AddItem() got = %v, want = %v", got, want)
		}

		// пробуем добавить еще раз, запись не должна измениться
		// потому что тот же самый link
		want.Title = "updated title"
		err = tdb.AddItem(context.Background(), want)
		if err != nil {
			t.Fatalf("Mongo.AddItem() error = %v", err)
		}

		got, err = tdb.Item(context.Background(), want.Link)
		if err != nil {
			t.Fatalf("Mongo.Item() error = %v", err)
		}

		if got == want {
			t.Errorf("Mongo.AddItem() got = %v, want = %v", got, got)
		}
	})

	t.Run("AddItems()", func(t *testing.T) {
		want := []item{
			{
				Id:          0,
				Oid:         primitive.NewObjectID(),
				Title:       "new title 1",
				PubDate:     253417963066, // ставим очень далекий год
				Description: "new desc 1",
				Link:        "https://testnewitem1.com",
			},
			{
				Id:          0,
				Oid:         primitive.NewObjectID(),
				Title:       "new title 2",
				PubDate:     253417963065,
				Description: "new desc 2",
				Link:        "https://testnewitem2.com",
			},
		}

		err := tdb.AddItems(context.Background(), want)
		if err != nil {
			t.Fatalf("Mongo.AddItems() error = %v", err)
		}

		got, err := tdb.Items(context.Background(), len(want))
		if err != nil {
			t.Fatalf("Mongo.Items() error = %v", err)
		}

		if len(got) != len(want) {
			t.Fatalf("Mongo.AddItems() got items = %d, want = %d", len(got), len(want))
		}

		for i := range want {
			if got[i] != want[i] {
				t.Errorf("Mongo.AddItems() got = %v, want = %v", got[i], want[i])
			}
		}
	})

	t.Run("UpdateItem()", func(t *testing.T) {
		want := testItem1
		want.Title = "upd title"
		want.Description = "upd desc"

		err := tdb.UpdateItem(context.Background(), want)
		if err != nil {
			t.Fatalf("Mongo.UpdateItem() error = %v", err)
		}

		got, err := tdb.Item(context.Background(), want.Link)
		if err != nil {
			t.Fatalf("Mongo.Item() error = %v", err)
		}

		if got != want {
			t.Errorf("Mongo.UpdateItem() got = %v, want = %v", got, want)
		}
	})
}
