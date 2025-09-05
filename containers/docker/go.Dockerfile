FROM golang:1.25.0-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/service

FROM alpine:3.20

RUN apk --no-cache add curl

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/static ./static

RUN adduser -D -s /bin/sh appuser \
    && chown -R appuser:appuser /app

USER appuser

CMD ["./main"]