package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"library/internal/data"
	"library/internal/validation"
	"net/http"
)

func (app *application) showCardHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, err := app.readID(params)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	card, err := app.models.Cards.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"card": card}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createCardHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var input struct {
		Title string `json:"title"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	c := &data.Card{
		Title:  input.Title,
		Events: data.Events{},
	}

	v := validation.New()

	if data.ValidateCard(v, c); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Cards.Insert(c)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/card/%d", c.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"card": c}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) updateCardHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	id, err := app.readID(params)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	card, err := app.models.Cards.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title *string `json:"title"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		card.Title = *input.Title
	}

	v := validation.New()
	if data.ValidateCard(v, card); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Cards.Update(card)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"card": card}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
