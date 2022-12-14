FROM golang:1.19-alpine

RUN mkdir /go/src/app
WORKDIR /go/src/app
ADD . /go/src/app

COPY go.mod go.sum ./
RUN go mod download

RUN apk update
RUN apk add git gcc musl-dev

ENV CGO_ENABLED=1
