package models

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/jmoiron/sqlx"
	"time"
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

type URL struct {
	ShortURL  string    `json:"short_url" db:"short_url"`
	LongURL   string    `json:"long_url" db:"short_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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
	db := &DB{sqlxDB}
	return db, nil
}
