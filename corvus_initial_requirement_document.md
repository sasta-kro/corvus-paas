
# PAAS PROJECT — COMPREHENSIVE REQUIREMENTS & IMPLEMENTATION PLAN (draft 1)


## PROJECT IDENTITY & PURPOSE
Name: (working title) "LaunchPad" or similar — a self-hosted, production-grade Platform-as-a-Service for static and dynamic web app deployment. Primary purpose: portfolio/resume/GitHub/LinkedIn demonstration piece to signal platform engineering capability to employers. Secondary purpose: a genuinely functional proof-of-concept that mirrors Netlify's core deployment loop. The software should be architected and coded as if it were going to production, even if it runs locally or on a tailnet during development. Code quality, structure, and architectural decisions matter as much as feature completeness, because employers will read the source.

---

### SCOPE DEFINITION

In scope for v1 (MVP - minimal viable produt):
- User uploads a pre-built static site (zip file or build folder) via a web UI (drag and drop) or API call.
- Backend extracts the archive, places files into a container, serves them publicly.
- User provides a GitHub repo URL (public); backend clones it, runs a user-specified build command (e.g. `npm run build`), then serves the output folder.
- Toggle to enable "rebuild on push to main" via GitHub webhooks.
- User-configurable options per deployment: build command, output directory, environment variables (non-secret for now, secrets handling is a v2 concern).
- Each deployment gets a unique public URL. Local dev: `<slug>.localhost` via Traefik. Production-ready path: Cloudflare Tunnel exposure so URLs are publicly reachable without port-forwarding.
- Basic deployment status feedback: deploying, live, failed.
- Re-deploy / rebuild trigger from UI.
- Delete a deployment (stops and removes container, cleans up files).

Out of scope for v1 but architecturally planned for:
- Logs streaming to UI (v2).
- Metrics and analytics (v2).
- Secret/env var encryption (v2).
- Private GitHub repo support via OAuth (v2).
- Custom domains (v2).
- Multi-user auth (v2).
- Dynamic (non-static) app hosting, i.e. Node/Python/Go servers (v2, requires port management strategy).

---

### HIGH-LEVEL ARCHITECTURE

```
User Browser
    │
    │ (1) Loads dashboard UI (static files)
    ▼
Netlify/Vercel CDN  ──── serves React/TS (the frontend hosted here)
    │
    │ (2) fetch() API calls (JSON over HTTPS)
    ▼
Cloudflare Tunnel (cloudflared)
    │
    │ (3) forwards to localhost:8080 on the VM
    ▼
Go Backend (Control Plane) :8080
    │
    ├── (4a) Talks to Docker daemon via /var/run/docker.sock
    │         └── Spawns per-deployment Nginx containers
    │
    ├── (4b) Talks to Traefik via Docker labels on containers
    │         └── Traefik auto-routes <slug>.mydomain.com → container
    │
    ├── (4c) Clones GitHub repos, runs build commands
    │         └── Uses exec.Command / Docker build containers
    │
    └── (4d) Reads/writes SQLite/Posgres DB (deployment metadata)

Traefik Reverse Proxy :80/:443
    │
    └── Routes based on Host header to correct container
```

Everything runs on Fedora home lab VM (inside a dedicated VM, not bare metal host, consistent with the established pattern). The VM is `paas-node`. Go backend, Traefik, and all user app containers are sibling Docker containers inside that VM. Docker socket is mounted into the Go backend container so it can spawn siblings.

---

## COMPONENT BREAKDOWN

### Component 1: Go Backend (The Control Plane)
Language: Go. This is the brain. It exposes a REST API that the frontend calls. It has no HTML rendering — pure JSON API server.

Responsibilities:
- Accept and validate all API requests.
- Manage deployment lifecycle (create, build, start, stop, delete).
- Interface with Docker daemon via Docker SDK.
- Clone GitHub repos and run build commands in ephemeral build containers.
- Manage SQLite database of deployments.
- Listen for incoming GitHub webhook POST requests.
- Generate unique slugs for deployment URLs.
- Write static files to a named volume or bind-mounted directory that the Nginx container will serve.

Key packages:
- `net/http` — HTTP server and routing (or `chi` router for cleaner route grouping).
- `github.com/docker/docker/client` — Docker SDK.
- `database/sql` + `github.com/mattn/go-sqlite3` — SQLite persistence.
- `os/exec` or Docker SDK for running build commands.
- `archive/zip` — extracting uploaded zip files.
- `encoding/json` — request/response marshalling.
- `github.com/google/uuid` — slug/ID generation.
- `crypto/hmac` + `crypto/sha256` — GitHub webhook signature verification.

