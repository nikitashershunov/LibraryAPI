package main

import (
	"net/http"
)

// healthcheckHandler is a handler for checking if server is running succefully
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Declare wrapper map containing the data for response.
	env := wrapper{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
