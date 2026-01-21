# Stage 1: Builder
FROM golang:alpine AS builder

# Install build dependencies
# opencv-dev: Headers for compiling GoCV
# build-base: GCC and standard build tools
# pkgconf: For finding libraries
RUN apk add --no-cache \
    git \
    opencv-dev \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    pkgconf \
    build-base \
    gst-plugins-base \
    gst-plugins-good \
    gst-plugins-bad \
    gst-plugins-ugly \
    gst-libav

# Install golangci-lint for CI/CD checks inside the container
# This allows running linting in the same environment as the build
RUN wget -O- -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5

WORKDIR /app

# Dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# -tags gocv: Required for GoCV
# -o main: Output binary name
RUN go build -tags gocv -o main ./cmd/server

# Stage 2: Production
FROM alpine:latest

# Install runtime dependencies
# These must be compatible with the versions linked in the builder stage
RUN apk add --no-cache \
    opencv \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    gst-plugins-base \
    gst-plugins-good \
    gst-plugins-bad \
    gst-plugins-ugly \
    gst-libav \
    libstdc++ \
    ca-certificates \
    tzdata

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy assets required at runtime
COPY --from=builder /app/web ./web
COPY --from=builder /app/models ./models

# Create logs directory
RUN mkdir -p logs && chmod 777 logs

EXPOSE 8080

CMD ["./main"]
