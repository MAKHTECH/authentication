# Build stage
FROM golang:1.24.1-alpine AS builder

# Install dependencies for CGO (required for SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /build

# Copy and download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the project
COPY . .

# Build the application with verbose output for debugging
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -o sso ./sso/cmd/sso

# Verify the file exists
RUN ls -l /build/sso

# Final stage
FROM alpine:latest

# Install SQLite runtime dependencies
RUN apk add --no-cache sqlite-libs

# Set working directory
WORKDIR /go/src/app

# Copy the built executable from the builder stage
COPY --from=builder /build/sso .

# Verify the file exists in the final stage
RUN ls -l /go/src/app/sso

# Set the command to run the application
CMD ["/go/src/app/sso"]