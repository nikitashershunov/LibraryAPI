package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// declare an HTTP server using the same settings as in main() function.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(app.logger, "", 0),
	}

	// background goroutine.
	go func() {
		// create quit channel which carries os.Signal values.
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// read signal from the quit channel. This code will block until signal is received.
		s := <-quit

		// log a message to say that the signal has been caught.
		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		// exit with a 0 success status code.
		os.Exit(0)
	}()

	// log starting server message.
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	return srv.ListenAndServe()
}
