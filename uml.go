package owl2proto

import (
	"fmt"

	"github.com/oxisto/owl2proto/internal/util"
	"github.com/oxisto/owl2proto/ontology"
)

func CreatePlantUMLFile(po ontology.OntologyPrepared) string {
	var output string

	output += "@startuml ontology\n"
	output += "/' Auto-generated code by owl2proto (https://github.com/oxisto/owl2proto) '/\n"

	// Sort preparedOntology.Resources map keys
	resourceMapKeys := util.SortMapKeys(po.Resources)

	// Create proto messages with comments
	for _, iri := range resourceMapKeys {
		class := po.Resources[iri]

		// Start class
		output += fmt.Sprintf("\nclass %s {\n", class.Name)

		// Add data properties, e.g., "bool enabled", "int64 interval", "int64 retention_period"
		output = addDataProperties(output, iri, po)

		// End class
		output += "}\n"

		// Draw relationships. First our parent
		parent, ok := po.Resources[class.Parent]
		if ok {
			output += fmt.Sprintf("\n%s <|-- %s\n", parent.Name, class.Name)
		}

		// Then, draw all object relationships
		for _, prop := range class.ObjectRelationship {
			output += fmt.Sprintf("\n%s <-- %s : %s\n", prop.Name, class.Name, prop.ObjectPropertyName)
		}
	}

	output += "@enduml"

	return output
}

// addObjectProperties adds all data properties for the given resource to the output string
func addDataProperties(output, iri string, po ontology.OntologyPrepared) string {
	// TODO(oxisto): Types; for now everything is untyped

	// Get all only the properties of the given resource
	props := po.Resources[iri].Relationship

	for _, prop := range props {
		output += fmt.Sprintf("\t%s\n", prop.Value)
	}

	return output
}
