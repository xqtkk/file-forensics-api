# ---- Этап 1: сборка ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Сначала копируем go.mod и go.sum — это позволяет кэшировать зависимости
COPY go.mod go.sum ./
RUN go mod download

# Потом копируем весь код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ---- Этап 2: финальный образ ----
FROM alpine:3.19

WORKDIR /app

# Копируем только бинарник из этапа сборки
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]