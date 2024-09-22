package main

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"library/internal/data"
	"library/internal/validation"
	"log"
	"net/http"
)

func (app *application) registerHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validation.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.models.Tokens.New(user.ID, data.TokenDuration, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {

		emailData := map[string]interface{}{
			"activationToken": token.PlainText,
			"userId":          user.ID,
		}

		err = app.mailer.Send(user.Email, "welcome.gohtml", emailData)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) sendTokenHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// TODO: body must not be empty - сделать чтобы выводилось сообщение вместо ошибки
	var input struct {
		Email string
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validation.New()

	data.ValidateEmail(v, input.Email)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if user.Activated {
		v.AddError("email", "user has been activated")
		return
	}

	token, err := app.models.Tokens.New(user.ID, data.TokenDuration, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		emailData := map[string]interface{}{
			"activationToken": token.PlainText,
		}
		err = app.mailer.Send(user.Email, "token.gohtml", emailData)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"message": "activation instructions sent on your email"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) activateUserHandle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var input struct {
		Token string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validation.New()

	if data.ValidateTokenPlainText(v, input.Token); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.Token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	log.Print(user)

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createAuthenticationToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var input struct {
		Email    string
		Password string
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validation.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePlaintextPassword(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.models.Tokens.New(user.ID, data.TokenDuration, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
