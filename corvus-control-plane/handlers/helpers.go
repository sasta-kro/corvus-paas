package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// writeJsonAndRespond serializes/marshalls/encodes/converts the given payload/data to JSON and writes it to the response.
// it sets Content-Type to application/json and the given HTTP status code.
// (deduplicates the 2 lines of w.Header().Set() and json.NewEncoder().Encode() that would be repeated in every handler)
// (or in this case, the 3 lines of [ responseWriter.Header().Set() + json.Marshal() + w.Write() ].
//
// if JSON encoding fails (which should not happen with well-defined response structs),
// it falls back to a plain text 500 response.
// all handlers use this function instead of calling json.NewEncoder directly,
// keeping the response format consistent across the entire API.
func writeJsonAndRespond(responseWriter http.ResponseWriter, statusCode int, dataPayload any) {
	responseWriter.Header().Set("Content-Type", "application/json")

	// .Marshal() is just a function that takes a Go value and returns its JSON encoding as a []byte (byte array)
	// literally just .convertToJsonByteArray() (this is just an example)
	// Marshaling = Serialization = Converting a value into a transmittable format (in this case, JSON string as bytes)
	serializedData, err := json.Marshal(dataPayload)
	// Marshal vs Encoder: json.Marshal() buffers the entire JSON dataPayload in
	//    memory and converts it into a []byte array before writing. This prevents
	//    the "200 OK trap" where json.NewEncoder() might encounter an error halfway through a stream
	//    after the HTTP 200 status header (Everything OK) has already been sent to the client.

	if err != nil {
		// encoding to json failed, fall back to a minimal plain text error.
		// this branch should never be reached with statically typed response structs,
		// but the fallback prevents a silent empty response which is harder to debug.
		http.Error(responseWriter, `{"error":"internal encoding error"}`, http.StatusInternalServerError)
		return
	}

	// if the HTTP status code header is not set, Go defaults to 200 OK everytime.
	// It is good practice to explicitly set the status code.
	responseWriter.WriteHeader(statusCode)

	// The strict order of operations for http.ResponseWriter is:
	// [1] set headers `.Header.Set()`, [2] call WriteHeader(statusCode), and finally use `Write()`
	//    to send the byte dataPayload.
	responseWriter.Write(serializedData) // nolint:errcheck -- write errors are not actionable on the server side
	// The nolint comment is used to tell linters (like golangci-lint) to ignore the error returned by Write(),
	// because in this context, if the client disconnects before the response is fully sent,
	// there's nothing the server can do about it.
}

// writeErrorJsonAndLogIt logs the error at level ERROR and
// writes a standard JSON error response to the client with the given HTTP status code and message.
// this keeps error response shape consistent:
//
//	{"error": "some human-readable message"}
//
// callers pass in a logger so the error is also logged server-side with context.
// the error message sent to the client is always a controlled string,
// never a raw Go error, to avoid leaking internal implementation details.
func writeErrorJsonAndLogIt(
	responseWriter http.ResponseWriter,
	statusCode int,
	message string,
	logger *slog.Logger,
) {
	logger.Error("request error", "status", statusCode, "message", message)
	writeJsonAndRespond(responseWriter, statusCode, map[string]string{"error": message})
}
