# Stage 1: Build the Go application
FROM golang:1.23.2-alpine AS builder

# Install git (required for go mod download if using private repositories)
RUN apk update && apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to leverage Docker's caching mechanism
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the backend source code into the container
COPY backend/ ./backend/

# Change directory to where main.go is located
WORKDIR /app/backend/cmd

# Build the Go application
RUN go build -o fms-backend .

# Stage 2: Create a minimal image to run the application
FROM alpine:latest

# Set the working directory inside the runtime container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/backend/cmd/fms-backend .

# Expose the port your app runs on
EXPOSE 8080

# Set environment variables (optional, can be managed via Railway)
ENV PORT=8080

# Command to run the application
CMD ["./fms-backend"]
