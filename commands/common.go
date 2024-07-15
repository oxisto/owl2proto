package commands

import (
	"encoding/xml"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/oxisto/owl2proto/ontology"
	"github.com/oxisto/owl2proto/owl"
)

type GenerateCmd struct {
	OwlFile          string `arg:""`
	RootResourceName string `required:""`
	preparedOntology *ontology.OntologyPrepared
}

func (cmd *GenerateCmd) prepare() {
	var (
		b   []byte
		err error
		ont owl.Ontology
	)

	// Set up logging
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	// Read Ontology XML
	b, err = os.ReadFile(cmd.OwlFile)
	if err != nil {
		slog.Error("error reading ontology file", "location", cmd.OwlFile, tint.Err(err))
		return
	}

	// Unmarshal file content (Ontology xml file)
	err = xml.Unmarshal(b, &ont)
	if err != nil {
		slog.Error("error while un-marshalling XML", tint.Err(err))
		return
	}

	cmd.preparedOntology = ontology.Prepare(&ont, cmd.RootResourceName)
}
