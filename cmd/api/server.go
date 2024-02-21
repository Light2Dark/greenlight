package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     log.New(os.Stdout, "", 0),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // will listen for incoming syscalls and relay to quit channel
		// notify doesn't wait for receiver to be ready, that's why we use buffered channel

		s := <-quit // blocks until signal is received
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx) // call shutdown on server and relay errors to shutdownError channel
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"address": srv.Addr,
		"env":     app.config.env,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) { // we see if server has not closed gracefully
		return err
	}

	err = <-shutdownError // w8 to receive message from shutdown chan
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server sucessfully", map[string]string{
		"address": srv.Addr,
	})
	return nil
}
