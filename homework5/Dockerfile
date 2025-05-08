FROM golang:1.24-alpine

WORKDIR /

RUN apk add --no-cache git


COPY . /app

WORKDIR /app/cmd/sub
RUN go build -mod vendor -o /app/sub

WORKDIR /app/cmd/pub
RUN go build -mod vendor -o /app/pub

WORKDIR /app/cmd/request
RUN go build -mod vendor -o /app/request

WORKDIR /app/cmd/reply
RUN go build -mod vendor -o /app/reply

