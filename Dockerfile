FROM golang:1.24-bookworm AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY Makefile /app
COPY pkg /app/pkg
COPY cmd /app/cmd
ARG COMMIT_SHA

RUN make

FROM busybox:1.37.0

COPY --from=builder /app/supervisor /usr/local/bin/supervisor
RUN chmod +x /usr/local/bin/supervisor

