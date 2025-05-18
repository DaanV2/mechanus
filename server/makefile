include .env
export

build:
	go build ./...

test:
	go tool ginkgo run --output-dir reports --cover --covermode atomic -coverprofile=coverage.out ./...

coverage-report:
	go test ./... --cover -coverprofile=reports/coverage.out --covermode atomic --coverpkg=./...
	go tool cover -html=reports/coverage.out

generate:
	go generate ./...

lint:
	go tool golangci-lint run -v --fix

checks:
	$(MAKE) build
	$(MAKE) generate
	$(MAKE) test
	$(MAKE) lint