package main

import (
	"context"
	"library/internal/data"
	"net/http"
)

type ctxKey string

const userCtxKey = ctxKey("user")

func (app *application) ctxSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userCtxKey, user)
	return r.WithContext(ctx)
}

func (app *application) cxtGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userCtxKey).(*data.User)
	if !ok {
		panic("missing user value in ctx")
	}
	return user
}
