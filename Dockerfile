FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o my_app .

FROM ubuntu:latest

WORKDIR /app

RUN apt-get update && \
    apt-get install -y sqlite3 && \
    mkdir web && \
    mkdir database

COPY --from=builder /app/my_app .
COPY scheduler.sql scheduler.sql
COPY web/ /web

ENV TODO_PORT=7540 \
    TODO_DBFILE=database/scheduler.db \
    TODO_PASSWORD=gofinalproject \
    TODO_SECRET=secretkey
EXPOSE 7540
CMD ["./my_app"]
