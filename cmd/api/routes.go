package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)
	//router.GET("/swagger/:any", httpSwagger.WrapHandler)

	router.HandlerFunc(http.MethodGet, "/v1/info", app.info)

	//router.HandlerFunc(http.MethodGet, "/v1/events/:id", app.showEventHandler)
	router.GET("/v1/events", app.listEventHandler)
	router.GET("/v1/events/:id", app.showEventHandler)
	router.POST("/v1/events", app.createEventHandler)
	router.PATCH("/v1/events/:id", app.updateEventHandler)
	router.DELETE("/v1/events/:id", app.deleteEventHandler)

	router.GET("/v1/cards/:id", app.showCardHandler)
	router.POST("/v1/cards", app.createCardHandler)
	router.PATCH("/v1/cards/:id", app.updateCardHandler)

	router.POST("/v1/users", app.registerHandler)
	router.PUT("/v1/users/activated", app.activateUserHandle)

	router.POST("/v1/tokens/activation", app.sendTokenHandler)

	return app.logRequests(app.recoverPanic(app.rateLimit(router)))
}
