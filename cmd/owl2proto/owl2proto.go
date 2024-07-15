package main

import (
	"github.com/alecthomas/kong"
	"github.com/oxisto/owl2proto/commands"
)

var cli struct {
	GenerateProto commands.GenerateProtoCmd `cmd:"" help:"Generates proto files."`
	GenerateUML   commands.GenerateUMLCmd   `cmd:"" help:"Generates proto files."`
}

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError())
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