### Component 2: Traefik (The Router)
Runs as a Docker container on `paas-node`. Listens on ports 80 (http) and 443 (https). Configured via Docker provider, it watches the Docker socket and reads labels on containers. When the Go backend spawns a new Nginx container with the correct Traefik labels, Traefik automatically adds a routing rule for that container's subdomain. No config file reload needed. This is the "Netlify magic", instant live URL the moment the container starts.

Traefik static config (`traefik.yml`):
```yaml
entryPoints:
  web:
    address: ":80"
providers:
  docker:
    exposedByDefault: false
api:
  dashboard: true
  insecure: true  # dev only, lock down for prod
```

Each user app container gets these Docker labels applied programmatically by Go:
```
traefik.enable=true
traefik.http.routers.<slug>.rule=Host(`<slug>.localhost`)
traefik.http.services.<slug>.loadbalancer.server.port=80
```


### Component 3: Per-Deployment Nginx Containers (The Web Servers)
Each deployed site runs in its own Nginx container. The container is minimal: `nginx:alpine`. The static files are placed into a bind-mounted directory on the VM that maps to `/usr/share/nginx/html` inside the container. The ==container has no internet access (no need),== sits on the same Docker network as Traefik so Traefik can proxy to it.

Container naming convention: `deploy-<slug>` for easy identification.

### Component 4: Build System (Ephemeral Build Containers)
For GitHub-sourced deployments (not zip upload), the build process needs:
1. Clone the repo.
2. Run `npm install && npm run build` (or whatever the user specified).
3. Copy the output directory out.
4. Start serving it.

This should **NOT** run in the Go backend process itself (security and isolation). Instead, spawn an ephemeral Docker container (e.g. `node:20-alpine`) that:
- Has the cloned repo bind-mounted into it.
- Runs the build command.
- Exits after completion.

Go backend then copies the output directory to the serving location and starts the Nginx container.

Alternatively for v1 simplicity (**not recommended**): run `git clone` and build commands directly on the VM via `exec.Command` in a sandboxed temp directory, then clean up. This is simpler but less isolated. Note the tradeoff: using exec.Command runs build tools on the VM itself, which means you need Node/npm etc. installed on the VM or you use Docker containers for the build step. Recommend Docker-based builds for correctness and isolation, even in v1.

### Component 5: SQLite Database (State)
Single file: `deployments.db`. Schema:

```sql
CREATE TABLE deployments (
    id          TEXT PRIMARY KEY,       -- UUID
    slug        TEXT UNIQUE NOT NULL,   -- URL slug, e.g. "graceful-fox"
    name        TEXT NOT NULL,          -- User-given name
    source_type TEXT NOT NULL,          -- "zip" | "github"
    github_url  TEXT,                   -- e.g. "https://github.com/user/repo"
    branch      TEXT DEFAULT 'main',
    build_cmd   TEXT,                   -- e.g. "npm run build"
    output_dir  TEXT DEFAULT 'dist',    -- e.g. "dist", "build", "out"
    env_vars    TEXT,                   -- JSON-encoded key-value map
    status      TEXT NOT NULL,          -- "deploying" | "live" | "failed"
    url         TEXT,                   -- e.g. "http://graceful-fox.localhost"
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    webhook_secret TEXT,                -- HMAC secret for GitHub webhook
    auto_deploy INTEGER DEFAULT 0       -- 0 = false, 1 = true (boolean)
);
```


### Component 6: React/TypeScript Frontend (The Dashboard)
Hosted on Netlify/Vercel (static). Communicates with Go backend via fetch(). This is intentionally simple but sleek and pretty — it is not the impressive part of the project. Keep it functional, clean, and minimal. Use React as the front end framework and component framework (shadcn/ui or Radix + Tailwind) to look professional without over-engineering the frontend.

Pages:
- `/` — Dashboard: list of all deployments with status badges, URLs, last deployed time, quick actions (redeploy, delete).
- `/new` — New Deployment: choose source (zip upload or GitHub URL), fill in build config, submit.
- `/deployments/:id` — Deployment Detail: status, URL, config, rebuild button, later: logs.

