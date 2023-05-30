CONTAINER_NAME=mf-importer

.PHONY: build start debug
build:
	docker build -t $(CONTAINER_NAME) -f build/Dockerfile .

start:
	docker compose -f deployment/compose.yml up -d

debug:
	docker compose -f deployment/compose.yml up
