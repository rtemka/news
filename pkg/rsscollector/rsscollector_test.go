package rsscollector

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

const xmlblob = `
		<rss>
			<channel>
				<item>
					<title>Тестовый заголовок</title>
					<link>https://test.com</link>
					<description>Тестовое описание</description>
					<pubDate>Thu, 16 Jun 2022 10:14:28 +0300</pubDate>
				</item>
			</channel>
		</rss>
		`

func Test_poller(t *testing.T) {

	want := item{
		Id:          0,
		Title:       "Тестовый заголовок",
		PubDate:     1655363668,
		Description: "Тестовое описание",
		Link:        "https://test.com",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprintln(w, xmlblob)
	}))
	defer ts.Close()

	poll := pollFunc(context.Background(), ts.URL)

	c, err := poll(&container{})
	if err != nil {
		t.Fatalf("poller() error = %v", err)
	}

	if got, ok := c.(*container); ok {
		if len(got.Items) != 1 {
			t.Fatalf("poller() got result length = %d, want = %d", len(got.Items), 1)
		}

		if got.Items[0] != want {
			t.Fatalf("poller() got = %v, want = %v", got, want)
		}
	} else {
		t.Fatalf("poller() unexpected return type = %T", c)
	}
}

func TestPoll(t *testing.T) {

	var count int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprintln(w, xmlblob)
	}))
	defer ts.Close()

	t.Run("oшибка_ноль_ссылок", func(t *testing.T) {
		_, _, err := Poll(context.Background(), time.Second, nil)
		if err == nil {
			t.Fatal("Poll() expected error, got nothing")
		}
	})

	t.Run("подсчёт_количества_опросов", func(t *testing.T) {

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		values, errs, err := Poll(ctx, time.Second, []string{ts.URL})
		if err != nil {
			t.Fatalf("Poll() error = %v", err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		gotv := 0

		go func() {
			defer wg.Done()

			for range values {
				gotv++
			}
		}()

		go func() {
			defer wg.Done()
			for range errs {
			}
		}()

		wg.Wait()

		if gotv != count {
			t.Fatalf("Poll() got values = %d, want values = %d", gotv, count)
		}
	})
}
