package app

import (
	"ago_goredis/cmd/service/app/dto"
	cacheMiddleware "ago_goredis/cmd/service/app/middlewares/cache"
	"ago_goredis/pkg/cache"
	"ago_goredis/pkg/news"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
)

type Server struct {
	newsSvc  *news.Service
	cacheSvc *cache.Service
	router   chi.Router
}

func NewServer(newsSvc *news.Service, router chi.Router, cacheSvc *cache.Service) *Server {
	return &Server{newsSvc: newsSvc, router: router, cacheSvc: cacheSvc}
}

func (s *Server) Init() error {

	cacheMd := cacheMiddleware.Cache(
		func(ctx context.Context, path string) ([]byte, error) {
			value, err := s.cacheSvc.Get(ctx, path)
			if err != nil && errors.Is(err, redis.ErrNil) {
				return nil, cacheMiddleware.ErrNotInCache
			}
			return value, err
		},

		func(ctx context.Context, path string, data []byte) error {
			return s.cacheSvc.Set(ctx, path, data)
		},

		func(writer http.ResponseWriter, data []byte) error {
			writer.Header().Set("Content-Type", "application/json")
			_, err := writer.Write(data)
			if err != nil {
				log.Println(err)
			}
			return err
		})

	s.router.With(cacheMd).Get("/api/news/latest", s.latestNews)
	s.router.Post("/api/news", s.createNews)

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s *Server) latestNews(writer http.ResponseWriter, request *http.Request) {
	items, err := s.newsSvc.GetLatest(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := make([]*dto.NewsDTO, len(items))
	for i, n := range items {
		data[i] = dto.FromModel(n)
	}

	writeJson(writer, data, http.StatusOK)
}

func (s *Server) createNews(writer http.ResponseWriter, request *http.Request) {
	itemToSave := dto.NewsDTO{}
	err := json.NewDecoder(request.Body).Decode(&itemToSave)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.newsSvc.Save(request.Context(), itemToSave.Title, itemToSave.Text)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusCreated)

	go func() {
		if err = s.cacheSvc.DeleteAllCache(context.Background()); err != nil {
			log.Println(err)
		}
	}()
}

func writeJson(w http.ResponseWriter, data interface{}, code int) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		http.Error(w, "response marshaling failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(body)
	if err != nil {
		log.Println(err)
	}
}
