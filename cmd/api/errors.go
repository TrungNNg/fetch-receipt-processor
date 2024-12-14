package main

import (
	"fmt"
	"net/http"
)

// logError logs an error message along with details of the current HTTP request.
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// errorResponse is a helper method for sending JSON-formatted error messages to the client.
//
// Parameters:
//   - w: The http.ResponseWriter where the response will be written.
//   - r: The *http.Request associated with the client request (useful for logging or context).
//   - status: The HTTP status code to include in the response.
//   - message: The error message to include in the response. This uses the `any` type
//     for flexibility, allowing various data types to be included in the error message.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	// If writeJSON failed, fall back to sending the client empty response with 500 Internal
	// Server Error status code.
	err := app.writeJSON(w, status, envelope{"description": message}, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// 500 Internal Server Error response
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// Generic 404 Not Found response
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// 404 Not Found response for getPoint endpoint
func (app *application) receiptIDNotFoundResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.errorResponse(w, r, http.StatusBadRequest, message)
}

// 405 Method Not Allowed response
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// 400 Bad Request response
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.errorResponse(w, r, http.StatusBadRequest, message)
}

// Rate Limit Exceeded response
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}
