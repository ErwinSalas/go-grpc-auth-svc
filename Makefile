proto:
	protoc -I=./proto --go_out=./proto --go_opt=paths=source_relative --go-grpc_out=./proto --go-grpc_opt=paths=source_relative ./proto/auth.proto

server:
	go run cmd/main.go

cert:
	sh cert/gen.sh  