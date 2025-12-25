FROM golang:alpine

RUN apk add --no-cache \
    opencv-dev \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    pkgconf \
    build-base \
    wget

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -tags gocv -o main ./cmd/server

RUN mkdir -p logs && chmod 777 logs

EXPOSE 8080

CMD ["./main"]
