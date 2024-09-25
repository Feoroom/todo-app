package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	Port int
	Env  string
	DB   struct {
		DSN          string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdleTime  string
	}
	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}
	STMP struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
	CORS struct {
		AllowedOrigins []string
	}
}

func (cfg *Config) SetEnvironment() {
	flag.IntVar(&cfg.Port, "port", 8000, "server Port")
	flag.StringVar(&cfg.Env, "env", "dev", "Environment(dev|stag|prod)")

	flag.StringVar(&cfg.DB.DSN, "db-dsn", os.Getenv("TODO_DB_DSN"), "PostgreSQL DSN")

	flag.IntVar(&cfg.DB.MaxOpenConns, "db-max-open-conns", 25, "PostgeSQL max open connection")
	flag.IntVar(&cfg.DB.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.DB.MaxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.Limiter.Rps, "limiter-rps", 4, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", 8, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.STMP.Host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.STMP.Port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.STMP.Username, "smtp-username", "4e6131523774d9", "SMTP username")
	flag.StringVar(&cfg.STMP.Password, "smtp-password", "e656e78aaac168", "STMP password")
	flag.StringVar(&cfg.STMP.Sender, "smtp-sender", "Todo <no-reply@todo.goserv.ru>", "SMTP sender")

	flag.Func("cors-allowed-origins", "Comma-separated list of allowed CORS origins", func(s string) error {
		cfg.CORS.AllowedOrigins = strings.Fields(s)
		return nil
	})

	flag.Parse()
}
