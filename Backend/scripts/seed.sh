#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080/api/v1}"

echo "Registering demo tenant..."
RESP=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_name": "Demo Store",
    "email": "demo@oms.local",
    "password": "password123",
    "full_name": "Demo Owner"
  }')

TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['tokens']['access_token'])" 2>/dev/null || echo "")

if [ -z "$TOKEN" ]; then
  echo "Registration failed or user exists. Trying login..."
  RESP=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"demo@oms.local","password":"password123"}')
  TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['tokens']['access_token'])")
fi

echo "Creating sample products..."
curl -s -X POST "$API_URL/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"sku":"SKU-001","name":"Widget A","sell_price":50000,"cost_price":30000,"low_stock_threshold":5}' > /dev/null

curl -s -X POST "$API_URL/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"sku":"SKU-002","name":"Widget B","sell_price":75000,"cost_price":45000,"low_stock_threshold":10}' > /dev/null

echo "Creating sample customer..."
curl -s -X POST "$API_URL/customers" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"John Customer","email":"john@example.com","phone":"08123456789"}' > /dev/null

echo "Seed complete! Login with demo@oms.local / password123"
