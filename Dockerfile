# Multi-stage build: compile Go binary, then copy to minimal Alpine container.
# Final image is ~15MB with zero development tools (tiny attack surface).

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gateway ./cmd/gateway

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/gateway /gateway
EXPOSE 8080
CMD ["/gateway"]
