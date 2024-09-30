package main

import (
	"library/internal/metrics"
	"net/http"
)

func (app *Application) info(w http.ResponseWriter, r *http.Request) {

	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"env":     app.config.Env,
			"version": metrics.Version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
