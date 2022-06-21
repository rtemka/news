package rsscollector

import (
	"context"
	"encoding/xml"
	"fmt"
	"mime"
	"net/http"
	"news/pkg/storage"
	"sync"
	"time"
)

// container - это наша зависимость, тип
// с которым работает пакет rsscollector
type container = storage.ItemContainer
type item = storage.Item

// Poll опрашивает переданные rss-ссылки c заданным интервалом времени.
func Poll(ctx context.Context, interval time.Duration, links []string) (<-chan any, <-chan error, error) {

	if len(links) == 0 {
		return nil, nil, fmt.Errorf("poll: no links provided")
	}

	dests := make([]<-chan any, len(links))  // по каналу для каждой rss-ссылки
	errs := make([]<-chan error, len(links)) // каналы ошибок для каждой горутины

	// Fan-Out мультиплексируем каналы
	for i := 0; i < len(links); i++ {

		ch := make(chan any)
		dests[i] = ch
		ech := make(chan error)
		errs[i] = ech

		go func(values chan<- any, errors chan<- error, url string) {

			defer func() {
				close(values)
				close(errors)
			}()

			poll := pollFunc(ctx, url) // функция опроса по ссылке

			for {

				select {

				case <-time.After(interval):
					v, err := poll(&container{}) // выполняем опрос
					if err != nil {
						errors <- err
					} else {
						values <- v
					}

				case <-ctx.Done():
					errors <- ctx.Err()
					return
				}
			}
		}(ch, ech, links[i])
	}

	return merge(dests...), merge(errs...), nil // Fan-In (демультиплексируем каналы)
}

// merge демультиплексирует(собирает) переданные
// каналы в единый канал и возвращает этот канал
func merge[T any](sources ...<-chan T) <-chan T {

	dest := make(chan T) // единый канал

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
type poller func(container any) (any, error)

// pollFunc создает функцию для опроса rss-канала по переданному URL
func pollFunc(ctx context.Context, url string) poller {

	rf := requestFunc(ctx, url) // функция для выполнения запроса по сети

	return func(container any) (any, error) {

		// логика декодирования тела ответа
		decfunc := responseHandlerFunc(func(r *http.Response) error {
			return xml.NewDecoder(r.Body).Decode(&container)
		})

		chain := bodyCloser(statusChecker(decfunc)) // цепочка обработчиков ответа

		return container, rf(chain)
	}
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
