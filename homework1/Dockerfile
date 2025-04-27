FROM golang:1.24-alpine

WORKDIR /

RUN apk add --no-cache git


COPY . /app

WORKDIR /app/cmd/consumer

RUN go build -mod vendor -o /app/consumer

WORKDIR /app/cmd/producer

RUN go build -mod vendor -o /app/producer