FROM golang:1.23.5 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o url-short ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/url-short /app/url-short
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["/app/url-short"]