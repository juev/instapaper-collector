FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /instapaper-collector ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache git && \
    adduser -D runner -u 1001

COPY --from=builder /instapaper-collector /usr/local/bin/instapaper-collector

USER runner
