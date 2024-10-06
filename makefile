tools:
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install github.com/spf13/cobra-cli@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

proto-v1:
	protoc --go_out=./server/internal/grpc/v1 \
		--go-grpc_out=./server/internal/grpc/v1 \
		shared/proto/v1/*.proto

proto: proto-v1