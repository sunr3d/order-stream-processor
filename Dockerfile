FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o order-stream-processor ./cmd/main.go

FROM alpine:3.21

WORKDIR /app
RUN adduser -D -g '' appuser
COPY --from=builder /app/order-stream-processor .
RUN chown appuser:appuser /app
USER appuser

EXPOSE 8081
CMD ["./order-stream-processor"]