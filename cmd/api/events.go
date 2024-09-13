package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"library/internal/data"
	"library/internal/validation"
	"net/http"
	"time"
)

func (app *application) createEventHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var input struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		TextBlocks  []string  `json:"text_blocks"`
		Date        data.Date `json:"date"`
		CardId      int64     `json:"card_id"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	e := &data.Event{
		Title:       input.Title,
		Description: input.Description,
		TextBlocks:  input.TextBlocks,
		Date:        input.Date,
		CardId:      input.CardId,
	}

	v := validation.New()

	if data.ValidateEvent(v, e); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Events.Insert(e)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateTitle):
			v.AddError("title", "event with this title already exists")
			app.failedValidationResponse(w, r, v.Errors)

		case errors.Is(err, data.ErrCardConstraint):
			v.AddError("card_id", "card is not present in the table")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//app.logger.Println(e.CreatedAt)

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/events/%d", e.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"event": e}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) showEventHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	id, err := app.readID(params)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	event, err := app.models.Events.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"event": event}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateEventHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	id, err := app.readID(params)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	event, err := app.models.Events.Get(id)
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
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		Date        *data.Date `json:"date"`
		TextBlocks  []string   `json:"text_blocks"`
		Version     *int64     `json:"version"`
		CardId      *int64     `json:"card_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		event.Title = *input.Title
	}

	if input.Description != nil {
		event.Description = *input.Description
	}

	if input.Date != nil {
		event.Date = *input.Date
	}

	if input.Version != nil {
		event.Version = *input.Version
	}

	if input.TextBlocks != nil {
		event.TextBlocks = input.TextBlocks
	}
	app.logger.PrintInfo(fmt.Sprintf("Card ID: %d", input.CardId), nil)

	if input.CardId != nil {
		event.CardId = *input.CardId
	}

	v := validation.New()

	if data.ValidateEvent(v, event); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Events.Update(event)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"event": event}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) deleteEventHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	id, err := app.readID(params)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Events.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "event deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listEventHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Добавлять фильтрацию по card_id?
	var input struct {
		Title string
		Date  data.Date
		data.Filters
	}

	v := validation.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Date.Time = app.readDate(qs, "date", time.Now(), v)
	input.Sort = app.readString(qs, "sort", "id")
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 5, v)
	input.SortSafeList = []string{"id", "title", "date", "-id", "-title", "-date"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	events, metadata, err := app.models.Events.GetAll(input.Title, input.Date, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "events": events}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
