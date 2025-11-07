FROM golang:1.23-alpine AS builder
WORKDIR /app

# Кешируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники и собираем
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o message-service ./cmd

FROM alpine:latest
WORKDIR /app

# Устанавливаем curl для healthcheck
RUN apk --no-cache add ca-certificates curl

# Копируем бинарник
COPY --from=builder /app/message-service .
RUN chmod +x message-service

ENTRYPOINT ["./message-service"]

EXPOSE 8080

