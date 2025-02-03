FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my_app

FROM ubuntu:latest

WORKDIR /app

RUN apt install sqlite3

COPY --from=builder /my_app /app/my_app

ENV TODO_PORT=7540 \
    TODO_DBFILE=database/scheduler.db \
    TODO_PASSWORD=gofinalproject  \
    TODO_SECRET=secretkey
EXPOSE 7540
CMD ["/my_app"]
