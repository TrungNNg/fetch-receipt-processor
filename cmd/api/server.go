package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// serve starts the HTTP server and handles graceful shutdowns upon receiving
// termination signals (SIGINT, SIGTERM).
//
// Returns:
//   - `nil` if the server starts and shuts down successfully.
//   - Any error encountered during ListenAndServe() or Shutdown() if they occur.
func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Use this to receive any errors returned by the graceful Shutdown() function.
	shutdownError := make(chan error)

	// Listening for termination signals (SIGINT, SIGTERM).
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit // block until detect a signal

		app.logger.Info("shutting down server", "signal", s.String())

		// Create a context with a 30-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Relay this return value to the shutdownError channel.
		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// ListenAndServe() will immediately return a http.ErrServerClosed error if
	// Shutdown is called. So we only need to report error that not http.ErrServerClosed.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// If Shutdown failed it will return an error.
	err = <-shutdownError
	if err != nil {
		return err
	}

	// Shutdown completed successfully
	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
