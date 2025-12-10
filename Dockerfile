# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

# Runtime stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /bin/server /app/server
COPY --from=builder /app/docs /app/docs
EXPOSE 8080
ENV PORT=8080
CMD ["/app/server"]
