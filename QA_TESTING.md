# Orchestra v1 — QA & User Acceptance Testing Guide

This document provides a structured test plan for validating Orchestra v1 against real servers, clusters, and application deployments.

## Prerequisites

- [ ] Docker & Docker Compose v2 installed
- [ ] At least one Linux server with SSH access (root or sudo user)
- [ ] Valid SSH private key (PEM format)
- [ ] Git repository with a simple app (or public Docker image)

## 1. Environment Setup

```bash
# Clone and start the stack
git clone https://github.com/enochcodes/orchestra.git
cd orchestra

# Create .env for production-like config
cp .env.docker.example .env
# Edit .env: set ENCRYPTION_KEY (openssl rand -hex 32), JWT_SECRET, SKIP_AUTH=false

docker compose up -d
```

**Verify:**
- [ ] `http://localhost:3000` loads the UI
- [ ] `http://localhost:8080/health` returns `{"status":"healthy"}`
- [ ] `http://localhost:8080/ready` returns `{"status":"ready","database":"ok"}`

## 2. Authentication

| Test | Steps | Expected |
|------|-------|----------|
| Login (default) | Go to /login, use admin@orchestra.local / admin123 | Redirect to dashboard |
| Invalid credentials | Use wrong email/password | Error message shown |
| Logout | Click user menu → Sign out (if implemented) | Redirect to login |
| Protected route | Visit /servers without token | Redirect to login |

## 3. Server Registration (Real Server)

| Test | Steps | Expected |
|------|-------|----------|
| Register server | Servers → Add Server → Fill IP, SSH user, paste SSH key | "Server registered, pre-flight check queued" |
| Preflight completes | Wait 30–60s, refresh | Server status: Ready, CPU/RAM/OS populated |
| View server | Click View on a server | Dialog shows specs, preflight report, Nginx configs |
| Delete server | View → Delete Server | Server removed from list |

**Edge cases:**
- [ ] Invalid IP format → Error
- [ ] Wrong SSH key → Preflight fails, status: Error
- [ ] Unreachable host → Preflight fails with timeout

## 4. Cluster Design

| Test | Steps | Expected |
|------|-------|----------|
| Create K8s cluster | Clusters → New Cluster → Select type: K8s, manager + workers | "Cluster design accepted, provisioning started" |
| Cluster becomes active | Wait 2–5 min | Status: Active |
| Create Swarm cluster | Same flow, type: Docker Swarm | Swarm init + join tasks run |
| Create Manual cluster | Same flow, type: Manual | Docker installed on nodes, cluster active |

**Validation:**
- [ ] Manager must be Ready before design
- [ ] Workers must be Ready (or idle) before design
- [ ] Cluster name required

## 5. Application Deployment

### 5a. Git Source

| Test | Steps | Expected |
|------|-------|----------|
| Deploy from Git | Applications → New Deployment → Source: Git, repo URL, branch | Deployment queued, status: building → deploying → running |
| View deployment logs | Deployments page | Logs show clone, build, deploy steps |
| Redeploy | Applications → Manage → Redeploy | New deployment created |

### 5b. Docker Image Source

| Test | Steps | Expected |
|------|-------|----------|
| Deploy public image | Source: Docker Image, e.g. `nginx:alpine` | Pull + deploy, status: running |
| Port mapping | Set Port: 80 | Container exposes port 80 |

### 5c. Manual Source

| Test | Steps | Expected |
|------|-------|----------|
| Deploy from path | Source: Manual, path e.g. `/opt/myapp` | Uses existing path, builds if Dockerfile present |

**Validation:**
- [ ] cluster_id required
- [ ] repo_url required for git
- [ ] docker_image required for docker_image
- [ ] manual_path required for manual
- [ ] Cluster must be Active

## 6. Environment Variables

| Test | Steps | Expected |
|------|-------|----------|
| Create environment | Environments → New → Select cluster, scope, add KEY=VALUE | Environment created |
| Push to servers | Environments → Push | Task queued, vars pushed to cluster servers |
| Edit environment | Update variables, save | Changes persisted |

## 7. Nginx Provisioning

| Test | Steps | Expected |
|------|-------|----------|
| Add Nginx config | Servers → View → Add Nginx → Domain, upstream port | Config created, Nginx provisioned on server |
| Delete Nginx | View → Delete on config | Config removed |

## 8. End-to-End Flow (Full UAT)

1. [ ] Start stack
2. [ ] Login
3. [ ] Register 2 servers (1 manager, 1 worker)
4. [ ] Wait for preflight (both Ready)
5. [ ] Design K8s cluster with both servers
6. [ ] Wait for cluster Active
7. [ ] Deploy app from Git (e.g. `https://github.com/your/simple-node-app`)
8. [ ] Verify deployment in Deployments page
9. [ ] Create environment, push to cluster
10. [ ] Add Nginx config for app domain
11. [ ] Redeploy app
12. [ ] Delete app, cluster, servers (cleanup)

## 9. E2E Automated Tests (Playwright)

```bash
# Ensure API + UI are running (e.g. docker compose up -d)
cd ui
npx playwright install
npm run test:e2e
```

**Coverage:**
- Auth: login, invalid credentials, protected routes
- Navigation: Servers, Clusters, Applications, Deployments
- (Optional) Server list, Cluster list, Application list (requires seeded data)

## 10. Known Limitations (v1)

- Host key verification is disabled (InsecureIgnoreHostKey)
- No RBAC beyond system admin
- No WebSocket for real-time deployment logs
- Manual source requires path to exist on manager server
- Let's Encrypt in Nginx may need DNS validation

## 11. Troubleshooting

| Issue | Check |
|-------|------|
| Preflight fails | SSH key format, firewall (port 22), user permissions |
| K3s install fails | Curl access, disk space, kernel modules |
| Deploy fails | Repo accessible from manager, Docker installed, build logs |
| 401 on API | Token expired, SKIP_AUTH setting |