Key API calls:
- `GET /api/deployments` — list all.
- `POST /api/deployments` — create new (multipart for zip, JSON for GitHub).
- `GET /api/deployments/:id` — get one.
- `DELETE /api/deployments/:id` — delete.
- `POST /api/deployments/:id/redeploy` — trigger rebuild.
- `POST /api/webhooks/github/:id` — GitHub webhook endpoint.

### Component 7: Cloudflare Tunnel (Public Exposure)
`cloudflared` runs as a process (or Docker container) on `paas-node`. Creates an outbound persistent connection to Cloudflare edge. Cloudflare gives a public HTTPS URL (or configure a custom domain if there is have one). All traffic to that URL is forwarded to the Go backend at `localhost:8080`. This is how the React frontend (hosted on Netlify) can call the backend API from the public internet, and how GitHub can POST webhook events to the server.

For wildcard subdomains (so `<slug>.yourdomain.com` works), need:
- A real domain (can get free: `eu.org`, or pay ~$10/yr for a real one).
- Cloudflare DNS with a wildcard CNAME.
- Cloudflare Tunnel configured to route `*.yourdomain.com` to Traefik on the VM at port 80.
- Traefik then does the final routing based on Host header to the correct Nginx container.

For local/tailnet dev without a domain: use `<slug>.localhost` and access via the tailnet. This is fine for development and demo purposes.

---

## API SPECIFICATION

`POST /api/deployments` — Create Deployment

For zip upload:
```
Content-Type: multipart/form-data
Fields:
  name: string
  source_type: "zip"
  file: <zip file>
  build_cmd: string (optional, run before serving — if zip is pre-built, leave empty)
  output_dir: string (default: ".")
  env_vars: JSON string (optional)
```

For GitHub:
```
Content-Type: application/json
{
  "name": "my-react-app",
  "source_type": "github",
  "github_url": "https://github.com/user/repo",
  "branch": "main",
  "build_cmd": "npm ci && npm run build",
  "output_dir": "dist",
  "env_vars": { "NODE_ENV": "production" },
  "auto_deploy": true
}
```

Response:
```json
{
  "id": "uuid",
  "slug": "graceful-fox",
  "status": "deploying",
  "url": "http://graceful-fox.localhost",
  "webhook_url": "https://your-tunnel.com/api/webhooks/github/uuid",
  "webhook_secret": "abc123"
}
```

`POST /api/webhooks/github/:id` — GitHub Webhook Handler

Headers checked: `X-GitHub-Event`, `X-Hub-Signature-256`. Backend verifies HMAC with stored `webhook_secret`. If event is `push` and branch matches configured branch and `auto_deploy` is true, triggers rebuild.

`POST /api/deployments/:id/redeploy` — Redeploy

Re-runs the full build+deploy pipeline for that deployment. Returns 202 Accepted immediately (async operation). Status transitions to "deploying" then "live" or "failed".

---

## DEPLOYMENT LIFECYCLE (state machine)

```
[created] → [deploying] → [live]
                       ↘ [failed]

[live] → [deploying] → [live]     (on redeploy)
                    ↘ [failed]

[live / failed] → [deleted]        (on delete — container stopped, files removed, DB row deleted)
```

The Go backend manages all state transitions. Status is stored in SQLite/Postgres and returned to the frontend on polling or on API calls. For v1, the frontend polls `GET /api/deployments/:id` every few seconds while status is "deploying". For v2, replace with WebSocket or Server-Sent Events for real-time log streaming.

---

### BUILD PIPELINE LOGIC (Go pseudocode structure)

```
func deploy(deployment Deployment):
    setStatus(deployment.id, "deploying")
    
    workDir = createTempDir("/tmp/builds/<id>")
    
    if deployment.source_type == "zip":
        extractZip(uploadedFile, workDir)
        if deployment.build_cmd != "":
            runBuildInContainer(workDir, deployment.build_cmd, deployment.env_vars)
        serveDir = workDir + "/" + deployment.output_dir
    
    elif deployment.source_type == "github":
        gitClone(deployment.github_url, deployment.branch, workDir)
        if deployment.build_cmd != "":
            runBuildInContainer(workDir, deployment.build_cmd, deployment.env_vars)
        serveDir = workDir + "/" + deployment.output_dir
    
    destDir = "/srv/deployments/<slug>"
    copyDir(serveDir, destDir)
    
    stopAndRemoveExistingContainer("deploy-<slug>")  // for redeployments
    
    startNginxContainer(
        name: "deploy-<slug>",
        bindMount: destDir → /usr/share/nginx/html,
        labels: traefikLabels(slug),
        network: "paas-network"
    )
    
    setStatus(deployment.id, "live")
    cleanup(workDir)

func runBuildInContainer(workDir, buildCmd, envVars):
    // Pull node:20-alpine if not present
    // Create container with workDir bind-mounted
    // Run buildCmd inside container
    // Wait for exit, check exit code
    // Remove ephemeral build container
```

