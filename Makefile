.PHONY: proto, gci

run:
	go run cmd/eldberg/*.go

proto:
	protoc --go_out=. --go-grpc_out=. internal/rpc/service.proto

gci:
	gci -w -local github.com/AlexeySadkovich/eldberg ./ 
