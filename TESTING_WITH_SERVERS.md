# Testing Orchestra with Local Test Servers

This guide walks through testing Orchestra end-to-end using Docker-based test servers that act as "physical" machines. No real hardware required.

## Prerequisites

- Docker & Docker Compose v2
- `make` (optional)

## Step 1: Generate SSH Keys

Orchestra uses SSH key authentication to connect to servers. Generate a key pair first:

```bash
chmod +x scripts/generate-ssh-keys.sh
./scripts/generate-ssh-keys.sh
```

This creates:
- `keys/id_rsa` — **Private key** (paste this when registering servers in Orchestra)
- `keys/id_rsa.pub` — **Public key** (automatically added to test servers)

> **Important:** Never commit `keys/` to git. It is in `.gitignore`.

## Step 2: Start the Stack with Test Servers

```bash
# Start Orchestra + 3 test server containers
docker compose -f docker-compose.yml -f docker-compose.test-servers.yml up -d

# Or with make:
make up-test-servers
```

This starts:
- **Orchestra:** db, redis, server, worker, ui
- **Test servers:** test-server-1, test-server-2, test-server-3 (SSH on port 22)

## Step 3: Register Servers in Orchestra

1. Open [http://localhost:3000](http://localhost:3000) and log in (admin@orchestra.local / admin123)
2. Go to **Servers** → **Add Server**
3. For each test server, register:

| Field    | test-server-1 | test-server-2 | test-server-3 |
|----------|----------------|---------------|---------------|
| Hostname | test-server-1  | test-server-2 | test-server-3 |
| IP       | test-server-1  | test-server-2 | test-server-3 |
| SSH Port | 22             | 22            | 22            |
| SSH User | orchestra      | orchestra     | orchestra     |
| SSH Key  | Paste contents of `keys/id_rsa` | same | same |

> **Note:** Use the container hostname as IP — the worker runs in the same Docker network and resolves these names.

4. Click **Register** — preflight checks will run automatically (CPU, RAM, OS detection)

## Step 4: Design a Cluster

1. Go to **Clusters** → **Create Cluster**
2. Choose type: **Kubernetes (K3s)**, **Docker Swarm**, or **Manual**
3. Select **test-server-1** as manager
4. Select **test-server-2**, **test-server-3** as workers (optional)
5. Click **Create** — provisioning tasks will run in the background

## Step 5: Deploy an Application

1. Wait for the cluster to become **Active**
2. Go to **Applications** → **New Deployment**
3. Choose source: **Git** (e.g. a simple Node/Go app), **Docker Image** (e.g. `nginx:alpine`), or **Manual**
4. Select your cluster and configure
5. Click **Deploy**

## Alternative: Standalone SSH Container

For a single ad-hoc test server (e.g. on host port 2222):

```bash
# Generate keys first
./scripts/generate-ssh-keys.sh

# Create standalone container (uses keys/id_rsa.pub automatically)
chmod +x setup-ssh-container.sh
./setup-ssh-container.sh test-ssh orchestra orchestra 2222 keys/id_rsa.pub
```

Then register in Orchestra with:
- **IP:** `host.docker.internal` (Mac/Windows) or your host IP
- **Port:** 2222
- **User:** orchestra
- **Key:** contents of `keys/id_rsa`

## Troubleshooting

| Issue | Check |
|-------|-------|
| Preflight fails | Ensure keys were generated and test servers started. Verify `keys/id_rsa.pub` exists. |
| "Connection refused" | Test servers may still be starting. Wait 30s and retry. |
| "Permission denied" | Ensure you pasted the **private** key (`id_rsa`), not the public key. |
| Worker can't reach servers | Ensure you used `docker compose -f docker-compose.yml -f docker-compose.test-servers.yml` so test servers are on the same network. |

## Cleanup

```bash
# Stop everything including test servers
docker compose -f docker-compose.yml -f docker-compose.test-servers.yml down

# Remove keys (optional)
rm -rf keys/
```
