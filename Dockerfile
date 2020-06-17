FROM golang:1.14-buster as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY Makefile /app
COPY pkg /app/pkg
COPY cmd /app/cmd
ARG COMMIT_SHA

RUN make

FROM busybox

COPY --from=builder /app/supervisor /usr/local/bin/supervisor
RUN chmod +x /usr/local/bin/supervisor

