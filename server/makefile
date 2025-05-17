test:
	go test ./... --cover -coverprofile=reports/coverage

coverage-report: test
	go tool cover -html=reports/coverage

generate:
	go generate ./...

lint:
	go tool golangci-lint run -v --fix

checks:
	go build ./...
	make generate
	make test
	make lint