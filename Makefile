HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH_RAW := $(shell uname -m)
HOST_ARCH := $(HOST_ARCH_RAW)
ifeq ($(HOST_ARCH_RAW),x86_64)
HOST_ARCH := amd64
endif
ifeq ($(HOST_ARCH_RAW),aarch64)
HOST_ARCH := arm64
endif

.PHONY: up down build logs cli web-install web-build web-dev

up:
	docker compose -f docker-compose.demo.yml up -d

down:
	docker compose -f docker-compose.demo.yml down

build:
	docker compose -f docker-compose.demo.yml build api

logs:
	docker compose -f docker-compose.demo.yml logs -f api

cli:
	docker run --rm \
		-e CGO_ENABLED=0 \
		-e GOOS=$(HOST_OS) \
		-e GOARCH=$(HOST_ARCH) \
		-v "$(PWD):/workspace" \
		-w /workspace \
		golang:1.24 \
		go build -o bin/vda ./cmd/vda

web-install:
	cd web && npm install

web-build:
	cd web && npm run build

web-dev:
	cd web && npm run dev -- --port 5173
