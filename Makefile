SHELL=/bin/bash
SQL_MIGRATE_BIN=../vendor_ci/sql-migrate # based by 'migration' Dir
CONTAINER_NAME=mf-importer
CONTAINER_NAME_MAW=mf-importer-maw
CONTAINER_NAME_FRONT=mf-importer-fe
CONTAINER_NAME_API=mf-importer-api
CONTAINER_NAME_METRICS=mf-importer-metrics
OPENAPI_YAML=internal/openapi/mfimporter-api.yaml
pwd := $(shell pwd)
API_BIN=build/bin/mf-importer-api
URL ?= http://127.0.0.1:8080
LATEST ?= 5

.PHONY: bin build start stop test debug migration api-bin mock-api local-ui report e2e
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

api-bin:
	go build -o $(API_BIN) ./cmd/mf-importer-api

# DB 不要のモック API を起動する (test/ 配下の CSV を fixture 化)
mock-api: api-bin
	$(API_BIN) mock --input-dir ./test

# モック API + ビルド済みフロントエンドを :8080 の単一プロセスで配信し、ブラウザから確認できるようにする
local-ui: api-bin
	cd frontend && npm run generate
	$(API_BIN) mock --input-dir ./test --static-dir frontend/.output/public

# 取り込み結果を 1 つの JSON サマリとして出力する (AI 向け。curl + jq を利用)
report:
	@URL=$(URL) LATEST=$(LATEST) ./scripts/report.sh

# Playwright による画面テストとスクリーンショット取得を実行する (要 cd frontend && npm install 済み)
e2e: api-bin
	cd frontend && npm run generate
	cd frontend && npx playwright install chromium
	cd frontend && npx playwright test
