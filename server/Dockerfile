# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Ensure go.mod & go.sum are perfectly synced
RUN go mod tidy

# Build lightweight static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server main.go

# Production Stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates and tzdata for secure network/TLS calls
RUN apk --no-cache add ca-certificates tzdata

# Copy compiled binary from builder stage
COPY --from=builder /app/server .

# Create uploads folder for local file storage
RUN mkdir -p ./uploads

EXPOSE 8000

ENV PORT=8000

CMD ["./server"]
