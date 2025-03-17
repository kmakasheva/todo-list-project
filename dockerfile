# 1. Этап сборки
FROM golang:1.23-bullseye AS builder

WORKDIR /todo-list-project

RUN apt-get update && apt-get install -y gcc libc6-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CC=gcc CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /todo-list ./cmd

# 2. Финальный минимальный образ
FROM debian:bullseye

WORKDIR /app

COPY --from=builder /todo-list .
COPY --from=builder /todo-list-project/config ./config

CMD ["/app/todo-list"]