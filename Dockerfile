FROM golang:1.25-alpine

RUN apk add --no-cache alpine-sdk

WORKDIR /testfixtures

COPY "go.mod"  "go.mod"
COPY "go.sum"  "go.sum"
COPY "go.work" "go.work"
COPY "go.work.sum" "go.work.sum"
COPY "cmd/testfixtures/go.mod" "cmd/testfixtures/go.mod"
COPY "cmd/testfixtures/go.sum" "cmd/testfixtures/go.sum"
COPY "dbtests/go.mod" "dbtests/go.mod"
COPY "dbtests/go.sum" "dbtests/go.sum"

RUN go mod download

COPY . .
