run:
	go run cmd/eldberg/*.go

proto:
	protoc --go_out=. --go-grpc_out=. internal/rpc/service.proto

format:
	gci -w --local github.com/AlexeySadkovich/eldberg ./

.PHONY: run, proto, format
