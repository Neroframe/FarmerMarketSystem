# Use the official Go image as the base image
FROM golang:1.23.2 as builder

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

# Use a lightweight image for the runtime environment
FROM debian:buster-slim

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /fms-backend .

# Expose the port your application listens on
EXPOSE 8080

# Command to run the application
CMD ["./fms-backend"]
