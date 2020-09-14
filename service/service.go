package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/georgiypetrov/auto-assignment/models"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type ConfigService struct {
	Port string `env:"SERVICE_PORT" envDefault:"9000"`
}

type Service struct {
	DB     *models.DB
	Server http.Server
	Log    *logrus.Logger
}

type SaveURLRequestBody struct {
	LongURL   string `json:"long_url"`
	CustomURL string `json:"custom_url"`
}

type appError struct {
	Error error
	Code  int
}

type appHandler struct {
	fn  func(http.ResponseWriter, *http.Request) *appError
	log *logrus.Logger
}

func ParseConfig() (*ConfigService, error) {
	cfg := &ConfigService{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func InitService(db *models.DB, log *logrus.Logger) (*Service, error) {
	cfg, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	s := &Service{
		DB:  db,
		Log: log,
	}
	s.Server = http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: CreateRouter(s),
	}
	return s, nil
}

func (s *Service) Serve() {
	defer s.DB.Close()
	port := strings.Split(s.Server.Addr, ":")[1]
	s.Log.Info(fmt.Sprintf("Start server on port %s", port))
	s.Log.Fatal(s.Server.ListenAndServe())
}

func (s *Service) Shutdown(ctx context.Context) error {
	err := s.DB.Close()
	if err != nil {
		return err
	}
	err = s.Server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn.fn(w, r); e != nil {
		fn.log.WithError(e.Error).WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.EscapedPath(),
		}).Error("")
		http.Error(w, e.Error.Error(), e.Code)
	}
}

func CreateRouter(s *Service) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/set", appHandler{s.saveURL, s.Log}.ServeHTTP).Methods(http.MethodPost)
	r.PathPrefix("/").HandlerFunc(appHandler{s.redirectByShortURL, s.Log}.ServeHTTP).Methods(http.MethodGet)
	return r
}

func (s *Service) saveURL(w http.ResponseWriter, r *http.Request) *appError {
	defer r.Body.Close()

	data := SaveURLRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return &appError{err, http.StatusInternalServerError}
	}

	shortURL, err := s.DB.SaveURL(data.LongURL, data.CustomURL)
	if err != nil {
		return &appError{err, http.StatusInternalServerError}
	}

	if err := json.NewEncoder(w).Encode(shortURL); err != nil {
		return &appError{err, http.StatusInternalServerError}
	}
	return nil
}

func (s *Service) redirectByShortURL(w http.ResponseWriter, r *http.Request) *appError {
	shortURL := r.URL.Path[1:]
	longURL, err := s.DB.GetLongURL(shortURL)
	if errors.Is(err, models.ErrShortUrlNotExist) {
		return &appError{err, http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, http.StatusInternalServerError}
	}
	http.Redirect(w, r, longURL, http.StatusPermanentRedirect)
	return nil
}
