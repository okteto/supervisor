COMMIT_SHA := $(shell git rev-parse --short HEAD)
TAG ?= "supervisor"

.DEFAULT_GOAL := build

.PHONY: build
build:
	CGO=0 GOOS=linux go build -o supervisor -ldflags "-X main.CommitString=${COMMIT_SHA}" -tags "osusergo netgo static_build" cmd/main.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: push
push:
	okteto build -t ${TAG} .

multi:
	# docker buildx create --name mbuilder
	docker buildx use mbuilder
	docker buildx build  --platform linux/amd64,linux/arm64 -t ${TAG} --push .