Error handling: any step failure → setStatus("failed") + log the error. Build container stdout/stderr captured and stored to a log file on disk for v2 log streaming feature. Even in v1, write the logs to disk so they are not getting thrown away, they will be needed for v2 without rewriting the build system.

---

### DIRECTORY STRUCTURE ON THE VM

```
/srv/paas/
├── deployments/
│   ├── graceful-fox/          ← bind-mounted into Nginx container
│   │   ├── index.html
│   │   └── assets/
│   └── happy-river/
│       └── index.html
├── logs/
│   ├── graceful-fox.log       ← captured build/deploy logs
│   └── happy-river.log
└── db/
    └── deployments.db
```

```
~/paas-control-plane/          ← Go project root
├── main.go
├── go.mod
├── go.sum
├── handlers/
│   ├── deployments.go
│   ├── webhooks.go
│   └── health.go
├── docker/
│   ├── client.go              ← Docker SDK wrapper
│   ├── builder.go             ← build container logic
│   └── nginx.go               ← Nginx container spawn logic
├── db/
│   ├── db.go                  ← SQLite connection and migrations
│   └── deployments.go         ← CRUD queries
├── build/
│   ├── pipeline.go            ← orchestrates the full deploy pipeline
│   ├── git.go                 ← git clone logic
│   └── zip.go                 ← zip extraction logic
├── models/
│   └── deployment.go          ← Deployment struct
├── config/
│   └── config.go              ← reads env vars / config file
└── docker-compose.yml         ← runs Traefik + Go backend together
```

```
~/paas-frontend/               ← React/TS project root
├── src/
│   ├── pages/
│   │   ├── Dashboard.tsx
│   │   ├── NewDeployment.tsx
│   │   └── DeploymentDetail.tsx
│   ├── components/
│   │   ├── DeploymentCard.tsx
│   │   ├── StatusBadge.tsx
│   │   └── DeploymentForm.tsx
│   ├── api/
│   │   └── client.ts          ← typed fetch wrappers
│   └── types/
│       └── deployment.ts      ← TypeScript interfaces matching Go structs
├── package.json
└── vite.config.ts
```

---

### DOCKER COMPOSE FOR THE PLATFORM ITSELF

```yaml
# ~/paas-control-plane/docker-compose.yml
services:

  traefik:
    image: traefik:v3.0
    container_name: paas-traefik
    ports:
      - "80:80"
      - "8081:8080"   # Traefik dashboard (dev only)
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik.yml:/etc/traefik/traefik.yml:ro
    networks:
      - paas-network
    restart: unless-stopped

  backend:
    build: .
    container_name: paas-backend
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock   # so Go can control Docker
      - /srv/paas:/srv/paas                          # shared volume for static files
    environment:
      - DB_PATH=/srv/paas/db/deployments.db
      - SERVE_ROOT=/srv/paas/deployments
      - LOG_ROOT=/srv/paas/logs
      - TRAEFIK_NETWORK=paas-network
    networks:
      - paas-network
    restart: unless-stopped
    depends_on:
      - traefik

networks:
  paas-network:
    name: paas-network
    external: false
```

Note on docker.sock: the Go backend container mounts the Docker socket. This gives it root-equivalent power over the VM's Docker daemon. This is intentional and standard for PaaS control planes. This is why the entire thing is run inside a VM (`paas-node`) and not on the bare Fedora host — the blast radius is contained.

---

### GITHUB WEBHOOK FLOW

1. User creates a deployment with `auto_deploy: true` and gets back a `webhook_url` and `webhook_secret`.
2. User goes to their GitHub repo settings → Webhooks → Add webhook.
3. Payload URL: the `webhook_url` from the API response (e.g. `https://your-tunnel.com/api/webhooks/github/<deployment-id>`).
4. Secret: the `webhook_secret`.
5. Content type: `application/json`. Events: "Just the push event."
6. When user pushes to main, GitHub POSTs to your backend.
7. The backend:
   a. Reads `X-Hub-Signature-256` header.
   b. Computes HMAC-SHA256 of the raw request body using the stored `webhook_secret`.
   c. Compares — if mismatch, return 401 (reject tampered/spoofed requests).
   d. Parses JSON body, checks `ref` field (e.g. `refs/heads/main`).
   e. If branch matches and `auto_deploy` is true → trigger redeploy pipeline asynchronously.
   f. Return 200 immediately (GitHub requires fast response or it marks webhook as failed).

