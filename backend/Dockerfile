# ==========================================
# Stage 1: Build Executable Binary
# ==========================================
FROM golang:1.24-alpine AS builder

# Install security certificates & git
RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /build

# Cache Go modules dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY . .

# Compile pure Go static binary (No CGO required)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X 'main.Version=1.0.0'" \
    -trimpath \
    -o cbt-backend ./cmd/api

# ==========================================
# Stage 2: Final Minimal Runtime Image
# ==========================================
FROM alpine:3.20

# Install runtime dependencies (certificates, timezone)
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S cbtgroup && adduser -S cbtuser -G cbtgroup

WORKDIR /app

# Ensure uploads directory exists and is owned by cbtuser
RUN mkdir -p /app/uploads/questions && chown -R cbtuser:cbtgroup /app/uploads

# Copy binary & timezone from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/cbt-backend /app/cbt-backend

# Set environment
ENV TZ=Asia/Jakarta
ENV PORT=8080

# Run as non-privileged user for container security
USER cbtuser:cbtgroup

EXPOSE 8080

# Health check to ensure service readiness
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./cbt-backend"]