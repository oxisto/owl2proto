package fileHeaders

func GetImports() string {
	return `
import "google/api/annotations.proto";
import "google/protobuf/struct.proto";
import "tagger/tagger.proto";
import "validate/validate.proto";
`
}

func GetSyntax() string {
	return `
syntax = "proto3";
`
}