---

## IMPLEMENTATION PHASES (ordered)

**Phase 0: Infrastructure Setup (Day 1)**
- Create `paas-node` VM from the existing Fedora Server template.
- Install Docker, set up Docker socket permissions.
- Write `docker-compose.yml` for Traefik + stub backend.
- Start Traefik, verify dashboard at `:8081`.
- Manually start an Nginx container with Traefik labels and verify routing to `test.localhost`.
- Win condition: browser hits `http://test.localhost` and sees Nginx default page.

**Phase 1: Go Backend Skeleton (Day 2-3)**
- Initialize Go module. Set up chi router.
- Implement health endpoint `GET /health → 200 OK`.
- Implement SQLite/Postgres setup: open DB, run schema migration on startup.
- Implement `GET /api/deployments` → returns empty array.
- Implement `POST /api/deployments` with JSON body (GitHub source type only for now) → inserts row, returns created deployment.
- Containerize the Go backend via Dockerfile, add to docker-compose.
- Win condition: curl hits the API, rows appear in SQLite.

**Phase 2: Docker Integration (Day 3-4)**
- Write Docker SDK wrapper in `docker/client.go`.
- Implement function to start an Nginx container with bind mount + Traefik labels given a slug and a directory path.
- Implement function to stop and remove a container by name.
- Test: call these functions from a temporary main() test, verify containers appear/disappear in `docker ps` and Traefik routes appear/disappear.

**Phase 3: Zip Upload Pipeline (Day 4-5)**
- Add `POST /api/deployments` multipart handling for zip uploads.
- Implement zip extraction to temp dir.
- Copy output dir to `/srv/paas/deployments/<slug>`.
- Call Docker spawn function.
- Update DB status.
- Win condition: curl with a zip file → site is live at `<slug>.localhost`.

**Phase 4: GitHub Clone + Build Pipeline (Day 5-7)**
- Implement `build/git.go`: exec.Command wrapping `git clone`.
- Implement `build/builder.go`: spawn ephemeral Node container, run build command inside it using Docker SDK exec (ContainerExecCreate/Start), wait for completion.
- Wire into deploy pipeline for `source_type: "github"`.
- Win condition: submit a GitHub URL + `npm run build` + output dir → site is built and live.

**Phase 5: Redeploy + Delete (Day 7-8)**
- `POST /api/deployments/:id/redeploy` → re-runs full pipeline, stops old container first.
- `DELETE /api/deployments/:id` → stop container, remove files, delete DB row.
- Status transitions: update DB at each step.

**Phase 6: GitHub Webhooks (Day 8-10)**
- Implement `POST /api/webhooks/github/:id`.
- HMAC-SHA256 verification.
- Branch matching.
- Async redeploy trigger (goroutine).
- Win condition: push to GitHub → site auto-rebuilds within seconds.

**Phase 7: React Frontend (Day 10-14)**
- Scaffold with Vite + TypeScript + Tailwind + shadcn/ui.
- Implement Dashboard page.
- Implement New Deployment form (zip upload + GitHub tabs).
- Implement Deployment Detail page with status polling.
- Deploy to Netlify/Vercel.

**Phase 8: Cloudflare Tunnel + Public Exposure (Day 14-15)**
- Install cloudflared on `paas-node`.
- Configure tunnel to route to Go backend `:8080`.
- Configure wildcard routing through tunnel to Traefik for `*.yourdomain.com`.
- Update React frontend API URL to tunnel URL.
- Win condition: end-to-end demo accessible from public internet.

**Phase 9: Polish for Portfolio (Day 15-20)**
- Write a detailed `README.md`: architecture diagram, setup instructions, feature list, tech stack justification, screenshots/GIFs of the deploy flow.
- Add `ARCHITECTURE.md` with the mermaid diagram and component explanations.
- Clean up code: proper error types, structured logging (use `log/slog` in Go), consistent error responses.
- Add basic request logging middleware.
- Ensure all Docker containers are named consistently, all temp dirs are cleaned up on failure.
- Record a short demo video: submit GitHub URL → build logs → site live at public URL. This is what goes on LinkedIn.

