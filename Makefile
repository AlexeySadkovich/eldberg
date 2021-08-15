.PHONY: proto, gci

run:
	go run cmd/node/main.go

proto:
	protoc --go_out=. --go-grpc_out=. internal/rpc/service.proto

gci:
	gci -w ./
