package main

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

var file string

type Ontology struct {
	Declarations []Declaration `xml:"Declaration"`
}

type Declaration struct {
	Class Class `xml:"Class"`
}

type Class struct {
	IRI string `xml:"IRI,attr"`
}

func main() {
	var (
		b   []byte
		err error
		o   Ontology
	)

	file = os.Args[1]

	// Set up logging
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	// Read XML
	b, err = os.ReadFile(file)
	if err != nil {
		slog.Error("error reading file", tint.Err(err))
		return
	}

	err = xml.Unmarshal(b, &o)
	if err != nil {
		slog.Error("error while unmarshalling XML", tint.Err(err))
		return
	}

	fmt.Printf("%+v", o)
}
