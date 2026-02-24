package handlers

// router.go constructs the chi router, registers all middleware, and wires all
// routes to their respective handlers. it is the single source of truth for
// the HTTP surface area of the corvus control plane API.
// adding a new endpoint means adding one line in this file, nothing else.

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/db"
)

// RouterDependencies groups all external dependencies that the router and
// its handlers need. passing a single struct instead of N arguments keeps
// CreateAndSetupRouter's signature stable as more handlers are added.
// adding a new dependency (eg, a docker client in a later phase) means
// adding one field here, not changing every call site.
type RouterDependencies struct {
	Logger   *slog.Logger
	Database *db.Database
}

// CreateAndSetupRouter constructs the chi multiplexer, attaches middleware, constructs
// all handlers with their dependencies, and registers all routes.
// it returns a plain http.Handler so main.go has no chi import or awareness.
// the server in main.go only needs to know it has something that satisfies http.Handler.
func CreateAndSetupRouter(dependencies RouterDependencies) http.Handler {

	router := chi.NewRouter() // type is *chi.Mux, implements http.Handler interface
	// Mux: Short for Multiplexer, this is the HTTP router (chi.Mux). It acts
	//    as a switchboard, inspecting incoming request URLs and routing them to
	//    the appropriate Go handler functions.

	// chi middleware runs on every request before the handler is called  (top to bottom).
	// Common use cases include authentication, rate limiting, CORS header injection,
	// and logging. They allow applying global rules without repeating code in every handler.
	// middleware.Logger logs the method, path, status code, and latency of every request.
	router.Use(middleware.Logger) // TODO replace with a custom slog middleware
	// middleware.Recoverer catches panics in handlers and returns a 500 instead of crashing the process.
	router.Use(middleware.Recoverer)
	// both are standard inclusions for any production HTTP service.

	// --- handler init/construction ---
	// each handler receives only the dependencies it actually needs.
	// handlers do not use for global variables (like having LOGGER as global) cuz of dependency injection.

	// /health needs only the logger
	healthHandler := NewHealthHandler(dependencies.Logger)

	// deployment handlers will need the database and logger
	deploymentHander := NewDeploymentHandler(dependencies.Database, dependencies.Logger)

	// --- route registration ---

	// The `/health` endpoint is intentionally kept at the root level rather
	// than under an /api prefix. External infrastructure components, such as
	// load balancers (AWS), container orchestrators (K8s), and uptime monitors, typically
	// expect health checks at standard root paths (`/health` and not `/api/health`) and do not have context
	// about the application's internal route grouping and API structure
	router.Get("/health", healthHandler.Health)

	// This is api route group (basically having an `/api/` prefix (/api/health) for all API routes
	// non-API routes like /health are kept outside this group intentionally.
	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Get("/deployments", deploymentHander.ListDeployments)
		// {id} is a placeholder for the actual id (like "happy-dog-1234"), `{id}` gets handled by chi library
		apiRouter.Get("/deployments/{id}", deploymentHander.GetDeployment)

		apiRouter.Post("/deployments", deploymentHander.CreateDeployment)

		// TODO redeploy and delete will be added in the future

		// placeholder to confirm the route group compiles correctly
		// if the routes are not registered yet
		_ = apiRouter

		// This is the less magic way to register routes without using chi's Route() grouping.
		/*
			// first, explicitly create a brand new, empty router
			apiRouter := chi.CreateAndSetupRouter()

			// attach the routes that want to be under /api to it
			apiRouter.GetDeployment("/deployments", deploymentHandler.ListDeployments)
			apiRouter.Post("/deployments", deploymentHandler.CreateDeployment)

			// mount the configured router into the main router under the "/api" path.
			router.Mount("/api", apiRouter)
		*/
	})

	return router
}