---

**FUTURE FEATURES (ARCHITECTURALLY PLANNED, NOT BUILT IN V1)**

Log streaming: the build pipeline already writes logs to `/srv/paas/logs/<slug>.log`. In v2, expose `GET /api/deployments/:id/logs` that streams the file via Server-Sent Events or chunked transfer. Frontend subscribes to the SSE stream and renders log lines in real-time. This is the most impactful visual feature for a portfolio demo.

Metrics: track deploy count, build duration, site request count (have Traefik emit metrics to Prometheus, scrape with a sidecar). In v2, expose a `/api/deployments/:id/metrics` endpoint.

Secret env vars: for v1, env vars are stored as plaintext JSON in SQLite (acceptable for a portfolio/PoC). For v2, encrypt with AES-GCM using a master key loaded from an env var at startup. Never log env var values.

Private GitHub repos: requires OAuth flow. In v2, implement GitHub OAuth App flow, store access token, use it for git clone via HTTPS token auth.

Dynamic app hosting (non-static): requires port allocation strategy. Each dynamic app needs a unique port inside its container exposed to Traefik. Maintain a port pool in the DB (e.g. 30000-32000), allocate on spawn, release on delete. Much more complex than static — save for v2.

Multi-user auth: add JWT-based auth to the Go API. Each user only sees their own deployments. Store user_id on deployment rows.

Custom domains: user provides a custom domain; backend updates Traefik router label `rule=Host('custom.com')` on the container; user sets DNS CNAME to your Cloudflare tunnel URL.

---

**KEY ARCHITECTURAL DECISIONS TO DOCUMENT (FOR PORTFOLIO)**

1. **Why Docker-inside-VM, not bare metal Docker**: same reasoning as your Minecraft server. Blast radius containment. If a user's build script does `rm -rf /`, it trashes the build container, not your host.

2. **Why Traefik over Nginx as reverse proxy**: Nginx requires config file reload to add routes. Traefik watches Docker socket and updates routing table instantly when containers are created/destroyed. This is what enables the "instant live URL" experience that defines Netlify's feel.

3. **Why SQLite over Postgres**: for a single-node PoC, SQLite is perfectly appropriate. No operational overhead. Go's standard `database/sql` + `go-sqlite3` is battle-tested. Document that swapping to Postgres requires only changing the driver and DSN — the query layer is unchanged.

4. **Why Go over Python/Kotlin for the control plane**: Go's Docker SDK is native (Docker itself is written in Go). Goroutines are the natural fit for concurrent deployment pipelines (each deploy is a goroutine). Static binary compilation means the backend container needs no runtime. These are defensible, specific, technically correct reasons — say them in interviews.

5. **Why Cloudflare Tunnel over port forwarding**: zero router configuration, no exposed home IP, Cloudflare handles TLS termination, DDoS protection. Architecturally identical to what small startups do before they can afford cloud VMs.

6. **Why async deploy with status polling (not sync HTTP)**: a deploy could take 30-120 seconds (npm install is slow). Keeping an HTTP connection open that long is unreliable and couples frontend to backend timing. Accept the request, return 202 with the deployment ID, let the client poll status. In v2, replace polling with SSE for a better UX without fundamentally changing the backend model.

---

**RESUME/PORTFOLIO FRAMING**

When describing this project to employers, the framing is:

> "Built a production-grade PaaS platform in Go that automates the full deployment lifecycle — from GitHub webhook to live URL — using Docker SDK for container orchestration, Traefik for dynamic reverse proxying, SQLite for state management, and Cloudflare Tunnel for zero-config public exposure. Architected for extensibility: static site hosting in v1, with the build pipeline, log capture, and port management designed to support dynamic apps and multi-user auth in subsequent versions."

Specific talking points that signal seniority:
- HMAC-SHA256 webhook verification (security awareness).
- Ephemeral build containers (isolation and reproducibility).
- docker.sock mounting with blast-radius awareness (why VM, not host).
- Traefik label-based routing (dynamic config, not static files).
- Async pipeline with state machine (resilience, not naive sync HTTP).
- Volume architecture: shared `/srv/paas` between Go backend and Nginx containers (cross-container file sharing via host bind mounts, not Docker-managed volumes, because Go needs to write files that Nginx containers read).