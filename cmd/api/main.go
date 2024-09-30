package main

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"library/internal/config"
	"library/internal/data"
	"library/internal/logger"
	"library/internal/mailer"
	_ "library/internal/metrics"
	"log/slog"
	"sync"
	"time"
)

type Application struct {
	config config.Config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {

	var cfg config.Config
	cfg.SetEnvironment()

	lgr := logger.SetSlogLogger(cfg.Env)

	db, err := openDB(cfg)
	if err != nil {
		lgr.Error(err.Error())
		return
	}
	defer db.Close()

	lgr.Info("database established")

	app := &Application{
		config: cfg,
		logger: lgr,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.STMP.Host, cfg.STMP.Port, cfg.STMP.Username, cfg.STMP.Password, cfg.STMP.Sender),
	}

	err = app.Serve()
	if err != nil {
		lgr.Error(err.Error())
		return
	}

}

func openDB(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.DB.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
