# Corvus PaaS

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![Docker SDK](https://img.shields.io/badge/Docker-SDK_for_Go-2496ED?logo=docker&logoColor=white)
![Traefik](https://img.shields.io/badge/Traefik-v3-24A1C1?logo=traefikproxy&logoColor=white)
![Cloudflare](https://img.shields.io/badge/Cloudflare-Tunnel-F38020?logo=cloudflare&logoColor=white)
![Nginx](https://img.shields.io/badge/Nginx-Alpine-009639?logo=nginx&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-yellow)
![SQLite](https://img.shields.io/badge/SQLite-database%2Fsql-003B57?logo=sqlite&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?logo=typescript&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-v4-06B6D4?logo=tailwindcss&logoColor=white)


A self-hosted Platform-as-a-Service that takes a zip file or a public GitHub repo URL and deploys it as a live website with a unique public URL. Built from scratch in Go. No Heroku buildpacks, no managed services, no abstraction layers between the code and the containers. The control plane talks directly to the Docker daemon via the Go SDK, attaches Traefik routing labels programmatically, and exposes everything to the public internet through a Cloudflare Tunnel.

> **Live demo (almost done, but not yet):** [corvus.sasta.dev](https://corvus.sasta.dev)

---

## What it does

1. You give it a **zip file**, a **GitHub repo URL**, or pick a **one-click preset**.
2. The backend clones the repo (or extracts the zip), runs your build command inside an **ephemeral Docker container**, and copies the output.
3. An **Nginx container** spins up to serve the static files.
4. **Traefik** detects the new container via Docker labels and routes a unique subdomain to it. Instantly, with zero config reload.
5. You get a **live public URL** in seconds.

Deployments auto-expire after a configurable TTL (default 15 minutes). A background cleanup loop tears down expired containers, removes files, and deletes the database row.

---

## Architecture

```
Browser
  |
  |-- corvus.sasta.dev (React frontend)
  |     |
  |     |  fetch() over HTTPS
  |     v
  |   Cloudflare Tunnel --> Go Control Plane :8080
  |                              |
  |                              |-- Docker SDK --> Ephemeral build containers (node:20-alpine)
  |                              |-- Docker SDK --> Per-deployment Nginx containers (nginx:alpine)
  |                              |-- SQLite --> Deployment state and metadata
  |                              +-- Structured logging (slog) --> Per-deployment log files
  |
  +-- <slug>.corvus.sasta.dev
        |
        v
      Cloudflare Tunnel --> Traefik :80 --> deploy-<slug> (Nginx container)
```

Everything runs on a single VM. The Go backend, Traefik, and all deployment containers are sibling Docker containers. The backend controls Docker via the mounted socket (`/var/run/docker.sock`). The entire stack sits inside a dedicated VM for blast-radius containment. If a user's build script does something destructive, it trashes the ephemeral build container, not the host.

---

## Tech Stack

| Layer | Technology | Why |
|---|---|---|
| **Control Plane** | Go, chi router, `log/slog` | Docker is written in Go, so the SDK is native, not a binding. Goroutines handle concurrent deploy pipelines without thread pool configuration. Static binary compiles to a single executable with zero runtime dependencies. |
| **Container Orchestration** | Docker SDK for Go | Direct programmatic control over the full container lifecycle (pull, create, start, wait, logs, remove). No shell-exec wrappers, no `docker` CLI calls. |
| **Reverse Proxy** | Traefik v3 (Docker provider) | Watches the Docker socket and picks up routing rules from container labels at the instant a container starts. No config file reload, no template generation. This is what enables the "instant live URL" experience. Nginx as a reverse proxy would require writing config files and triggering reloads on every deploy. |
| **Build Isolation** | Ephemeral `node:20-alpine` containers | User build commands run in throwaway containers with bind-mounted source directories. The Go backend never runs `npm install`, `npx`, or any user code in its own process. The build container is created, started, waited on, log-drained, and removed, all through the Docker SDK. |
| **Web Servers** | Per-deployment `nginx:alpine` containers | Each deployed site gets its own isolated container with a read-only bind mount to its static files. Restart policy `unless-stopped` keeps deployments alive across VM reboots without an external process manager. |
| **Database** | SQLite via `database/sql` + `go-sqlite3` | Single-file, zero-ops persistence for a single-node system. The query layer uses Go's standard `database/sql` interface, so swapping to Postgres is a driver and DSN change, not a rewrite. |
| **Public Exposure** | Cloudflare Tunnel | Zero router config, no exposed home IP, free TLS termination, DDoS protection. Wildcard subdomain routing through the tunnel to Traefik. |
| **Frontend** | React 19, TypeScript, Vite, Tailwind CSS v4, Radix UI | Communicates with the backend via `fetch()`. Minimal black-and-white theme designed for a web developer to restyle. |

---

## Features

### Deployment Sources
- **One-click presets:** Vite Starter, React App, About Corvus, or a custom "Your Message" page with user-provided text injected as a build-time environment variable
- **Zip upload:** Drag-and-drop a `.zip` file (up to 50MB) with optional build command and output directory
- **GitHub repo:** Paste a public repo URL with branch, build command, and output directory

### Build Pipeline
- Git clone via `exec.Command` with stdout/stderr captured to per-deployment log files on disk
- Build commands executed in ephemeral `node:20-alpine` containers with the source directory bind-mounted at `/workspace`
- User-defined environment variables decoded from JSON, passed to the build container, and available to the build process (supports `VITE_*` and similar framework env vars)
- Build output validated (output directory must exist), copied to persistent asset storage, then served via Nginx
- Automatic cleanup of temp directories and ephemeral build containers on both success and failure

### Deployment Lifecycle
- **Status tracking:** `deploying` > `live` > `expired`, or `deploying` > `failed`
- **Redeploy:** Re-runs the full pipeline for the same deployment (GitHub re-clones and rebuilds, zip re-serves from stored assets)
- **Delete:** Full teardown: stops the Nginx container, removes static files from disk, removes the log file, deletes the database row
- **Auto-expiration:** A background goroutine on a 30-second ticker queries for deployments past their TTL and runs the same full teardown sequence as manual delete
- **TTL system:** Default 15-minute TTL, with extended TTL granted when a valid friend code is provided at deploy time

### Routing
- Each deployment gets a unique slug in `adjective-noun-hex` format (e.g. `swift-hawk-c142`)
- Traefik auto-discovers containers via Docker labels and routes `<slug>.corvus.sasta.dev` to the correct Nginx container
- Wildcard DNS + Cloudflare Tunnel handles public routing without any per-deployment DNS configuration

### Frontend
- Landing page with tabbed deploy panel (Quick Deploy / Zip Upload / GitHub Repo) and real-time progress view
- Live deployment card with countdown timer, clickable URL, copy-to-clipboard, and action buttons (open, redeploy, delete)
- Deployment detail page (`/d/:id`) with full metadata, source info, and timestamps
- One active deployment per browser session, enforced via `localStorage`
- Friend code input in the header for extended TTL
- Toast notification system for API responses and errors

---

## What makes this non-trivial

This is not a CRUD app with a different skin. The deployment pipeline crosses multiple process boundaries and failure domains that do not exist in typical web applications:

- **Cross-container file sharing.** The Go backend writes static files to a host directory. A separate Nginx container reads those same files via a read-only bind mount. Getting the volume architecture, ownership, and permissions right between two independent containers requires understanding how Docker bind mounts interact with the host filesystem at the UID/GID level.

- **Ephemeral container lifecycle management.** Each build spins up a throwaway container, bind-mounts the source directory, runs the user's build command, waits for exit, reads the multiplexed stdout/stderr logs via `stdcopy.StdCopy`, checks the exit code, removes the container, and then copies the output. All orchestrated through the Docker SDK. Any step can fail, and the pipeline handles partial failures, cleans up resources, and updates status correctly at every point.

- **Dynamic reverse proxy routing without config reloads.** Traefik watches the Docker socket and picks up routing rules from container labels at the moment a container starts. The Go backend attaches the correct labels programmatically when creating each Nginx container so that the subdomain is routable before the create-deployment API call even responds. Getting the label format, network attachment timing, and container naming conventions right is the kind of infrastructure plumbing that only shows up when you're building the platform itself, not building on top of one.

- **Async pipeline with state machine.** Deployments take 30 to 120 seconds. The HTTP handler returns immediately with `202 Accepted` and the pipeline runs in a background goroutine with its own `context.Background()` (because the request context is already canceled by the time the build starts). The backend manages state transitions (`deploying` / `live` / `failed`) in SQLite, and each transition must be consistent with the actual container state.

- **TTL-based auto-expiration.** A background goroutine runs a cleanup loop on a ticker, queries for expired deployments, and runs the full teardown sequence (stop container, remove files, remove logs, delete DB row) for each one. This is the same teardown code path as the DELETE endpoint, reused to avoid drift between manual and automatic cleanup.

- **No frameworks, no PaaS SDKs, no abstraction layers.** The control plane talks directly to the Docker daemon via the Go SDK. Traefik configuration is entirely label-driven. The build pipeline is hand-rolled. The slug generator, the directory copier, the zip extractor, the pipeline logger, the env var decoder: all written from scratch because they had to be. Every layer of the stack is visible in the source.

None of these problems are individually unsolvable. But they all have to work together, across process boundaries, in the right order, with correct cleanup on every failure path. That is what makes this a platform engineering project rather than an application engineering project.

---

## Project Structure

```
corvus-paas/
|-- corvus-control-plane/       # Go backend: API server, build pipeline, Docker orchestration
|-- corvus-frontend/            # React/TypeScript frontend: deploy UI, progress tracking, deployment management
+-- server-infrastructure/      # Traefik reverse proxy configuration (Docker Compose)
```

---

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/api/deployments` | List all deployments |
| `POST` | `/api/deployments` | Create deployment (multipart/form-data) |
| `GET` | `/api/deployments/:uuid` | Get deployment by ID |
| `DELETE` | `/api/deployments/:uuid` | Delete deployment (full teardown) |
| `POST` | `/api/deployments/:uuid/redeploy` | Trigger redeploy |
| `GET` | `/api/validate-code` | Validate a friend code |

---

## Running Locally

### Prerequisites
- Go 1.23+
- Docker Engine
- Node.js 20+ (for the frontend dev server)
- Git

### Backend

```bash
cd corvus-control-plane

# Create the shared Docker network (one time)
docker network create corvus-paas-network

# Start Traefik
cd ../server-infrastructure/router-traefik
docker compose up -d
cd ../../corvus-control-plane

# Run the backend
go run main.go
```

Configuration via environment variables (all have sensible defaults):

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DB_PATH` | `./corvus.db` | SQLite database file path |
| `ASSET_STORAGE_ROOT` | `/srv/corvus-paas/deployments` | Static file storage root |
| `LOG_ROOT` | `/srv/corvus-paas/logs` | Per-deployment log file directory |
| `TRAEFIK_NETWORK` | `corvus-paas-network` | Docker network shared with Traefik |
| `LOG_FORMAT` | `text` | `json` or `text` |
| `CORS_ORIGIN` | `*` | Allowed CORS origin |
| `FRIEND_CODE` | *(empty)* | Secret code for extended TTL |
| `DEFAULT_TTL_MINUTES` | `15` | Deployment lifetime |
| `EXTENDED_TTL_MINUTES` | `60` | Extended lifetime with friend code |

### Frontend

```bash
cd corvus-frontend
npm install
npm run dev
```

Create `.env.development`:

```
VITE_API_BASE_URL=http://localhost:8080
```

---

## Deployment State Machine

```
[create] --> [deploying] --> [live] --> [expired / deleted]
                  |
                  +--> [failed]

[live] --> [redeploy] --> [deploying] --> [live]
                               |
                               +--> [failed]
```

The backend manages all state transitions. The expiration cleanup loop runs every 30 seconds, queries for deployments past their TTL, and runs the full teardown (stop container, remove files, remove log, delete DB row).

---

## Roadmap

Corvus v1 is a working proof-of-concept that proves the pipeline works end-to-end: from source input to live public URL, with full lifecycle management and automatic cleanup. The goal is to evolve it into a general-purpose, open-source PaaS engine that can be used from a single home lab VM to a multi-node production cluster.

**Near-term:**
- **GitHub webhook auto-deploy.** The `webhook_secret` and `auto_deploy` fields already exist in the database schema and are generated/stored on every deployment. The missing piece is the webhook handler endpoint with HMAC-SHA256 signature verification and async redeploy trigger.
- **Log streaming.** Build logs are already written to per-deployment files on disk. Adding a `GET /api/deployments/:id/logs` SSE endpoint is a read path, not a pipeline change.
- **Repository split.** The frontend and backend will move to separate repositories, with the backend becoming a standalone engine and the frontend becoming one possible UI for it.

**Medium-term:**
- **Configurable build images.** Replace the hardcoded `node:20-alpine` with user-selectable runtimes: Python, Go, Rust, Bun, or any custom Docker image. The ephemeral build container logic is already image-agnostic at the SDK level.
- **Dynamic app hosting.** Serve Node, Python, Go, and other server-side applications, not just static sites. This requires a port allocation strategy and health check system, but the container orchestration layer already supports it.
- **Private GitHub repos.** OAuth flow for access tokens, used in git clone via HTTPS auth.
- **Custom domains.** Updating Traefik labels to route custom hostnames. User sets a DNS CNAME.

**Long-term:**
- **Multi-user auth.** JWT-based authentication, per-user deployment isolation.
- **Multi-node orchestration.** Container placement, cross-node networking, shared storage. The path from single-VM to distributed PaaS.
- **Run-anywhere packaging.** Single binary or Docker Compose bundle that works on any Linux server, cloud VM, or home lab with Docker installed.

The v1 is a single-node system. Everything after that is about removing the single-node assumption and turning this into software that can run at any scale.

---

## License

[MIT](./LICENSE)
