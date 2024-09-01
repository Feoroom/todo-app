package main

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	_ "library/cmd/api/docs"
	"library/internal/config"
	"library/internal/data"
	"library/internal/logger"
	"os"
	"time"
)

// @title Library API
// @version 1.0
// @description Simple Rest API
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email soberkoder@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8000
// @BasePath /

const version = "1.0"

type application struct {
	config config.Config
	logger *logger.Logger
	models data.Models
}

func main() {

	var cfg config.Config
	cfg.SetEnvironment()

	lgr := logger.New(os.Stdout, logger.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		lgr.PrintFatal(err, nil)
	}
	defer db.Close()

	lgr.PrintInfo("database established", nil)

	app := &application{
		config: cfg,
		logger: lgr,
		models: data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		lgr.PrintFatal(err, nil)
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
