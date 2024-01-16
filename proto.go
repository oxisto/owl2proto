package owl2protobuf

//go:generate buf generate
//go:generate buf generate --template buf.openapi.gen.yaml --path api/ -o openapi/
