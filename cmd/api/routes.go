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
	router.GET("/v1/events", app.requireActivatedUser(app.listEventHandler))
	router.GET("/v1/events/:id", app.requireActivatedUser(app.showEventHandler))
	router.POST("/v1/events", app.requireActivatedUser(app.createEventHandler))
	router.PATCH("/v1/events/:id", app.requireActivatedUser(app.updateEventHandler))
	router.DELETE("/v1/events/:id", app.requireActivatedUser(app.deleteEventHandler))

	router.GET("/v1/cards/:id", app.requireActivatedUser(app.showCardHandler))
	router.POST("/v1/cards", app.requireActivatedUser(app.createCardHandler))
	router.PATCH("/v1/cards/:id", app.requireActivatedUser(app.updateCardHandler))

	router.POST("/v1/users", app.registerHandler)
	router.PUT("/v1/users/activated", app.activateUserHandle)

	router.POST("/v1/tokens/activation", app.sendTokenHandler)
	router.POST("/v1/tokens/authentication", app.createAuthenticationToken)

	return app.logRequests(app.recoverPanic(app.rateLimit(app.authenticate(router))))
}
