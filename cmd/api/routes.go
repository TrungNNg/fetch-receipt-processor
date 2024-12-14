package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes sets up the application's HTTP routes and their corresponding handlers.
// https://github.com/julienschmidt/httprouter
//
// This method configures the httprouter instance, specifying custom handlers for
// 404 Not Found and 405 Method Not Allowed responses. It also registers routes
// for API endpoints, linking HTTP methods and URL patterns to specific handler functions.
//
// Additionally, the method wraps the router with middleware, such as panic recovery
// and rate limiting, before returning the final http.Handler instance.
//
// Returns:
// - An http.Handler instance with all routes and middleware configured.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Set up default response for 404 and 405
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register routes
	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/receipts/process", app.processReceiptHandler)
	router.HandlerFunc(http.MethodGet, "/receipts/:id/points", app.getPointsHandler)

	// Register rateLimit and recoverPanic middleware
	return app.recoverPanic(app.rateLimit(router))
}
