include .env
export

assembly:
	go build -gcflags="-S" ./mechanus/screens > cmd.S 2>&1

build:
	go build ./...

test:
	go tool ginkgo run --vv --output-dir reports --cover --covermode atomic -coverprofile=coverage.out ./...

coverage-report:
	go test ./... --cover -coverprofile=reports/coverage.out --covermode atomic --coverpkg=./...
	go tool cover -html=reports/coverage.out

generate:
	go generate ./...

lint:
	go tool golangci-lint run -v --fix

format:
	go fmt ./...

checks:
	$(MAKE) build
	$(MAKE) test
	$(MAKE) format
	$(MAKE) lint
	$(MAKE) documentation

documentation:
	go run ./scripts/documentation

start-server:
	go run main.go server