# Corvus PaaS

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![Docker SDK](https://img.shields.io/badge/Docker-SDK_for_Go-2496ED?logo=docker&logoColor=white)
![Traefik](https://img.shields.io/badge/Traefik-v3-24A1C1?logo=traefikproxy&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-database%2Fsql-003B57?logo=sqlite&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-v4-06B6D4?logo=tailwindcss&logoColor=white)
![Cloudflare](https://img.shields.io/badge/Cloudflare-Tunnel-F38020?logo=cloudflare&logoColor=white)
![Nginx](https://img.shields.io/badge/Nginx-Alpine-009639?logo=nginx&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-yellow)

A self-hosted Platform-as-a-Service that takes a zip file or a public GitHub repo URL and deploys it as a live website with a unique public URL. Built from scratch in Go. No Heroku buildpacks, no managed services, no abstraction layers between the code and the containers. Raw Docker daemon control with Go, wired to Traefik for instant subdomain routing, exposed to the public internet through Cloudflare Tunnel.

> **Live demo (almost done, but not yet):** [corvus.sasta.dev](https://corvus.sasta.dev)

---

## What it does

1. You give it a **zip file**, a **GitHub repo URL**, or pick a **one-click showcase preset**.
2. The backend clones the repo (or extracts the zip), runs your build command inside an **ephemeral Docker container**, and copies the output.
3. An **Nginx container** spins up to serve the static files.
4. **Traefik** detects the new container via Docker labels and routes a unique subdomain to it. Instantly, with zero config reload.
5. You get a **live public URL** in seconds.

Deployments auto-expire after a configurable TTL (default time-to-live 15 minutes). A background cleanup loop tears down expired containers, removes files, and deletes the database row.

---

## Architecture

```
Browser
  │
  ├── corvus.sasta.dev (React frontend)
  │     │
  │     │  fetch() over HTTPS
  │     ▼
  │   Cloudflare Tunnel ──→ Go Control Plane :8080
  │                              │
  │                              ├── Docker SDK ──→ Ephemeral build containers (node:20-alpine)
  │                              ├── Docker SDK ──→ Per-deployment Nginx containers (nginx:alpine)
  │                              ├── SQLite ──→ Deployment state & metadata
  │                              └── Structured logging (slog) ──→ Per-deployment log files
  │
  └── <slug>.corvus.sasta.dev
        │
        ▼
      Cloudflare Tunnel ──→ Traefik :80 ──→ deploy-<slug> (Nginx container)
```

Everything runs on a single VM in a home lab or cloud server. The Go backend, Traefik, and all deployment containers are sibling Docker containers. The backend controls Docker via the mounted socket (`/var/run/docker.sock`). The entire stack sits inside a dedicated VM for blast-radius containment. If a user's build script does something destructive, it trashes the ephemeral build container, not the host.

---

## Tech Stack

| Layer | Technology | Why |
|---|---|---|
| **Control Plane** | Go, chi router, `log/slog` | Docker itself is written in Go, so the SDK is native. Goroutines handle concurrent deploy pipelines naturally. Static binary means no runtime in the container. |
| **Container Orchestration** | Docker SDK for Go | Direct programmatic control over container lifecycle. No shell-exec wrappers. |
| **Reverse Proxy** | Traefik v3 | Watches the Docker socket and picks up routing rules from container labels instantly. No config file reload needed. This is what enables the "instant live URL" experience. |
| **Build Isolation** | Ephemeral `node:20-alpine` containers | User build commands run in throwaway containers with bind-mounted source dirs. The Go backend never runs `npm` or user code directly. |
| **Web Servers** | Per-deployment `nginx:alpine` containers | Each site gets its own container with a read-only bind mount to its static files. Restart policy `unless-stopped` keeps deployments alive across VM reboots. |
| **Database** | SQLite via `database/sql` + `go-sqlite3` | Single-file, zero-ops persistence. The query layer uses standard `database/sql`, so swapping to Postgres requires changing the driver and DSN, nothing else. |
| **Public Exposure** | Cloudflare Tunnel | Zero router config, no exposed home IP, free TLS termination. Wildcard subdomain routing through the tunnel to Traefik. |
| **Frontend** | React 19, TypeScript, Vite, Tailwind CSS v4, Radix UI | Hosted on Vercel. Communicates with the backend via `fetch()`. Black-and-white theme designed for a web developer to restyle. |

---

## Features

### Deployment Sources
- **One-click presets:** Vite Starter, React App, About Corvus, or a custom "Your Message" page with user-provided text injected as a build-time env var
- **Zip upload:** Drag-and-drop a `.zip` file (up to 50MB) with optional build command and output directory
- **GitHub repo:** Paste a public repo URL with branch, build command, and output directory

### Build Pipeline
- Git clone via `exec.Command` with stdout/stderr captured to per-deployment log files
- Build commands run in ephemeral `node:20-alpine` containers with the source directory bind-mounted at `/workspace`
- Environment variables passed to the build container (supports `VITE_*` env vars for frontend frameworks)
- Build output copied to persistent asset storage, then served via Nginx
- Automatic cleanup of temp directories and ephemeral containers on success or failure

### Deployment Lifecycle
- **Status tracking:** `deploying` → `live` → `expired`, or `deploying` → `failed`
- **Redeploy:** Re-runs the full pipeline, stops the old container, starts a new one
- **Delete:** Stops container, removes static files, removes log file, deletes DB row
- **Auto-expiration:** Background goroutine checks for expired deployments every 30 seconds and runs the full teardown sequence
- **TTL system:** Default 15-minute TTL, extended TTL with a friend code

### Routing
- Each deployment gets a unique slug (`adjective-noun-hex`, e.g. `swift-hawk-c142`)
- Traefik auto-discovers containers via Docker labels and routes `<slug>.corvus.sasta.dev` to the correct Nginx container
- Wildcard DNS + Cloudflare Tunnel handles public routing

### Frontend
- Landing page with hero section, tabbed deploy panel (Quick Deploy / Zip Upload / GitHub Repo), and active deployment view
- Real-time progress view with simulated steps driven by status polling
- Live deployment card with countdown timer, live URL, copy-to-clipboard, and action buttons
- Deployment viewer page (`/d/:id`) with full metadata, timestamps, and source info
- Session enforcement: one active deployment per browser via `localStorage`
- Friend code system for extended deployment TTL
- Toast notifications for API errors and status changes
- Fully responsive with Tailwind utility classes

---


## What makes this non-trivial

This is not a CRUD app with a different skin. The deployment pipeline crosses multiple process boundaries and failure domains that do not exist in typical web applications:

- **Cross-container file sharing.** The Go backend writes static files to a host directory. A separate Nginx container reads those same files via a bind mount. Getting the volume architecture, ownership, and permissions right between two independent containers requires understanding how Docker bind mounts interact with the host filesystem at the UID/GID level.

- **Ephemeral container lifecycle management.** Each build spins up a throwaway `node:20-alpine` container, bind-mounts the source directory, runs the user's build command, waits for exit, reads the logs, checks the exit code, removes the container, and then copies the output. All of this is orchestrated through the Docker SDK in Go, not shell scripts. Any step can fail, and the pipeline must handle partial failures, clean up resources, and update status correctly.

- **Dynamic reverse proxy routing without config reloads.** Traefik watches the Docker socket and picks up routing rules from container labels at the moment a container starts. This means the Go backend has to attach the correct Traefik labels programmatically when creating each Nginx container so that the subdomain is routable before the API even responds. Getting the label format, network attachment, and timing right is the kind of infrastructure plumbing that only shows up when you're building the platform itself, not building on top of one.

- **Async pipeline with state machine.** Deployments take 30 to 120 seconds. The HTTP handler returns immediately with a `202 Accepted` and the pipeline runs in a background goroutine. The frontend polls for status. The backend manages state transitions (`deploying` / `live` / `failed`) in SQLite, and each transition must be atomic relative to the container lifecycle. A status update that disagrees with the actual container state is a bug that is invisible until production.

- **TTL-based auto-expiration.** A background goroutine runs a cleanup loop on a ticker, queries for expired deployments, and runs the full teardown sequence (stop container, remove files, remove logs, delete DB row) for each one. This is the same teardown path as the DELETE endpoint, reused to avoid drift between manual and automatic cleanup.

- **No frameworks, no PaaS SDKs, no abstraction layers.** The control plane talks directly to the Docker daemon via the Go SDK. Traefik configuration is entirely label-driven. The build pipeline is hand-rolled. The slug generator, the directory copier, the zip extractor, the log writer, the env var decoder: all written from scratch. Every layer of the stack is visible in the source.

None of these problems are individually unsolvable. But they all have to work together, across process boundaries, in the right order, with correct cleanup on every failure path. That is what makes this a platform engineering project rather than an application engineering project.

---

## Project Structure

```
corvus-paas/
├── corvus-control-plane/          # Go backend
│   ├── main.go                    # Entry point, server setup, graceful shutdown
│   ├── config/config.go           # Env-based configuration with defaults
│   ├── handlers/
│   │   ├── router.go              # chi router, middleware, route registration
│   │   ├── deployments.go         # CRUD + redeploy handlers
│   │   ├── health.go              # GET /health
│   │   ├── cors.go                # CORS middleware
│   │   └── helpers.go             # JSON response helpers
│   ├── build/
│   │   ├── pipeline.go            # DeployerPipeline struct and constructor
│   │   ├── pipeline_github_deploy.go   # GitHub clone → build → serve pipeline
│   │   ├── pipeline_zip_deploy.go      # Zip extract → build → serve pipeline
│   │   ├── pipeline_nginx_server.go    # Shared: copy to asset storage, start Nginx
│   │   ├── pipeline_cleanup.go    # Teardown: stop container, remove files, delete DB row
│   │   ├── pipeline_logger.go     # Dual logging (slog + per-deployment log file)
│   │   ├── expiration.go          # Background TTL cleanup loop
│   │   ├── git_clone.go           # git clone wrapper
│   │   ├── zip_extract.go         # archive/zip extraction
│   │   └── env_vars.go            # JSON env var decoding
│   ├── docker/
│   │   ├── client.go              # Docker SDK wrapper, image pulling
│   │   ├── builder.go             # Ephemeral build container logic
│   │   └── nginx.go               # Nginx container creation with Traefik labels
│   ├── db/
│   │   ├── db.go                  # SQLite connection, schema migration
│   │   └── deployments.go         # Deployment CRUD queries
│   ├── models/models.go           # Deployment struct, status constants
│   └── util/
│       ├── slug.go                # "adjective-noun-hex" slug generator
│       └── copy.go                # Recursive directory copy
│
├── corvus-frontend/               # React/TypeScript frontend
│   └── src/
│       ├── api/                   # fetch wrappers (client.ts, deployments.ts)
│       ├── components/
│       │   ├── deploy/            # DeployPanel, tabs, presets, drag-drop, GitHub form
│       │   ├── deployment/        # ActiveDeploymentView, DetailCard, StatusBadge, Countdown
│       │   ├── progress/          # DeployProgressView, ProgressStep
│       │   ├── layout/            # Header, Footer, HeroSection, Logo
│       │   └── shared/            # FriendCodeInput, Toast
│       ├── hooks/                 # useDeploymentPolling, useCountdown, useActiveDeployment, useFriendCode
│       ├── pages/                 # LandingPage, DeploymentViewerPage
│       ├── config/constants.ts    # API URL, TTL, polling intervals, preset configs
│       ├── types/deployment.ts    # TypeScript types matching backend JSON
│       └── lib/utils.ts           # formatFileSize, URL parsing, countdown formatting
│
└── server-infrastructure/
    └── router-traefik/
        └── docker-compose.yml     # Traefik container config
```

---

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/api/deployments` | List all deployments |
| `POST` | `/api/deployments` | Create deployment (multipart/form-data for both zip and GitHub) |
| `GET` | `/api/deployments/:uuid` | Get deployment by ID |
| `DELETE` | `/api/deployments/:uuid` | Delete deployment (stops container, removes files, deletes row) |
| `POST` | `/api/deployments/:uuid/redeploy` | Trigger redeploy (re-runs full pipeline) |
| `GET` | `/api/validate-code` | Validate a friend code |

All create requests use `multipart/form-data`, even GitHub deploys that have no file, because the backend parses everything through a single multipart handler.

---

## Key Architectural Decisions

**Why Go?** Docker is written in Go. The Docker SDK is native, not a binding. Goroutines handle concurrent deployment pipelines without thread pool configuration. The compiled binary has zero runtime dependencies.

**Why Traefik over Nginx as reverse proxy?** Nginx requires a config file reload to add new routes. Traefik watches the Docker socket and updates its routing table the instant a container is created or destroyed. This is what makes the "deploy and get a live URL in seconds" experience work.

**Why SQLite?** For a single-node system, SQLite is the correct choice. No connection pooling, no separate process, no ops overhead. The `database/sql` interface means swapping to Postgres later is a driver change, not a rewrite.

**Why Docker-in-VM, not bare metal?** Blast radius containment. If a user's build command does something destructive, it trashes the ephemeral build container inside the VM, not the host machine.

**Why ephemeral build containers?** User-provided build commands (`npm ci && npm run build`) must not run in the control plane process. Each build gets its own `node:20-alpine` container with the source bind-mounted at `/workspace`. The container exits after the build, gets removed, and the output is copied to persistent storage.

**Why async deploy with polling?** A deployment can take 30 to 120 seconds (npm install is slow). Holding an HTTP connection open that long is unreliable. The backend accepts the request, returns `202`, and the frontend polls `GET /api/deployments/:id` every 2 seconds until the status reaches `live` or `failed`.

**Why Cloudflare Tunnel?** Zero router configuration, no exposed home IP, free TLS termination, DDoS protection. Architecturally identical to what startups use before they have cloud VMs.

---

## Running Locally

### Prerequisites
- Go 1.23+
- Docker Engine
- Node.js 20+ (for the frontend)
- Git

### Backend

```bash
cd corvus-control-plane

# Set up the Traefik network (one time)
docker network create corvus-paas-network

# Start Traefik
cd ../server-infrastructure/router-traefik
docker compose up -d
cd ../../corvus-control-plane

# Run the backend
go run main.go
```

The backend reads configuration from environment variables with sensible defaults:

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DB_PATH` | `./corvus.db` | SQLite database file path |
| `ASSET_STORAGE_ROOT` | `/srv/corvus-paas/deployments` | Where static files are stored |
| `LOG_ROOT` | `/srv/corvus-paas/logs` | Where per-deployment log files are written |
| `TRAEFIK_NETWORK` | `corvus-paas-network` | Docker network shared with Traefik |
| `LOG_FORMAT` | `text` | `json` or `text` |
| `CORS_ORIGIN` | `*` | Allowed CORS origin |
| `FRIEND_CODE` | *(empty)* | Secret code for extended TTL |
| `DEFAULT_TTL_MINUTES` | `15` | Deployment lifetime in minutes |
| `EXTENDED_TTL_MINUTES` | `60` | Extended lifetime with friend code |

### Frontend

```bash
cd corvus-frontend
npm install
npm run dev
```

Create a `.env.development` file:

```
VITE_API_BASE_URL=http://localhost:8080
```

---

## Deployment State Machine

```
[create request] ──→ [deploying] ──→ [live] ──→ [expired / deleted]
                          │
                          └──→ [failed]

[live] ──→ [redeploy] ──→ [deploying] ──→ [live]
                               │
                               └──→ [failed]
```

The backend manages all state transitions. The expiration cleanup loop runs every 30 seconds, queries for deployments past their TTL, and runs the full teardown (stop container → remove files → remove log → delete DB row).

---

## Roadmap

Corvus is currently a single-repo monolith with the frontend and backend side by side. The plan is to split these into separate repositories as the project matures, with the backend becoming a standalone open-source PaaS engine and the frontend becoming one possible interface for it.

The current codebase is designed so that the following can be added without rewriting existing systems:

- **Log streaming.** Build logs are already written to per-deployment files on disk. Adding a `GET /api/deployments/:id/logs` SSE endpoint is a read path, not a pipeline change.
- **GitHub webhooks.** `auto_deploy` and `webhook_secret` fields exist in the DB schema. The handler needs HMAC-SHA256 verification and async redeploy trigger.
- **Private repos.** GitHub OAuth flow for access token, used in git clone via HTTPS auth.
- **Custom domains.** Update Traefik labels with `Host('custom.com')`, user sets DNS CNAME.
- **Dynamic app hosting.** Port allocation pool for Node/Python/Go servers instead of static Nginx.
- **Multi-user auth.** JWT-based auth, `user_id` on deployment rows.
- **Multi-node orchestration.** The current architecture is single-node. Scaling to multiple nodes requires a coordination layer (container placement, health checks, shared storage). This is the path toward a general-purpose PaaS.

The long-term goal is a PaaS platform that works at any scale: from a single home lab VM to a multi-node production cluster. The v1 exists to prove the pipeline works end-to-end. Everything after that is about removing the single-node assumption.

---

## License

[MIT](./LICENSE)
