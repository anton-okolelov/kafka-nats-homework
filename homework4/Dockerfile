FROM golang:1.24-alpine

WORKDIR /

RUN apk add --no-cache git curl nano bash postgresql-client ca-certificates

COPY . /app

WORKDIR /app/cmd/microservice1

RUN go build -mod vendor -o /app/microservice1

WORKDIR /app/cmd/microservice2

RUN go build -mod vendor -o /app/microservice2