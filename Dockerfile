# Use the official Go image as the build stage
FROM golang:1.23.2 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app
RUN go build -o /fms-backend ./backend/cmd

# Use a newer Debian image with the required GLIBC version for the runtime environment
FROM debian:bookworm-slim

# Set the working directory
WORKDIR /app

# Install necessary dependencies
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from the builder stage
COPY --from=builder /fms-backend .

# Copy the templates directory from the builder stage
COPY --from=builder /app/web/templates /app/web/templates

# Expose the port your application listens on
EXPOSE 8080

# Command to run the application
CMD ["./fms-backend"]
