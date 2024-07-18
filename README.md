# owl2proto

## Build

```bash
go build ./cmd/owl2proto/owl2proto.go
```

## Run 

```bash
./owl2proto generate-proto --root-resource-name=ex:Resource example/cloud.owx --header-file=example/example_header.proto --output-path=example/example.proto
```

## Generate Go Structs

Finally, go structs for the example can be created using `buf generate && buf format -w`.
