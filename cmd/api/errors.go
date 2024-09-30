package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (app *Application) logError(r *http.Request, err error) {
	app.logger.Error(err.Error(),
		slog.String("request_method", r.Method),
		slog.String("request_url", r.URL.String()),
	)
}

func (app *Application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *Application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *Application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the request resource could not be found"

	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *Application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)

	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *Application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *Application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *Application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *Application) unauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *Application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

func (app *Application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "you do not have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

func (app *Application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {

	app.errorResponse(w, r, http.StatusUnauthorized, "you have entered an invalid authentication credentials")
}

func (app *Application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	app.errorResponse(w, r, http.StatusUnauthorized, "invalid or missing authentication token")
}
