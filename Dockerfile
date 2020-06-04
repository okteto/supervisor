FROM golang:1.14 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY Makefile /app
COPY pkg /app/pkg
COPY cmd /app/cmd
COPY .git /app/.git

RUN git rev-parse --short HEAD
RUN make

FROM alpine

COPY --from=builder /app/supervisor /usr/local/bin/supervisor
RUN chmod +x /usr/local/bin/supervisor

