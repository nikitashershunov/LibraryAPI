package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes is router for my main application.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// convert the app.notFoundResponse helper to http.Handler using the http.HandlerFunc()
	// and set it as custom error handler for 404 Not Found.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// convert the app.methodNotAllowedResponse helper to http.Handler
	// and set it as custom error handler for 405 Method Not Allowed.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// healthcheck handler and corresponding endpoint
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// books handlers and corresponding endpoints
	router.HandlerFunc(http.MethodGet, "/v1/books", app.listBooksHandler)
	router.HandlerFunc(http.MethodPost, "/v1/books", app.createBookHandler)
	router.HandlerFunc(http.MethodGet, "/v1/books/:id", app.getBookHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/books/:id", app.updateBookHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/books/:id", app.deleteBookHandler)

	return app.recoverPanic(app.rateLimit(router))
}
