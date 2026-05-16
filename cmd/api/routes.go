package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Health check route
	router.GET("/v1/healthcheck", app.healthCheckHandler)

	// Add other routes here

	//Energy document Routes
	router.HandlerFunc(http.MethodGet, "/v1/EnergyDocuments", app.requireActivatedUser(app.listEnergyDocuments))
	router.HandlerFunc(http.MethodPost, "/v1/EnergyDocument", app.requireActivatedUser(app.createEnergyDocumentHandler))
	router.HandlerFunc(http.MethodGet, "/v1/EnergyDocument/:id", app.requireActivatedUser(app.showEnergyDocumentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/EnergyDocument/:id", app.requireActivatedUser(app.updateEnergyDocument))
	router.HandlerFunc(http.MethodDelete, "/v1/EnergyDocument/:id", app.requireActivatedUser(app.deleteEnergyDocument))

	//user routes
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.POST("/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())
	return app.metrics(app.recoverPanic(app.enableCORS(app.cancelRequest(app.rateLimit(app.authenticate(router))))))
}
