# E2E Testing (Playwright)

## Prerequisites

1. **Backend API running** – E2E tests require the Orchestra API to be available:
   ```bash
   # From project root - start db, redis, and api
   make dev-deps && cd backend && go run cmd/api/main.go
   # Or: docker-compose up -d db redis api
   ```

2. **Frontend .env** – Ensure `NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1` in `.env`

## Run Tests

```bash
# Install Playwright browsers (first time only)
npx playwright install

# Run all E2E tests (starts dev server automatically)
npm run test:e2e

# Run with UI mode
npm run test:e2e:ui
```

## Environment

- `PLAYWRIGHT_BASE_URL` – App URL (default: http://localhost:3000)
- `PLAYWRIGHT_API_URL` – API URL for reference (default: http://localhost:8080)
