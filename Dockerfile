FROM golang:1.19-alpine

ARG GOOS=linux
ARG GOARCH=amd64

RUN mkdir /go/src/app
WORKDIR /go/src/app
ADD . /go/src/app

COPY go.mod go.sum ./
RUN go mod download

RUN apk update && apk add git
