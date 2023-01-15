.PHONY: build
build:
	-docker buildx create --name edgedb-ingest-builder --driver docker-container --bootstrap --platform linux/amd64,linux/arm64
	docker buildx use edgedb-ingest-builder
	docker buildx build --platform linux/amd64,linux/arm64 -t mfreudenberg/edgedb-ingest:latest --push .
	docker buildx build -t edgedb-ingest --load .

.PHONY: build-local
build-local:
	CGO_ENABLED=0 go build -o build/edgedb-ingest ./cmd/edgedb-ingest

.PHONY: modules
modules:
	go mod tidy

.PHONY: run
run:
	go run cmd/edgedb-ingest/*.go

.PHONY: up
up:
	docker-compose up --build
