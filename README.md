# Corvus PaaS

> **Currently being built.** This is not ready for use yet.

Corvus is a self-hosted PaaS I'm building from scratch. The idea is simple. Give it a zip file or a GitHub repo, and it deploys the site to a live public URL. Basically a simplified Netlify/Vercel. 


I wanted to build something that touches real infrastructure, not just another CRUD app. Container orchestration, reverse proxy routing, build pipelines, webhooks. The kind of stuff that actually runs behind the platforms we use every day.

## What it will do

- Accept a **zip upload** or **GitHub repo URL** and deploy it as a live site
- Run user-defined build commands (`npm run build`, etc.) inside isolated containers
- Assign each deployment a unique public URL, routed automatically
- Listen for GitHub webhooks to **auto-redeploy on push**
- Provide a clean web dashboard to create, monitor, and manage deployments

## How it works (how it will work) - High level overview

A Go backend acts as the control plane. It talks to the Docker daemon to spin up per-deployment Nginx containers, each serving a single site. Traefik sits in front as a reverse proxy, automatically discovering containers and routing traffic based on subdomain. The whole thing is exposed to the public internet through a Cloudflare Tunnel, so no port forwarding or static IP needed.

```
Browser → Cloudflare Tunnel → Go API → Docker → Traefik → Nginx (per site)
```

## Why I'm building this

Most portfolio projects don't go deeper than a web framework and a database. I wanted to build something that forces me to think about systems, process isolation, container lifecycle, networking, routing, and how all these pieces fit together in a real deployment pipeline. This project is that.

## License

See [LICENSE](./LICENSE).
