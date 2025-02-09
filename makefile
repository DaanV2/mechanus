tools:
	go install github.com/spf13/cobra-cli@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	go install github.com/google/wire/cmd/wire@latest

proto:
	buf generate

# If not working: docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf format --write
proto-format:
	buf format --write

image:
	docker build . -t mechanus

server:
	go build -o mechanus ./server/main.go