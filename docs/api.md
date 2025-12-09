# Wallet Service API

Authentication:
- `Authorization: Bearer <jwt>` – full access.
- `x-api-key: <key>` – must be active, unexpired, and include permission(s).

Permissions for API keys: `deposit`, `transfer`, `read`. Max 5 active keys/user. Expiry options: `1H`, `1D`, `1M`, `1Y`.

## Auth
- `GET /auth/google` → redirect to Google consent.
- `GET /auth/google/callback?code=` → creates user+wallet if missing, returns JWT + wallet info.

## API Keys (JWT only)
- `POST /keys/create`
  - Body: `{ "name": "github.com/CyberwizD/Wallet-Service", "permissions": ["deposit","transfer","read"], "expiry": "1D" }`
  - Response: `{ "api_key": "...", "expires_at": "...", "permissions": "deposit,transfer,read" }`
- `POST /keys/rollover`
  - Body: `{ "expired_key_id": "...", "expiry": "1M" }`
  - Reuses the expired key's permissions.

## Wallet
- `POST /wallet/deposit` (JWT or API key with `deposit`)
  - Body: `{ "amount": 5000 }` (kobo)
  - Response: `{ "reference": "...", "authorization_url": "https://paystack.co/..." }`
- `POST /wallet/paystack/webhook`
  - Validates Paystack signature, idempotently credits wallet on `success`.
  - Response: `{ "status": true }`
- `GET /wallet/deposit/:reference/status`
  - Response: `{ "reference": "...", "status": "success|failed|pending", "amount": 5000 }`
- `GET /wallet/balance` (permission `read`)
  - Response: `{ "balance": 15000, "wallet_number": "..." }`
- `POST /wallet/transfer` (permission `transfer`)
  - Body: `{ "wallet_number": "dest", "amount": 3000 }`
  - Response: `{ "status": "success", "message": "Transfer completed" }`
- `GET /wallet/transactions` (permission `read`)
  - Response: list of transactions ordered newest first.
