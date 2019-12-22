FROM golang:1.13.5

WORKDIR /testfixtures
COPY . .

RUN go mod download
