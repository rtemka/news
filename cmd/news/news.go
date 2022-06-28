package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"news/pkg/api"
	"news/pkg/rsscollector"
	"news/pkg/storage/postgres"
	"news/pkg/storage/streamwriter"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// имя подсистемы для логирования
var (
	rsscolName = fmt.Sprintf("%16s", "[RSS Collector] ")
	dwName     = fmt.Sprintf("%16s", "[DB Writer] ")
	apiName    = fmt.Sprintf("%16s", "[WEB API] ")
)

// config - структура для хранения конфигурации
// передаваемой в качестве аргумента коммандной строки
type config struct {
	Links        []string `json:"rss"`            // массив ссылок для опроса
	SurveyPeriod int      `json:"request_period"` // период опроса ссылок в минутах
}

// readConfig функция для чтения файла конфигурации
func readConfig(path string) (*config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c config

	return &c, json.NewDecoder(f).Decode(&c)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <path-to-config-file>\n", os.Args[0])
		os.Exit(1)
	}

	config, err := readConfig(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	connstr := os.Getenv("NEWS_DB_CONN_STRING")
	if connstr == "" {
		fmt.Fprintln(os.Stderr, "$NEWS_DB_CONN_STRING environment variable must be set")
		os.Exit(1)
	}

	db, err := postgres.New(connstr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()

	// логгеры для подсистем
	rsslog := log.New(os.Stdout, rsscolName, log.Lmsgprefix|log.LstdFlags)
	dbwriterlog := log.New(os.Stdout, dwName, log.Lmsgprefix|log.LstdFlags)
	apilog := log.New(os.Stdout, apiName, log.Lmsgprefix|log.LstdFlags)

	collector := rsscollector.New(rsslog).DebugMode(true)               // RSS-обходчик
	sw := streamwriter.NewStreamWriter(dbwriterlog, db).DebugMode(true) // объект пишуший в БД
	webapi := api.New(db, apilog)                                       // REST API

	// конфигурируем сервер
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           webapi.Router(),
		IdleTimeout:       3 * time.Minute,
		ReadHeaderTimeout: time.Minute,
	}

	// создаем контекст для регулирования закрытия всех подсистем
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interval := time.Second * time.Duration(config.SurveyPeriod)
	values, errs, err := collector.Poll(ctx, interval, config.Links)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	// читаем канал с ошибками
	go func() {
		errChecker(cancel, errs)
		wg.Done()
	}()

	// читаем канал с новостями и пишем в БД
	go func() {
		_, err = sw.WriteToStorage(ctx, values)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		wg.Done()
	}()

	// сервер
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		} else {
			log.Println(err) // server closed
		}
		wg.Done()
	}()

	// ловим сигналы прерывания типа CTRL-C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-stop // получили сигнал прерывания
		log.Println("got os signal", s)

		// закрываем сервер
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}

		cancel() // закрываем контекст приложения
	}()

	wg.Wait() // ждём всех
}

// можем что-то делать с ошибками приложения
func errChecker(cancel context.CancelFunc, errs <-chan error) {

	threshold := 100 // например, установить допустимый предел
	current := 0

	for err := range errs {
		// или проверять на конкретную ошибку
		var addrerr *net.AddrError
		if errors.As(err, &addrerr) {
			fmt.Fprintf(os.Stderr, "the net addr error: %v\n", addrerr)
		}
		var operr *net.OpError
		if errors.As(err, &operr) {
			fmt.Fprintf(os.Stderr, "the net op error: %v is temporary: %t\n", operr, operr.Temporary())
		}

		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "%T %v\n", err, err)
			current++
		}
		if current >= threshold {
			fmt.Fprintln(os.Stderr, "errors threshold exceeded")
			cancel()
		}
	}
}
