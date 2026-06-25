# OMS SaaS Mini — Next.js Frontend

Inventory & Order Management System dashboard built with Next.js 15, TypeScript, and Tailwind CSS.

## Stack

- Next.js 15 (App Router)
- TypeScript
- Tailwind CSS
- TanStack Query
- React Hook Form + Zod
- Recharts

## Quick Start

```bash
# Install dependencies
npm install

# Copy env
cp .env.local.example .env.local

# Start dev server (requires Go API running)
npm run dev
```

Frontend runs at `http://localhost:3000`

## Pages

| Route | Description |
|-------|-------------|
| `/login` | User login |
| `/register` | Tenant registration |
| `/dashboard` | KPI dashboard |
| `/products` | Product CRUD |
| `/inventory` | Stock movements |
| `/orders` | Order management |
| `/customers` | Customer CRUD |
| `/reports` | Sales & inventory reports |
| `/settings/billing` | Subscription plans |

## Environment

```
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## Backend

This frontend connects to the Go API in the sibling folder `Backend`.
