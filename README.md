# Orchestra 🎻

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/go-1.22+-00ADD8.svg?logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/next.js-14+-black.svg?logo=next.js&logoColor=white)
![Kubernetes](https://img.shields.io/badge/kubernetes-k3s-326ce5.svg?logo=kubernetes&logoColor=white)

> **Orchestra** is an internal PaaS for managing physical servers. It eliminates "snowflake servers" by automating provisioning, cluster orchestration, and application deployment on bare metal.

## 🚀 Features

- **Inventory Management**: Auto-discovery and pre-flight checks (CPU, RAM, OS, Cgroups) via SSH.
- **Cluster Topology**: Visual designer to designate Manager/Worker nodes and form Kubernetes clusters (K3s).
- **Application Deployment**: GitOps-style deployment engine using standard Kubernetes manifests.
- **Zero-Agent Architecture**: Uses `crypto/ssh` for bootstrapping; no permanent agent required on nodes.
- **Real-time Observability**: Live logs and metrics streaming.

## 🛠️ Architecture

Orchestra consists of a **100% Go backend** and a **Next.js frontend**.

```mermaid
graph TD
    Client[Web Browser] -->|HTTP/WebSocket| Frontend[Next.js App]
    Frontend -->|REST API| API[Go API Server]
    API -->|Persist| DB[(PostgreSQL)]
    API -->|Queue Task| Redis[(Redis)]
    Worker[Go Worker] -->|Consume Task| Redis
    Worker -->|SSH| Server1[Physical Server 1]
    Worker -->|SSH| Server2[Physical Server 2]
    Worker -->|K8s API| Cluster[K3s Cluster]
```

## 📦 Quick Start

### Prerequisites
- Docker & Docker Compose
- `make` (optional)

### Running with Docker

1. **Clone the repository:**
   ```bash
   git clone https://github.com/enochcodes/orchestra.git
   cd orchestra
   ```

2. **Start the full stack** (PostgreSQL, Redis, API, Worker, Frontend):
   ```bash
   make up
   # OR
   docker-compose up -d
   ```

3. **Access the dashboard:**
   Open [http://localhost:3000](http://localhost:3000) in your browser.

   - **API**: `http://localhost:8080`
   - **Database**: `postgres://orchestra:orchestra_password@localhost:5432/orchestra`
   - **Redis**: `localhost:6379`

4. **Login** (system admin is seeded on first run):
   - **Email**: `admin@orchestra.local`
   - **Password**: `admin123`

5. **Stop the stack:**
   ```bash
   make down
   ```

   To also remove volumes (database data):
   ```bash
   make clean
   ```

### Optional: Production env vars

Create a `.env` file for production use:

```env
ENCRYPTION_KEY=<32-byte-hex>  # openssl rand -hex 32
JWT_SECRET=<your-secret>
SKIP_AUTH=false
```

## 🔧 Development

### Backend (Go)

Run just the DB and Redis from Docker, then run the backend locally:

```bash
# Start DB and Redis only
make dev-deps
# OR: docker-compose up -d db redis

# Copy env and run backend (uses localhost to connect to Docker DB)
cd backend
cp .env.example .env
# Edit .env: set ENCRYPTION_KEY (openssl rand -hex 32)
go run cmd/api/main.go
# In another terminal:
go run cmd/worker/main.go
```

Or run the full stack in Docker:
```bash
docker-compose up -d
```

### Frontend (Next.js)

```bash
cd frontend
cp .env.example .env   # NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
npm install
npm run dev
```

### E2E Testing (Playwright)

Ensure the API is running, then:

```bash
cd frontend
npx playwright install   # First time only
npm run test:e2e
```

See [frontend/e2e/README.md](frontend/e2e/README.md) for details.

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
