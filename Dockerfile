FROM golang:1.25-alpine AS build
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Install build dependencies for CGO + SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Enable CGO
ENV CGO_ENABLED=1

# Copy the source
COPY . ./

# Build the binary
RUN go build -o limit-order-book main.go

# Final minimal image
FROM alpine:latest

RUN apk add --no-cache sqlite

WORKDIR /app

COPY schema.sql ./
RUN mkdir -p trades && \
    sqlite3 trades/orderbook.db < schema.sql

COPY --from=build /app/limit-order-book ./
EXPOSE 3000
CMD ["./limit-order-book"]
