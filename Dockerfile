# Stage 1: The Builder Stage
# We start with a Go image to compile the application
FROM golang:1.24-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go module files (go.mod and go.sum)
# This allows Docker to cache the dependency download step
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
# -o app: names the output binary 'app'
# -ldflags: strips debugging symbols for a smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w' -o app ./cmd/api

# ---

# Stage 2: The Final Runtime Stage
# We use a minimal base image, like alpine, to run the compiled binary
FROM alpine:latest

# Set a non-root user (good security practice)
RUN adduser -D nonroot
USER nonroot

# Set the working directory to the directory where the app will run
WORKDIR /app

# Copy the compiled binary from the 'builder' stage
COPY --from=builder /app/app .

# Expose the port your Go application listens on (e.g., 8080)
EXPOSE 8080

# Command to run the application when the container starts
CMD ["./app"]