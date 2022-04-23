protoc \
--go_out=./tinapi --go_opt=paths=source_relative \
--go-grpc_out=./tinapi --go-grpc_opt=paths=source_relative \
--proto_path=./proto ./proto/*.proto
