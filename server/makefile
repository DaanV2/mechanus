test:
	go test ./... --cover -coverprofile=reports/coverage

coverage-report: test
	go tool cover -html=reports/coverage

generate:
	go generate ./...