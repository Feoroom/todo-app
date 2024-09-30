package main

import (
	"context"
	"library/internal/data"
	"net/http"
)

type ctxKey string

const userCtxKey = ctxKey("user")
const roleCtxKey = ctxKey("role")

func (app *Application) ctxSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userCtxKey, user)

	return r.WithContext(ctx)
}

func (app *Application) ctxGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userCtxKey).(*data.User)
	if !ok {
		panic("missing user value in ctx")
	}
	return user
}

func (app *Application) ctxSetRole(r *http.Request, role string) *http.Request {
	ctx := context.WithValue(r.Context(), roleCtxKey, role)

	return r.WithContext(ctx)
}
