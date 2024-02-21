package main

import (
	"net/http"
	"strconv"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	var dataEnvelope = envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"port":        strconv.Itoa(app.config.port),
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, dataEnvelope, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
