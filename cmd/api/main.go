package main

import (
	"flag"
	"log/slog"
	"os"

	"fetch.trungnng.github.io/internal/data"
)

// Application version number
const version = "1.0.0"

// Configuration settings for application.
type config struct {
	port    int
	env     string
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

// Hold the dependencies for HTTP handlers, helpers, middleware
type application struct {
	config config
	logger *slog.Logger
	model  *data.Models
}

func main() {
	var cfg config

	// Default port number 4000 and the environment "development"
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	// Create new structured logger to standard out
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Declare an instance of the application struct, containing the config struct and
	// the logger.
	app := &application{
		config: cfg,
		logger: logger,
		model:  data.NewModels(),
	}

	// Call app.serve() to start the server.
	err := app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
