package main

import (
	"net/http"
)

// @summary Info handler
// @description Info Handler
// @router /v1/info [get]
func (app *application) info(w http.ResponseWriter, r *http.Request) {

	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"env":     app.config.Env,
			"version": version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
