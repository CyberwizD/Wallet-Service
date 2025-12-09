# Testing & Local Runs

## Automated Tests
```
# disable parent go.work if present
$env:GOWORK="off"   # PowerShell
GOWORK=off go test ./...
```
Tests use in-memory sqlite to validate:
- API key limits and validation.
- Wallet transfers (atomic balance updates).

## Manual Smoke
1) Export env vars (`.env.example` has a template).
2) Start Postgres and run `go run ./cmd/server`.
3) Complete Google OAuth via `/auth/google` to obtain a JWT.
4) Hit wallet endpoints with `Authorization: Bearer <jwt>` or `x-api-key` as needed.

Paystack: point dashboard webhook to `/wallet/paystack/webhook`; only webhook credits deposits.
