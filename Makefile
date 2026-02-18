.PHONY: help build build-linux docker-build docker-push docker-run docker-clean helm-lint helm-template helm-install helm-upgrade helm-uninstall helm-package proto generate clean test lint

APP_NAME := captcha
REGISTRY := ghcr.io
ORG ?= your-org
IMAGE := $(REGISTRY)/$(ORG)/$(APP_NAME)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
CHART_DIR := charts/$(APP_NAME)
BIN_DIR := bin

GO := go
GOFLAGS := -v
CGO_ENABLED := 0
GOOS := $(shell $(GO) env GOOS)
GOARCH := $(shell $(GO) env GOARCH)

DOCKER := docker
DOCKER_BUILDX := docker buildx

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Build targets:"
	@echo "  build           Build Go binary for current platform"
	@echo "  build-linux     Build Go binary for Linux (amd64)"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-push     Push Docker image to registry"
	@echo "  docker-run      Run Docker container locally"
	@echo "  docker-clean    Remove local Docker images"
	@echo ""
	@echo "Helm targets:"
	@echo "  helm-lint       Lint Helm chart"
	@echo "  helm-template   Render Helm templates"
	@echo "  helm-package    Package Helm chart"
	@echo "  helm-install    Install Helm chart to Kubernetes"
	@echo "  helm-upgrade    Upgrade Helm release"
	@echo "  helm-uninstall  Uninstall Helm release"
	@echo ""
	@echo "Development targets:"
	@echo "  generate        Generate code (protobuf)"
	@echo "  test            Run tests"
	@echo "  lint            Run linters"
	@echo "  clean           Clean build artifacts"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  REGISTRY=$(REGISTRY)"
	@echo "  ORG=$(ORG)"

build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(APP_NAME) cmd/proxy/main.go

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 cmd/proxy/main.go

docker-build:
	$(DOCKER) build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

docker-buildx:
	$(DOCKER_BUILDX) create --use builder
	$(DOCKER_BUILDX) build --platform linux/amd64,linux/arm64 -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

docker-push:
	$(DOCKER) push $(IMAGE):$(VERSION)
	$(DOCKER) push $(IMAGE):latest

docker-run:
	$(DOCKER) run -d --name $(APP_NAME) -p 9001:9001 $(IMAGE):$(VERSION)

docker-clean:
	$(DOCKER) rmi $(IMAGE):$(VERSION) $(IMAGE):latest 2>/dev/null || true

docker-stop:
	$(DOCKER) stop $(APP_NAME) 2>/dev/null || true
	$(DOCKER) rm $(APP_NAME) 2>/dev/null || true

helm-lint:
	helm lint $(CHART_DIR)

helm-template:
	helm template $(APP_NAME) $(CHART_DIR)

helm-template-debug:
	helm template $(APP_NAME) $(CHART_DIR) --debug

helm-package:
	helm package $(CHART_DIR) --version $(VERSION) --app-version $(VERSION)

helm-install:
	helm install $(APP_NAME) $(CHART_DIR) --namespace $(APP_NAME) --create-namespace

helm-upgrade:
	helm upgrade $(APP_NAME) $(CHART_DIR) --namespace $(APP_NAME)

helm-uninstall:
	helm uninstall $(APP_NAME) --namespace $(APP_NAME)

helm-rollback:
	helm rollback $(APP_NAME) --namespace $(APP_NAME)

proto:
	which protoc >/dev/null 2>&1 || (echo "protoc not found" && exit 1)
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/captcha/v1/captcha.proto

test:
	$(GO) test $(GOFLAGS) ./...

lint:
	which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || $(GO) vet ./...

clean:
	rm -rf $(BIN_DIR)
	rm -f $(APP_NAME)-*.tgz

.DEFAULT_GOAL := help
