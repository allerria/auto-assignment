package models

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"net/url"
	"strings"
)

const (
	ShortUrlLength        = 6
	ShortUrlGenTriesLimit = 3
)

var (
	ErrCantGenerateShortURL = errors.New("can't generate short url")
	ErrShortURLExist        = errors.New("short url is already exist")
	ErrShortUrlNotExist     = errors.New("short url doesn't exist")
)

type DB struct {
	*sqlx.DB
}

// ConfigDB contains database connection information. Also used by "env" library for parsing this struct from
// environment variables.
type ConfigDB struct {
	User     string `env:"POSTGRES_USER,required"`
	Pass     string `env:"POSTGRES_PASS,required"`
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	Database string `env:"POSTGRES_DATABASE,required"`
	SslMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
}

func ParseConfig() (*ConfigDB, error) {
	cfg := &ConfigDB{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func InitDB() (*DB, error) {
	cfg, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Database, cfg.SslMode)
	sqlxDB, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, err
	}
	db := &DB{DB: sqlxDB}
	return db, nil
}

func getFullURL(URL string) (string, error) {
	if !strings.Contains(URL, "http") {
		URL = fmt.Sprintf("http://%s", URL)
	}
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return "", fmt.Errorf("can't parse url: %v", err)
	}
	return URL, nil
}

func (db *DB) SaveURL(longURL string, customURL string) (string, error) {
	longURL, err := getFullURL(longURL)
	if err != nil {
		return "", fmt.Errorf("invalid long url: %v", err)
	}
	if customURL == "" {
		return db.SaveGeneratedURL(longURL)
	}
	return db.SaveCustomURL(longURL, customURL)
}

func (db *DB) SaveGeneratedURL(longURL string) (string, error) {
	var shortURL string
	urlGenerated := false
	for i := 0; i < ShortUrlGenTriesLimit; i++ {
		shortURL = GenerateURL(ShortUrlLength)
		if exist, err := db.CheckShortURLExist(shortURL); err != nil {
			return "", err
		} else if exist {
			continue
		}
		urlGenerated = true
		break
	}
	if !urlGenerated {
		return "", ErrCantGenerateShortURL
	}
	return db.SaveShortURL(longURL, shortURL)
}

func (db *DB) SaveCustomURL(longURL string, customURL string) (string, error) {
	if exist, err := db.CheckShortURLExist(customURL); err != nil {
		return "", err
	} else if exist {
		return "", ErrShortURLExist
	}
	return db.SaveShortURL(longURL, customURL)
}

func (db *DB) SaveShortURL(longURL string, shortURL string) (string, error) {
	_, err := db.Exec("INSERT INTO urls (short_url, long_url) VALUES ($1, $2)", shortURL, longURL)
	if err != nil {
		return "", fmt.Errorf("error save short url: %v", err)
	}
	return shortURL, nil
}

func (db *DB) CheckShortURLExist(shortURL string) (bool, error) {
	urls := []string{}
	err := db.Select(&urls, "SELECT short_url FROM urls WHERE short_url = $1", shortURL)
	if err != nil {
		return false, fmt.Errorf("error checking url exist: %v", err)
	}
	if len(urls) == 0 {
		return false, nil
	}
	return urls[0] == shortURL, nil
}

func (db *DB) GetLongURL(shortURL string) (string, error) {
	if exist, err := db.CheckShortURLExist(shortURL); err != nil {
		return "", err
	} else if !exist {
		return "", ErrShortUrlNotExist
	}
	longUrl := ""
	err := db.Get(&longUrl, "SELECT long_url FROM urls WHERE short_url = $1", shortURL)
	if err != nil {
		return "", fmt.Errorf("error select long url: %v", err)
	}
	return longUrl, nil
}
