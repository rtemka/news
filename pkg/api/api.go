package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"news/pkg/storage"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type stor = storage.Storage
type item = storage.Item

// API приложения.
type Api struct {
	r         *mux.Router
	db        stor
	logger    *log.Logger
	debugMode bool
}

func New(storage stor, logger *log.Logger) *Api {
	api := Api{
		r:         mux.NewRouter(),
		db:        storage,
		logger:    logger,
		debugMode: false,
	}
	api.endpoints()
	return &api
}

func (api *Api) DebugMode(mode bool) *Api {
	api.debugMode = mode
	return api
}

// Router возвращает маршрутизатор запросов.
func (api *Api) Router() *mux.Router {
	return api.r
}

func (api *Api) endpoints() {
	api.r.Use(api.HeadersMiddleware)
	// получить n последних новостей
	api.r.HandleFunc("/news/{n}", api.itemsHandler).Methods(http.MethodGet, http.MethodOptions)
	// веб-приложение
	api.r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

func (api *Api) HeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// itemsHandler возвращает все новости
func (api *Api) itemsHandler(w http.ResponseWriter, r *http.Request) {

	api.logger.Printf("[DEBUG] method=%s, path=%s, host=%s", r.Method, r.URL.Path, r.Host)

	// Считывание параметра запроса {n} из пути запроса.
	s := mux.Vars(r)["n"]
	limit, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	items, err := api.db.Items(ctx, limit)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
