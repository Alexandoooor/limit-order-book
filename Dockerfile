FROM golang:1.21-alpine AS build
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy the source
COPY . ./

# Build the binary
RUN go build -o limit-order-book main.go

# Final minimal image
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/limit-order-book ./
EXPOSE 3000
CMD ["./limit-order-book"]
