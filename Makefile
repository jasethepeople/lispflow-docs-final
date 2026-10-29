.PHONY: deps run migrate integration openapi-docs helm-template helm-install helm-upgrade

deps:
	go mod tidy

run:
	go run ./cmd/api

migrate:
	psql $$DATABASE_URL -f schema.sql

integration:
	go test -v -tags=integration ./tests/integration/...

openapi-docs:
	npx @redocly/cli preview-docs openapi/openapi.yaml -p 8081

helm-template:
	helm template lispflow ./helm/lispflow

helm-install:
	helm install lispflow ./helm/lispflow --create-namespace --namespace lispflow

helm-upgrade:
	helm upgrade --install lispflow ./helm/lispflow --namespace lispflow

all:
	go build ./...
