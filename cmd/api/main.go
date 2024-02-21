package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	"github.com/Light2Dark/greenlight/internal/data"
	"github.com/Light2Dark/greenlight/internal/jsonlog"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // need pq for establishing conn to db, but we alias to avoid go lint
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string // data source name
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	err := godotenv.Load(".env")
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "dev", "Environment (staging|dev|prod)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL data source name (dsn)")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL maximum open connections (in-use & idle)")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL maximum idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL maximum idle connection liftetime")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
		os.Exit(1)
	}
	defer db.Close()
	logger.PrintInfo("db connection pool succesfully established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	os.Exit(1)
}

// returns a connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn) // opens a connection pool
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// establish conn. with db, which will throw error if not established in 5s
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
