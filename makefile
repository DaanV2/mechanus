proto:
	buf generate

# If not working: docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf format --write
proto-format:
	buf format --write

proto-lint:
	buf lint

checks:
	$(MAKE) proto
	$(MAKE) proto-format
	$(MAKE) proto-lint
	$(MAKE) image

image:
	docker build . -t mechanus

server:
	go build -o mechanus ./server/main.go