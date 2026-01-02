FROM golang:alpine

RUN apk add  \
    opencv-dev \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    pkgconf \
    build-base \
    wget \
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

RUN mkdir -p logs && chmod 777 logs

EXPOSE 8080

CMD ["./main"]
