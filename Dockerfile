FROM golang:1.26-alpine AS builder

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

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -tags gocv -o main ./cmd/server

FROM alpine:latest

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

COPY --from=builder /app/main .

COPY --from=builder /app/web ./web
COPY --from=builder /app/models ./models

RUN mkdir -p logs && chmod 777 logs

EXPOSE 8080

CMD ["./main"]
