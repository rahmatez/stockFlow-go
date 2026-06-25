# StockFlow

**Inventory & Order Management System** — a SaaS platform to manage products, stock, orders, and business reports from a single dashboard.

StockFlow is built for small businesses and operations teams that need a simple yet complete back-office system. Each registered business gets its own **isolated workspace** (multi-tenant).

---

## Features

| Module | Description |
|--------|-------------|
| **Auth & Multi-tenant** | Business registration, JWT login, per-tenant data isolation |
| **Products** | Product catalog (SKU, pricing, categories, low-stock threshold) |
| **Inventory** | Per-warehouse stock, manual adjustments, inter-warehouse transfers |
| **Orders** | Order creation, status workflow, automatic stock deduction on confirmation |
| **Customers** | Customer records with search & pagination |
| **Reports** | KPI dashboard, sales charts, top-selling products |
| **Billing** | Subscription plans (Free / Starter / Pro) via Stripe |

### Order status flow

```
draft → confirmed → processing → shipped → delivered
```

Stock is deducted when an order moves from `draft` to `confirmed`.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.22, Chi, pgx, JWT, Stripe |
| **Frontend** | Next.js 16, TypeScript, Tailwind CSS v4 |
| **Database** | PostgreSQL 16 |
| **State & Forms** | TanStack Query, React Hook Form, Zod |
| **Charts** | Recharts |
| **Icons** | Phosphor Icons |

---

## Project Structure

```
OMS-SaaS-go/
├── Backend/          # Go REST API + PostgreSQL
│   ├── cmd/api/      # Entry point
│   ├── internal/     # Handlers, repositories, middleware
│   ├── db/migrations/
│   └── docs/openapi.yaml
└── Frontend/         # Next.js admin dashboard
    └── src/
        ├── app/      # Pages (App Router)
        ├── components/
        └── lib/      # API client, auth, utils
```

---

## Prerequisites

- [Go](https://go.dev/) 1.22+
- [Node.js](https://nodejs.org/) 18+
- [Docker](https://www.docker.com/) (for local PostgreSQL)
- `make` (optional, for Backend commands)

---

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/rahmatez/stockFlow-go.git
cd OMS-SaaS-go
```

### 2. Run the Backend

```bash
cd Backend

# Start PostgreSQL (port 5433)
make up

# Set up environment
cp .env.example .env

# Run the API (migrations run automatically on startup)
make run
```

API runs at **http://localhost:8080**

Health check: `GET /health`

### 3. Run the Frontend

```bash
cd Frontend

npm install
cp .env.local.example .env.local
npm run dev
```

Dashboard runs at **http://localhost:3000**

### 4. Create an account or seed demo data

**Option A — Register manually**

Open http://localhost:3000/register and create a new workspace.

**Option B — Seed demo data**

```bash
cd Backend
bash scripts/seed.sh
```

Sign in with:

| Email | Password |
|-------|----------|
| `demo@oms.local` | `password123` |

---

## Frontend Pages

| Route | Description |
|-------|-------------|
| `/login` | Sign in to your workspace |
| `/register` | Register a new business |
| `/dashboard` | KPI overview & charts |
| `/products` | Manage product catalog |
| `/inventory` | Stock & inventory movements |
| `/orders` | Order management |
| `/customers` | Customer records |
| `/reports` | Sales reports |
| `/settings/billing` | Subscription plans & usage limits |

---

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

| Module | Endpoints |
|--------|-----------|
| Auth | `POST /auth/register`, `/login`, `/refresh`, `/logout`, `/forgot-password`, `/reset-password` |
| Users | `GET/PATCH /users/me`, `GET/POST/PATCH/DELETE /users` |
| Tenant | `GET /tenant/me`, `PATCH /tenant/me` |
| Products | CRUD `/products`, `/categories`, `GET /products/export` |
| Warehouses | CRUD `/warehouses` |
| Inventory | `GET /inventory`, `POST /inventory/adjust`, `/transfer` |
| Orders | CRUD `/orders`, `PATCH /orders/:id/status` |
| Customers | CRUD `/customers` |
| Reports | `GET /reports/dashboard`, `/sales`, `/inventory`, export CSV |
| Notifications | `GET /notifications`, `PATCH /notifications/:id/read` |
| Billing | `GET /billing/plan`, `POST /billing/checkout`, `POST /billing/portal` |

Full documentation: [`Backend/docs/openapi.yaml`](Backend/docs/openapi.yaml)

---

## Environment Variables

### Backend (`Backend/.env`)

```env
APP_PORT=8080
DATABASE_URL=postgres://oms_app:oms@localhost:5433/oms?sslmode=disable
MIGRATION_DATABASE_URL=postgres://oms:oms@localhost:5433/oms?sslmode=disable
JWT_SECRET=change-me-to-a-long-random-secret-key
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
CORS_ORIGINS=http://localhost:3000
STRIPE_SECRET_KEY=sk_test_your_key
STRIPE_WEBHOOK_SECRET=whsec_your_secret
STRIPE_PRICE_STARTER=price_starter
STRIPE_PRICE_PRO=price_pro
SMTP_HOST=
SMTP_USER=
SMTP_PASS=
SMTP_FROM=
```

> **SMTP** is required for transactional email: password reset (`/auth/forgot-password`), team invite emails when adding members in **Settings → Team**, and other notifications sent by the API. If `SMTP_HOST` is empty, the API uses a no-op mailer and invite/password emails are skipped silently.

### Frontend (`Frontend/.env.local`)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

> Docker PostgreSQL uses **port 5433** to avoid conflicts with a local PostgreSQL instance on port 5432.
>
> The API connects as `oms_app` (non-superuser, RLS-enforced). Migrations run as `oms` via `MIGRATION_DATABASE_URL`.

---

## Development

### Backend

```bash
cd Backend
make dev      # Start DB + API together
make build    # Build binary
make test     # Run tests
make down     # Stop PostgreSQL
```

### Frontend

```bash
cd Frontend
npm run dev     # Development server
npm run build   # Production build
npm run lint    # ESLint
```

---

## Deployment

| Component | Recommended |
|-----------|-------------|
| Go API | Railway, Fly.io, VPS |
| PostgreSQL | Supabase, Neon, RDS |
| Next.js | Vercel |

**Backend**

1. Set all env vars from `.env.example` in your production environment
2. Ensure `DATABASE_URL` points to your production PostgreSQL instance
3. Build: `go build -o api ./cmd/api`
4. Migrations run automatically when the API starts

**Frontend**

1. Deploy the `Frontend/` folder to Vercel
2. Set `NEXT_PUBLIC_API_URL` to your production API URL
3. Ensure `CORS_ORIGINS` on the Backend includes your frontend domain

**Stripe (optional)**

- Configure `STRIPE_SECRET_KEY` and price IDs
- Register a webhook pointing to `POST /api/v1/billing/webhook/stripe`

---

## CI

A GitHub Actions pipeline is available at [`Backend/.github/workflows/ci.yml`](Backend/.github/workflows/ci.yml) — it runs Go build & tests with a PostgreSQL service.

---

## License

This project is private and does not currently have an open-source license. Contact the maintainer for further usage.

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m 'feat: short description'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request

---

<p align="center">
  <strong>StockFlow</strong> — Manage stock and orders in one place.
</p>
