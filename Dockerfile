FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

RUN go mod download

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -o my_app ./cmd/app

FROM alpine:latest

WORKDIR /app

RUN apk update && \
    apk add sqlite && \
    mkdir web && \
    mkdir data && \
    mkdir migration

COPY --from=builder /app/my_app .
COPY migration/scheduler.sql migration/scheduler.sql
COPY web/ web

ENV TODO_PORT=7540 \
    TODO_DBFILE=data/scheduler.db \
    TODO_PASSWORD=gofinalproject \
    TODO_SECRET=secretkey
EXPOSE ${TODO_PORT}
CMD ["./my_app"]
