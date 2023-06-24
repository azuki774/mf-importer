CONTAINER_NAME=mf-importer
CONTAINER_NAME_MAW=mf-importer-maw

.PHONY: bin build start debug
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
	pytest -v

test:
	gofmt -l .
	go vet -composites=false ./...
	staticcheck ./...
	go test -v ./...
