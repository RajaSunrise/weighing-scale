# Use official Golang image based on Debian Bookworm
# Debian Bookworm includes OpenCV 4.6+ and Tesseract 5 which are compatible
FROM golang:1.25-bookworm

# Install system dependencies
# libopencv-dev: OpenCV development headers and libraries
# tesseract-ocr: OCR engine
# tesseract-ocr-eng: English data for OCR (or ind for Indonesian if available/needed)
# pkg-config: For cgo to find libraries
RUN apt-get update && apt-get install -y \
    libopencv-dev \
    tesseract-ocr \
    tesseract-ocr-eng \
    pkg-config \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# -tags gocv enables the OpenCV implementation
RUN go build -tags gocv -o main ./cmd/server

# Expose port (adjust if needed, default 8080)
EXPOSE 8080

# Run the application
CMD ["./main"]
