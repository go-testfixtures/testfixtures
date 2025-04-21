FROM golang:1.23-alpine

RUN apk update
RUN apk add alpine-sdk

WORKDIR /testfixtures

COPY go.mod go.sum ./
RUN go mod download

COPY . .
