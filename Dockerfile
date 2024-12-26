# Use an official Golang image as the base image
FROM golang:1.20 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o main .

# Create a minimal image for the final application
FROM alpine:3.18

# Install certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Set the working directory in the final image
WORKDIR /root/

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/main .

# Expose the application's port
EXPOSE 3000

# Command to run the application
CMD ["./main"]
