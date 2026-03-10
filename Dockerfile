# Build stage
FROM golang:1.25 AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run code generators (generate banner and other generated files)
RUN go generate main.go

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o zerobot-plugin .

# Runtime stage
FROM alpine:latest

# [新增] 替换为阿里云镜像源，解决 TLS 连接错误和速度慢的问题
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /home/appuser/

# Copy the binary from builder stage
COPY --from=builder /app/zerobot-plugin .

# Create data directory and set ownership
RUN mkdir -p /home/appuser/data \
	&& chown -R appuser:appuser /home/appuser/

# Declare data as a mountable volume
VOLUME ["/home/appuser/data"]

# Switch to non-root user
USER appuser

RUN /home/appuser/zerobot-plugin -s /home/appuser/config.json || true

# Run the bot by default (absolute path)
CMD ["/home/appuser/zerobot-plugin"]