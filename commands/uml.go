package commands

import (
	"log/slog"

	"github.com/lmittmann/tint"
	"github.com/oxisto/owl2proto"
)

type GenerateUMLCmd struct {
	GenerateCmd
	OutputPath string `optional:"" default:"api/ontology.puml"`
}

func (cmd *GenerateUMLCmd) Run() (err error) {
	cmd.prepare()

	// Generate UML
	output := owl2proto.CreatePlantUMLFile(cmd.preparedOntology)

	// Write UML
	err = writeFile(cmd.OutputPath, output)
	if err != nil {
		slog.Error("error writing UML file to storage", tint.Err(err))
	}

	slog.Info("UML file written to storage", slog.String("output folder", cmd.OutputPath))
	return
}
