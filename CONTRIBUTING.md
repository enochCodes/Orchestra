# Contributing to Orchestra

First off, thanks for taking the time to contribute!

## Project Structure

```
Orchestra/
├── core/           # Go backend — API server, worker, engine
│   ├── cmd/
│   │   ├── server/     # API server entrypoint
│   │   └── worker/     # Background worker entrypoint
│   ├── internal/
│   │   ├── handler/    # HTTP handlers and routing
│   │   ├── model/      # Domain models (GORM)
│   │   ├── service/    # Business logic
│   │   ├── engine/     # Provisioning engine (SSH tasks, K8s, Swarm)
│   │   ├── buildpack/  # Dockerfile generation
│   │   ├── config/     # Configuration
│   │   └── store/      # Database connection and migrations
│   └── pkg/
│       └── ssh/        # SSH client and preflight checks
├── ui/             # Next.js frontend
│   └── src/
│       ├── app/        # Page routes
│       ├── components/ # React components
│       └── lib/        # API client, auth, utilities
├── docker-compose.yml
└── Makefile
```

## How to Contribute

### Reporting Bugs

1. **Check existing issues** to see if the bug has already been reported.
2. **Verify validity**: Ensure the issue is reproducible.
3. **Use a clear and descriptive title**.
4. **Describe exact reproduction steps**.
5. **Include screenshots** if applicable.

### Pull Requests

1. Follow Go standard style (`gofmt`) for core code
2. Follow Prettier for UI code
3. Ensure `go vet ./...` passes
4. Ensure `npm run lint` passes in `ui/`
5. Write tests for new functionality

### Git Commit Messages

- Use present tense ("Add feature" not "Added feature")
- Use imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit first line to 72 characters
- Reference issues after the first line

## Development

### Core (Go)

```bash
make dev-deps           # Start DB and Redis
cd core
cp .env.example .env    # Set ENCRYPTION_KEY
go run cmd/server/main.go
go run cmd/worker/main.go
```

### UI (Next.js)

```bash
cd ui
cp .env.example .env
npm install
npm run dev
```

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
