COMMIT_SHA := $(shell git rev-parse --short HEAD)

.DEFAULT_GOAL := build

.PHONY: build
build:
	CGO=0 GOOS=linux go build -o supervisor -ldflags "-X main.CommitString=${COMMIT_SHA}" -tags "osusergo netgo static_build" cmd/main.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: push
push:
	okteto build -t okteto/supervisor:0.1.1 .
