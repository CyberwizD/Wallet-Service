# Wallet Service (Go)

Backend wallet service with Google JWT auth, API keys, Paystack deposits, webhook crediting, wallet transfers, balances, and transaction history.

## Stack
- Go 1.22+
- Gin HTTP framework
- PostgreSQL (GORM)
- Paystack payments
- Google OAuth2 for sign-in → service JWT

## Project Layout
- `cmd/server/main.go` – entrypoint
- `internal/config` – env loading, expiry parsing
- `internal/database` – DB bootstrap
- `internal/models` – GORM entities
- `internal/services` – business logic (users, wallet, Paystack, API keys)
- `internal/handlers` – HTTP handlers
- `internal/middleware` – JWT/API-key auth + permission checks
- `internal/server` – router wiring
- `internal/util` – helpers (IDs, random, permissions)

## Environment
```
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/wallet?sslmode=disable
JWT_SECRET=super-secret
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
PAYSTACK_SECRET_KEY=sk_test_xxx
# PAYSTACK_BASE_URL optional (defaults to https://api.paystack.co)
```

> Amounts are stored and processed in kobo (smallest currency unit).

## Run
```
# if a parent go.work interferes, disable it for commands
$env:GOWORK="off"   # PowerShell; or GOWORK=off for *nix shells
go run ./cmd/server
```
Server starts on `:$PORT`, auto-migrating the schema.

## API (high level)
- `GET /auth/google` – redirect to Google consent
- `GET /auth/google/callback` – exchanges code, upserts user+wallet, returns JWT
- `POST /keys/create` – JWT only. Body: `{ "name": "...", "permissions": ["deposit","transfer","read"], "expiry": "1D" }` (max 5 active keys/user)
- `POST /keys/rollover` – JWT only. Body: `{ "expired_key_id": "...", "expiry": "1M" }`
- `POST /wallet/deposit` – JWT or API key with `deposit`. Body: `{ "amount": 5000 }` → `{ reference, authorization_url }`
- `POST /wallet/paystack/webhook` – Paystack webhook (signature verified). Idempotently credits on `success`.
- `GET /wallet/deposit/:reference/status` – status only (never credits)
- `GET /wallet/balance` – JWT or API key with `read`
- `POST /wallet/transfer` – JWT or API key with `transfer`. Body: `{ "wallet_number": "...", "amount": 3000 }`
- `GET /wallet/transactions` – JWT or API key with `read`

### Auth rules
- `Authorization: Bearer <jwt>` → full wallet access
- `x-api-key: <key>` → must be active, unexpired, and include required permission
- API keys expire (1H/1D/1M/1Y), can be revoked/rolled over, max 5 active/user

### Paystack
- `/wallet/deposit` initializes a Paystack transaction with a unique reference.
- Only the webhook credits wallets (idempotent on repeated payloads).
- Webhook signature checked with HMAC-SHA512 using `PAYSTACK_SECRET_KEY`.

## Dev Notes
- Database migrations are handled via GORM auto-migrate on startup.
- Transfers are executed inside DB transactions with row-level locking to avoid race conditions and ensure atomic balance updates.
- Transaction history is per wallet and ordered by newest first.

## Docs
- API: `docs/api.md`
- OpenAPI: `docs/swagger.yaml`
- Deployment: `docs/deployment.md`
- Testing & manual smoke: `docs/testing.md`

## Quick Smoke (after setting env + Postgres)
```
$env:GOWORK="off"
go run ./cmd/server
# In another shell:
curl -i http://localhost:8080/auth/google   # follow OAuth flow to obtain JWT
```
