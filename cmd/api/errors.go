package main

import (
	"fmt"
	"net/http"
)

// logError method is helper for logging error message in *application,
// also requested method and request URL.
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

// errorResponse method is helper for sending JSON error messages to the client with a given status code.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	wrap := wrapper{"error": message}

	err := app.writeJSON(w, status, wrap, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// serverErrorResponse method is used for unexpected problem at runtime.
// it logs error message, then uses errorResponse() helper to send
// 500 Internal Server Error status code and JSON response to client.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the application encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// notFoundResponse method is used to send 404 Not Found status code and JSON response to client.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// methodNotAllowedResponse method is used to send 405 Method Not Allowed status code and JSON response to client.
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// badRequestResponse sends JSON error message with 400 Bad Request status code.
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// failedValidationResponse sends JSON error message to client
// with Unprocessable Entity 422 status code when validation fails.
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// editConflictResponse sends JSON error message to client with 409 Conflict status code.
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}
