FROM golang:1.12.4

WORKDIR /testfixtures
COPY . .

RUN go mod download
