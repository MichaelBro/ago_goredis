package app

import (
	"ago_goredis/cmd/service/app/dto"
	"ago_goredis/pkg/news"
	"encoding/json"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type Server struct {
	newsSvc *news.Service
	router  chi.Router
}

func NewServer(newsSvc *news.Service, router chi.Router) *Server {
	return &Server{newsSvc: newsSvc, router: router}
}

func (s *Server) Init() error {
	s.router.Get("/api/news/latest", s.latestNews)
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
