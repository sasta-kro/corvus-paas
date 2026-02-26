// Package handlers  contains all HTTP handler functions for the corvus control plane API.
// each handler file groups related endpoints by resource or concern.
// handlers receive a decoded request, call into the db or deployerPipeline layer, and write a JSON response.
// no business logic lives in handlers; they are thin translation layers between HTTP and the domain.

package handlers

import (
	"log/slog"
	"net/http"
	"time"
)

// Go HTTP Handler Design Quirks:
//
// Constructors: Go lacks classes and magical constructors like other OOP languages (Java, Python)
// Initialization is done via standard functions conventionally named New[Type]().
//
// Methods vs Functions: Methods require a receiver (an already created, existing instance
//    in memory). Therefore, a constructor must be a standalone function,
//    while behaviors (like handling a request) are methods attached to the instance.
//
// The Chi Router: Chi routes traffic but relies entirely on the standard (net/http)
//    library's http.ResponseWriter and *http.Request. This keeps handlers
//    framework-agnostic and highly portable. (doesn't get framework-locked)

// HealthHandler holds the dependencies needed by the health endpoint.
// even though health currently needs no dependencies, using a struct keeps
// the pattern consistent with all other handlers that do need db or logger access.
// this avoids a later refactor when dependencies are added.
type HealthHandler struct {
	logger *slog.Logger
}

// NewHealthHandler constructs a HealthHandler with the given logger. (basically the constructor)
// the constructor pattern (NewXxx) is the standard Go way to create a struct
// that has unexported fields, since callers outside the package cannot set them directly.
func NewHealthHandler(inputLogger *slog.Logger) *HealthHandler {
	return &HealthHandler{logger: inputLogger}
}

// healthResponse is the JSON body returned by the health endpoint.
// keeping the response struct local to this file means it is not accidentally
// reused or confused with domain models.
type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Health handles GET /health.
// returns a 200 OK with a JSON body confirmation if the service is running.
// this endpoint is intentionally simple: no db check, no auth, no business logic.
// it is the minimum signal that the process is alive and the HTTP stack works.
// a more thorough readiness check (db ping, docker socket check) can be added
// at GET /ready in a future step.
// handler functions always have the signature (http.ResponseWriter, *http.Request) at the beginning,
// because they are called by the net/http library with these args when a request is routed to them.
// The ResponseWriter is used to construct the HTTP response to the client/brower/api,
// and the Request contains all the details of the incoming HTTP request from the client.
func (handler *HealthHandler) Health(responseWriter http.ResponseWriter, request *http.Request) {
	response := healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		// RFC3339 is the universally accepted standard string format for safely
		// transmitting UTC timestamps in JSON web APIs.
	}

	// WriteJSON is called as a helper here.
	// the response body is always JSON, so centralizing the encoding & content-type header
	// maker in a helper avoids repeating the same 4 lines in every handler function.
	writeJsonAndRespond(responseWriter, http.StatusOK, response)
}
