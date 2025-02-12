# Stage 1: Build the application
FROM golang:1.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy Go modules manifests and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application source code
COPY . .

# Build the Go application
RUN go build -o email-service .

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

# Install required certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/email-service .

# Expose the application port
EXPOSE 5004

# Command to run the service
CMD ["./email-service"]
