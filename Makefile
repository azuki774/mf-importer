SHELL=/bin/bash
SQL_MIGRATE_BIN=../vendor_ci/sql-migrate # based by 'migration' Dir
CONTAINER_NAME=mf-importer
CONTAINER_NAME_MAW=mf-importer-maw

.PHONY: bin build start debug migration
bin:
	go build -a -tags "netgo" -installsuffix netgo  -ldflags="-s -w -extldflags \"-static\" \
	-X main.version=$(git describe --tag --abbrev=0) \
	-X main.revision=$(git rev-list -1 HEAD) \
	-X main.build=$(git describe --tags)" \
	-o build/bin/ ./...

build: bin
	docker build -t $(CONTAINER_NAME) -f build/Dockerfile .
	docker build -t $(CONTAINER_NAME_MAW) -f build/Dockerfile-maw .

start:
	docker compose -f deployment/compose.yml up -d

debug:
	docker compose -f deployment/compose.yml up

pytest:
	dbpass="password" pytest -v

test: pytest
	gofmt -l .
	go vet -composites=false ./...
	staticcheck ./...
	go test -v ./...

migration:
	cd migration; \
	${SQL_MIGRATE_BIN} up -env=local; \
	cd ../
