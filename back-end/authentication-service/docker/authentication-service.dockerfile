# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# Set the working directory inside the builder container
WORKDIR /app

# Copy Go module files first (for better caching)
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the source code (including SQL files)
COPY src /app/src

# Build the Go application binary
RUN go build -o /app/bin/authentication-serviceBinary ./src/cmd/

# Stage 2: Final minimal image
FROM alpine:latest

# Set working directory in final image
WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/bin/authentication-serviceBinary .

# Copy the SQL directory for initialization
COPY --from=builder /app/src/sql /app/sql

# ✅ Set the ENV variable so Go app can read the SQL init script path
ENV AUTHENTICATION_INIT_SQL_FILE_PATH=/app/sql/init_users_table.sql

# Default command to run the binary
CMD ["/app/authentication-serviceBinary"]
