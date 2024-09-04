SHELL=/bin/bash
SQL_MIGRATE_BIN=../vendor_ci/sql-migrate # based by 'migration' Dir
CONTAINER_NAME=mf-importer
CONTAINER_NAME_MAW=mf-importer-maw
CONTAINER_NAME_FRONT=mf-importer-fe
CONTAINER_NAME_API=mf-importer-api
CONTAINER_NAME_DOC=mf-importer-doc
CONTAINER_NAME_METRICS=mf-importer-metrics
OPENAPI_YAML=internal/openapi/mfimporter-api.yaml
pwd := $(shell pwd)

.PHONY: bin build start stop test debug migration doc
bin:
	go build -a -tags "netgo" -installsuffix netgo  -ldflags="-s -w -extldflags \"-static\" \
	-X main.version=$(git describe --tag --abbrev=0) \
	-X main.revision=$(git rev-list -1 HEAD) \
	-X main.build=$(git describe --tags)" \
	-o build/bin/ ./...

build:
	docker build -t $(CONTAINER_NAME) -f build/Dockerfile .
	docker build -t $(CONTAINER_NAME_MAW) -f build/maw/Dockerfile .
	docker build -t $(CONTAINER_NAME_API) -f build/api/Dockerfile .
	docker build -t $(CONTAINER_NAME_METRICS) -f build/metrics/Dockerfile .
	docker build -t $(CONTAINER_NAME_FRONT) -f build/fe/Dockerfile .

start:
	docker compose -f deployment/compose.yml up -d

stop:
	docker compose -f deployment/compose.yml down

debug:
	docker compose -f deployment/compose.yml up

test: 
	gofmt -l .
	go vet -composites=false ./...
	staticcheck ./...
	go test -v ./...

migration:
	cd migration && \
	${SQL_MIGRATE_BIN} up -env=local && \
	cd ../

generate:
	oapi-codegen -package "openapi" -generate "chi-server" ${OPENAPI_YAML} > internal/openapi/server.gen.go
	oapi-codegen -package "openapi" -generate "spec"       ${OPENAPI_YAML} > internal/openapi/spec.gen.go
	oapi-codegen -package "openapi" -generate "types"      ${OPENAPI_YAML} > internal/openapi/types.gen.go

doc:
	docker build -t $(CONTAINER_NAME_DOC) -f docs/Dockerfile .
	docker run --rm -it -v $(pwd)/internal/openapi/:/data/ $(CONTAINER_NAME_DOC)
	mv -f internal/openapi/api.html docs/api.html
