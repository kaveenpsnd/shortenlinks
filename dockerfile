# Stage 1: Builder (Compile the Go code)
FROM golang:alpine as builder
WORKDIR /app

# Copy dependency files first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code (This respects .dockerignore)
COPY . .

# Build the binary
# We point to your specific main file location: cmd/api/main.go
RUN go build -o main ./cmd/api/main.go

# Stage 2: Runner (Tiny production image)
FROM alpine:latest
WORKDIR /root/

# Copy only the compiled binary from Stage 1
COPY --from=builder /app/main .

# Expose the API port
EXPOSE 8080

# Command to start the server
CMD ["./main"]