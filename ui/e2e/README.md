# E2E Testing (Playwright)

## Prerequisites

1. **API + DB + Redis running** — E2E tests require the Orchestra API:
   ```bash
   # From project root
   docker compose up -d
   # Or for local dev: make dev-deps, then run core server + worker
   ```

2. **UI .env** — Ensure `NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1` in `ui/.env`

## Run Tests

```bash
cd ui

# Install Playwright browsers (first time only)
npx playwright install

# Run all E2E tests (starts dev server automatically)
npm run test:e2e

# Run with UI mode
npm run test:e2e:ui
```

## Test Suites

- **auth.spec.ts** — Login, invalid credentials, protected routes
- **navigation.spec.ts** — Basic navigation after login
- **flows.spec.ts** — Full flow: Servers, Clusters, Applications, Deployments, Environments, Monitoring, Settings

## Environment

- `PLAYWRIGHT_BASE_URL` — App URL (default: http://localhost:3000)
- `CI` — When set, uses 2 retries and 1 worker
