package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/nikitashershunov/LibraryAPI/internal/data"
	"github.com/nikitashershunov/LibraryAPI/internal/jsonlog"
)

// define config struct.
type config struct {
	port int
	env  string
	// db struct field holds configuration settings for database connection pool.
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// define application struct to hold dependencies for HTTP handlers, helpers.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

const (
	version = "1.0.0"
)

func main() {
	var cfg config

	// Read DSN value from db-dsn command-line flag in config struct.
	// Default is my development DSN if no flag is provided.
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("BOOKS_DB_DSN"), "PostgreSQL DSN")

	// Read connection pool settings from command-line flags in config struct.
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")

	// Read value of port and env command-line flags in config struct.
	// Default port number 4000 and environment "development".
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|production)")

	flag.Parse()

	// Initialize new jsonlog.Logger that writes any messages above INFO level to standard output stream.
	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	// Call openDB() function (below) to create connection pool.
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer call to db.Close().
	defer func() {
		if err := db.Close(); err != nil {
			logger.PrintFatal(err, nil)
		}
	}()

	logger.PrintInfo("database connection pool established", nil)

	// Declare an instance of the application struct.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// Call app.serve() to start the server.
	if err := app.serve(); err != nil {
		logger.PrintFatal(err, nil)
	}
}

// openDB returns a sql.DB connection pool to postgres database.
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open connections in the pool.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connection in the pool.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
