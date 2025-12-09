# Deployment (Docker)

## Build
```
docker build -t wallet-service:latest .
```

## Run (single container)
```
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e DATABASE_URL=postgres://user:pass@db:5432/wallet?sslmode=disable \
  -e JWT_SECRET=replace-me \
  -e GOOGLE_CLIENT_ID=... \
  -e GOOGLE_CLIENT_SECRET=... \
  -e GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback \
  -e PAYSTACK_SECRET_KEY=sk_live_xxx \
  -e PAYSTACK_WEBHOOK_SECRET=sk_live_xxx \
  wallet-service:latest
```

## With docker-compose (app + Postgres)
```yaml
version: "3.9"
services:
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: wallet
      POSTGRES_PASSWORD: wallet
      POSTGRES_DB: wallet
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U wallet"]
      interval: 5s
      timeout: 5s
      retries: 5

  api:
    build: .
    depends_on:
      db:
        condition: service_healthy
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      DATABASE_URL: postgres://wallet:wallet@db:5432/wallet?sslmode=disable
      JWT_SECRET: replace-me
      GOOGLE_CLIENT_ID: your-client-id
      GOOGLE_CLIENT_SECRET: your-client-secret
      GOOGLE_REDIRECT_URL: http://localhost:8080/auth/google/callback
      PAYSTACK_SECRET_KEY: sk_live_xxx
      PAYSTACK_WEBHOOK_SECRET: sk_live_xxx
```

> Ensure `PAYSTACK_WEBHOOK_SECRET` matches the signature secret configured in Paystack. Update `GOOGLE_REDIRECT_URL` to your deployed domain in production.
