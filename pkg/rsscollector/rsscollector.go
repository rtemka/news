package rsscollector

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"news/pkg/storage"
	"sync"
	"time"

	"golang.org/x/net/html/charset"
)

// container - это наша зависимость, тип
// с которым работает пакет rsscollector
type container = storage.ItemContainer
type item = storage.Item

type Collector struct {
	log  *log.Logger
	poll poller
}

func New(log *log.Logger) *Collector {
	return &Collector{
		log:  log,        // логгер
		poll: pollFunc(), // функция опроса по ссылке
	}
}

// Poll опрашивает переданные rss-ссылки c заданным интервалом времени.
func (c *Collector) Poll(ctx context.Context, interval time.Duration, links []string) (<-chan any, <-chan error, error) {

	if len(links) == 0 {
		return nil, nil, fmt.Errorf("collector poll: no links provided")
	}

	dests := make([]<-chan any, len(links))  // по каналу для каждой rss-ссылки
	errs := make([]<-chan error, len(links)) // каналы ошибок для каждой горутины

	// Fan-Out мультиплексируем каналы
	for i := 0; i < len(links); i++ {

		ch := make(chan any)
		dests[i] = ch
		ech := make(chan error)
		errs[i] = ech

		go func(id int, values chan<- any, errors chan<- error, url string) {

			var ec uint // считаем ошибки во время работы горутины
			var pc uint // считаем опросы горутины

			defer func() {
				c.log.Printf("unit #%03d >> totals: polls=%d errors=%d >> task: poll=%s",
					id, pc, ec, url)
				close(values)
				close(errors)
			}()

			for {

				select {

				case <-time.After(interval):

					v, err := c.poll(ctx, &container{}, url) // выполняем опрос

					if err != nil && err != context.DeadlineExceeded {
						ec++
						c.log.Printf("unit #%03d >> error=%v >> task: poll=%s", id, err, url)
						errors <- err
					} else {
						pc++
						values <- v
					}

				case <-ctx.Done():
					errors <- ctx.Err()
					return
				}
			}
		}(i, ch, ech, links[i])
	}

	return merge(dests...), merge(errs...), nil // Fan-In (демультиплексируем каналы)
}

// merge демультиплексирует(собирает) переданные
// каналы в единый канал и возвращает этот канал
func merge[T any](sources ...<-chan T) <-chan T {

	dest := make(chan T, len(sources)) // единый канал

	var wg sync.WaitGroup
	wg.Add(len(sources))

	// читаем все источники и направляем в единый канал
	for _, ch := range sources {
		go func(c <-chan T) {
			defer wg.Done()

			for val := range c {
				dest <- val
			}
		}(ch)
	}

	// ждём закрытия всех источников
	// и закрываем единый канал
	go func() {
		wg.Wait()
		close(dest)
	}()

	return dest
}

// responseHandler - обработчик http-ответа
type responseHandler interface {
	process(*http.Response) error
}

// responseHandlerFunc - адаптер для responseHandler,
// а-ля http.HandlerFunc только для работы с http.Response
type responseHandlerFunc func(*http.Response) error

func (f responseHandlerFunc) process(resp *http.Response) error {
	return f(resp)
}

// poller - функция для опроса rss-канала,
// тело ответа декодируется в переданный контейнер
type poller func(ctx context.Context, container any, url string) (any, error)

// pollFunc создает функцию для опроса rss-канала по переданному URL
func pollFunc() poller {

	return poller(func(ctx context.Context, container any, url string) (any, error) {

		request := requestFunc(ctx, url) // функция для выполнения запроса по сети

		// логика декодирования тела ответа
		decfunc := responseHandlerFunc(func(r *http.Response) error {
			return decoderWithSettings(r.Body).Decode(&container)
		})

		chain := bodyCloser(statusChecker(decfunc)) // цепочка обработчиков ответа

		return container, request(chain)

	})
}

func decoderWithSettings(r io.Reader) *xml.Decoder {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel // некоторые rss-каналы возвращают не UTF-8
	return decoder
}

type requester func(responseHandler) error

// requestFunc возвращает функцию requester, которая будет выполнять
// стандартный http-запрос и передавать ответ обработчику,
// который она принимает в качестве аргумента
func requestFunc(ctx context.Context, url string) requester {

	return func(handler responseHandler) error {
		c, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(c, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		// чтобы rss-каналы не посылали нам ошибку 403
		// ставим заголовок User-Agent
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		return handler.process(resp)
	}
}

func bodyCloser(next responseHandler) responseHandler {
	return responseHandlerFunc(func(resp *http.Response) error {
		defer func() {
			_ = resp.Body.Close()
		}()
		return next.process(resp)
	})
}

func statusChecker(next responseHandler) responseHandler {
	return responseHandlerFunc(func(resp *http.Response) error {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("[%s] response code is %d", resp.Request.URL, resp.StatusCode)
		}
		return next.process(resp)
	})
}

func xmlEnforcer(next responseHandler) responseHandler {
	return responseHandlerFunc(func(resp *http.Response) error {
		ct := resp.Header.Get("Content-Type")
		mediatype, _, err := mime.ParseMediaType(ct)
		if err != nil {
			return err
		}
		if mediatype != "text/xml" {
			return fmt.Errorf("[%s] Content-Type is not 'text/xml'", resp.Request.URL)
		}
		return next.process(resp)
	})
}
