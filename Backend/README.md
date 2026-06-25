# OMS SaaS Mini — Go Backend

Inventory & Order Management System REST API built with Go, Chi, PostgreSQL.

## Stack

- **Go 1.22+** with Chi router
- **PostgreSQL 16** via Docker Compose
- **JWT** auth with multi-tenant support
- **Stripe** for subscription billing

## Quick Start

```bash
# Start PostgreSQL
make up

# Copy env and run API
cp .env.example .env
make run
```

API runs at `http://localhost:8080`

Health check: `GET /health`

## API Endpoints

Base URL: `/api/v1`

| Module | Endpoints |
|--------|-----------|
| Auth | `POST /auth/register`, `/login`, `/refresh`, `/logout` |
| Tenant | `GET /tenant/me`, `PATCH /tenant/me` |
| Products | CRUD `/products`, `/categories` |
| Inventory | `GET /inventory`, `POST /inventory/adjust`, `/transfer` |
| Orders | CRUD `/orders`, `PATCH /orders/:id/status` |
| Customers | CRUD `/customers` |
| Reports | `GET /reports/dashboard`, `/sales`, `/inventory` |
| Billing | `GET /billing/plan`, `POST /billing/checkout` |

See [docs/openapi.yaml](docs/openapi.yaml) for API documentation.

## Environment Variables

See `.env.example`

## Project Structure

```
cmd/api/          # Entry point
internal/         # Application code
db/migrations/    # SQL migrations
docs/             # OpenAPI spec
```

## Order Status Flow

`draft` → `confirmed` → `processing` → `shipped` → `delivered`

Stock is deducted when order moves from `draft` to `confirmed`.

## Development

```bash
make dev     # Start DB + API
make build   # Build binary
make test    # Run tests
bash scripts/seed.sh  # Seed demo data
```

## Deployment

| Component | Recommended Platform |
|-----------|---------------------|
| Go API | Railway, Fly.io, or any VPS |
| PostgreSQL | Supabase, Neon, or RDS |
| Next.js | Vercel |

### Go API Deploy

1. Set environment variables from `.env.example`
2. Ensure `DATABASE_URL` points to production PostgreSQL
3. Run `go build -o api ./cmd/api` and deploy the binary
4. Migrations run automatically on startup

### Next.js Deploy

Deploy the `Frontend` folder to Vercel:

1. Set `NEXT_PUBLIC_API_URL` to your production API URL
2. Deploy to Vercel with `npm run build`
