package handlers

import "net/http"

// CORSMiddleware adds the required CORS headers to every response so the
// frontend (hosted on a different origin like Vercel/Netlify) can make
// fetch() requests to this API.
// For production, AllowOrigin should be restricted to the actual frontend
// domain. During development, "*" or "http://localhost:5173" is acceptable.
func CORSMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			responseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			responseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			// preflight requests (OPTIONS) get an immediate 204 response
			// with no body. the browser sends these automatically before
			// cross-origin POST/DELETE requests to check if the server allows them.
			if request.Method == http.MethodOptions {
				responseWriter.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(responseWriter, request)
		})
	}
}
