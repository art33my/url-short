FROM golang:1.23.5 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o url-short ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/url-short .
COPY --from=builder /app/.env .
EXPOSE 8080
CMD ["./url-short"]