SHELL=/bin/bash
SQL_MIGRATE_BIN=../vendor_ci/sql-migrate # based by 'migration' Dir
CONTAINER_NAME=mf-importer
CONTAINER_NAME_MAW=mf-importer-maw
CONTAINER_NAME_FRONT=mf-importer-frontend

.PHONY: bin build start test debug migration
bin:
	go build -a -tags "netgo" -installsuffix netgo  -ldflags="-s -w -extldflags \"-static\" \
	-X main.version=$(git describe --tag --abbrev=0) \
	-X main.revision=$(git rev-list -1 HEAD) \
	-X main.build=$(git describe --tags)" \
	-o build/bin/ ./...

build:
	docker build -t $(CONTAINER_NAME) -f build/Dockerfile .
	docker build -t $(CONTAINER_NAME_MAW) -f build/maw/Dockerfile .
	docker build -t $(CONTAINER_NAME_FRONT) -f build/frontend/Dockerfile .

start:
	docker compose -f deployment/compose.yml up -d

debug:
	docker compose -f deployment/compose.yml up

test: 
	gofmt -l .
	go vet -composites=false ./...
	staticcheck ./...
	go test -v ./...

migration:
	cd migration; \
	${SQL_MIGRATE_BIN} up -env=local; \
	cd ../
