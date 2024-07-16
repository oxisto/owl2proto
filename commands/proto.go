package commands

import (
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/lmittmann/tint"
	"github.com/oxisto/owl2proto/internal/util"
	"github.com/oxisto/owl2proto/ontology"
)

type GenerateProtoCmd struct {
	GenerateCmd
	HeaderFile string
	OutputPath string `optional:"" default:"api/ontology.proto"`
	// DeterministicFieldNumbers is an option to enable deterministic field numbers based on a cryptographic hash. If
	// disabled, ascending field numbers are used sorted by parent class and then name
	DeterministicFieldNumbers bool `optional:"" default:"true"`
	// counter for generating the field number if ascending order is chosen
	i int
}

// createProto creates the proto file
func (cmd *GenerateProtoCmd) createProto(header string) string {
	output := ""

	// Add "auto-generated" header
	output += "// Auto-generated code by owl2proto (https://github.com/oxisto/owl2proto)"

	//Add header
	output += "\n\n" + header

	// Add EnumValueOptions
	output += `
extend google.protobuf.MessageOptions {
	repeated string resource_type_names = 60000;
}`

	// Sort preparedOntology.Resources map keys
	resourceMapKeys := util.SortMapKeys(cmd.preparedOntology.Resources)

	// Create proto messages with comments
	for _, rmk := range resourceMapKeys {
		class := cmd.preparedOntology.Resources[rmk]

		// is the counter for the message field numbers
		cmd.i = 0
		// i := 1

		// Add message comment
		if len(class.SubResources) == 0 {
			output += fmt.Sprintf("\n// %s is an entity class in our ontology. It can be instantiated and contains all of its properties as well of its implemented interfaces.", class.Name)
		} else {
			output += fmt.Sprintf("\n// %s is an abstract class in our ontology, it cannot be instantiated but acts as an \"interface\".", class.Name)
		}

		// Add comment
		for _, w := range class.Comment {
			output += "\n// " + w
		}

		// Start message
		output += fmt.Sprintf("\nmessage %s {\n", class.Name)

		if len(class.SubResources) == 0 {
			// Add class hierarchy as message options
			output = cmd.addClassHierarchy(output, rmk)

			// Add data properties, e.g., "bool enabled", "int64 interval", "int64 retention_period"
			output = cmd.addDataProperties(output, rmk)

			// Add object properties, e.g., "string compute_id", "ApplicationLogging application_logging", "TransportEncryption transport_encrypton"
			output = cmd.addObjectProperties(output, rmk)
		} else {
			// Get all leafs from object property and write it as 'oneOf {}'
			leafs := findAllLeafs(class.Iri, cmd.preparedOntology)
			// begin oneof X {}
			output += fmt.Sprintf("\n\toneof %s {", "type")
			for _, v := range leafs {
				var fieldNumber = 0
				fieldNumber, cmd.i = util.GetFieldNumber(cmd.DeterministicFieldNumbers, cmd.i, cmd.getResourceTypeList(v)...)
				output += fmt.Sprintf("\n\t\t%s %s = %d;", v.Name, util.ToSnakeCase(v.Name), fieldNumber)
			}

			// close oneOf{}
			output += "\n\t}"
		}

		// Close message
		output += "\n}\n"
	}

	return output

}

// addObjectProperties adds all object properties for the given resource to the output string
// Object properties (e.g., "AccessRestriction access_restriction", "HttpEndpoint http_endpoint", "TransportEncryption transport_encryption")
func (cmd *GenerateProtoCmd) addObjectProperties(output, rmk string) string {
	var fieldNumber = 0
	// Get all data properties of the given resource (rmk) and the parent resources
	objectProperties := cmd.findAllObjectProperties(rmk)

	// Sort slice of object properties
	sort.Slice(objectProperties, func(i, j int) bool {
		a := objectProperties[i]
		b := objectProperties[j]
		return a.Name < b.Name
	})

	// Create output for the object properties
	for _, o := range objectProperties {
		resourceTypeList := cmd.getResourceTypeList(cmd.preparedOntology.Resources[rmk])

		// Get field number
		resourceTypeList = append(resourceTypeList, o.Name)
		fieldNumber, cmd.i = util.GetFieldNumber(cmd.DeterministicFieldNumbers, cmd.i, resourceTypeList...)

		if o.Name != "" && o.ObjectProperty != "" {
			value, typ, name := cmd.preparedOntology.GetObjectDetail(o.ObjectPropertyName, cmd.preparedOntology.Resources[o.Class])
			if value != "" && typ != "" {
				output += fmt.Sprintf("\n\t%s%s %s  = %d;", value, typ, util.ToSnakeCase(name), fieldNumber)
			} else if typ != "" && name != "" {
				output += fmt.Sprintf("\n\t%s %s = %d;", typ, util.ToSnakeCase(name), fieldNumber)
			}
		}
	}

	return output
}

// findAllLeafs returns a resource list of all leaf nodes of a given resource/class
func findAllLeafs(class string, preparedOntology *ontology.OntologyPrepared) []*ontology.Resource {
	var leafs []*ontology.Resource

	r := preparedOntology.Resources[class]

	if len(r.SubResources) == 0 {
		leafs = append(leafs, r)
	} else {
		for _, s := range r.SubResources {
			leafs = append(leafs, findAllLeafs(s.Iri, preparedOntology)...)
		}
	}

	return leafs
}

// findAllObjectProperties adds all object properties for the given entity and the parents
func (cmd *GenerateProtoCmd) findAllObjectProperties(rmk string) []*ontology.ObjectRelationship {
	var (
		objectRelationsships []*ontology.ObjectRelationship
		parent               string
	)

	// "owl.Thing" is the root of the ontology and is not needed for the protobuf files. We do not have it in the prepared Ontology, so we have to skip, if we come to "owl.Thing".
	res, ok := cmd.preparedOntology.Resources[rmk]
	if !ok {
		return nil
	}
	objectRelationsships = append(objectRelationsships, res.ObjectRelationship...)

	parent = cmd.preparedOntology.Resources[rmk].Parent
	if parent == "" || rmk == cmd.preparedOntology.RootResourceName {
		return objectRelationsships
	} else {
		objectRelationsships = append(objectRelationsships, cmd.findAllObjectProperties(parent)...)
	}

	return objectRelationsships
}

// addObjectProperties adds all data properties for the given resource to the output string
// Data properties (e.g., "bool enabled", "int64 interval", "int64 retention_period")
func (cmd *GenerateProtoCmd) addDataProperties(output, rmk string) string {
	// Get all data properties of the given resource (rmk) and the parent resources
	dataProperties := cmd.preparedOntology.FindAllDataProperties(rmk)

	// Sort slice of data properties
	sort.Slice(dataProperties, func(i, j int) bool {
		a := dataProperties[i]
		b := dataProperties[j]
		return a.Value < b.Value
	})

	// Create output for the data properties
	for _, r := range dataProperties {
		if r.Typ != "" && r.Value != "" {
			var opts string = ""
			var fieldNumber = 0

			// Get list of resource types for given  object
			resourceTypeList := cmd.getResourceTypeList(cmd.preparedOntology.Resources[rmk])

			// Get field number
			resourceTypeList = append(resourceTypeList, r.Value)
			fieldNumber, cmd.i = util.GetFieldNumber(cmd.DeterministicFieldNumbers, cmd.i, resourceTypeList...)

			// Make name and id mandatory
			// TODO(oxisto): somehow extract this out of the ontology file itself which fields have constraints
			if r.Value == "name" || r.Value == "id" {
				opts = "[ (buf.validate.field).required = true ]"
			}

			// Add data property comment if available
			if r.Comment != "" {
				output += fmt.Sprintf("\n\t// %s", r.Comment)
			}

			output += fmt.Sprintf("\n\t%s %s = %d%s;", r.Typ, util.ToSnakeCase(r.Value), fieldNumber, opts)
		}
	}

	return output
}

func (cmd *GenerateProtoCmd) addClassHierarchy(output, rmk string) string {
	for _, typ := range cmd.getResourceTypeList(cmd.preparedOntology.Resources[rmk]) {
		output += fmt.Sprintf("\toption (resource_type_names) = \"%s\";\n", typ)
	}

	return output
}

// getResourceTypeList returns a list of all derived resources
func (cmd *GenerateProtoCmd) getResourceTypeList(resource *ontology.Resource) []string {
	var resource_types []string

	if resource == nil {
		return nil
	}

	if resource.Parent == "" {
		return []string{resource.Name}
	} else {
		resource_types = append(resource_types, resource.Name)
		resource_types = append(resource_types, cmd.getResourceTypeList(cmd.preparedOntology.Resources[resource.Parent])...)
	}

	return resource_types
}

func (cmd *GenerateProtoCmd) Run() (err error) {
	cmd.prepare()

	// Read header content from file
	b, err := os.ReadFile(cmd.HeaderFile)
	if err != nil {
		slog.Error("error reading header file", "location", cmd.HeaderFile, tint.Err(err))
		return nil
	}

	// Generate proto content
	output := cmd.createProto(string(b))

	// Write proto content to file
	err = util.WriteFile(cmd.OutputPath, output)
	if err != nil {
		slog.Error("error writing proto file to storage", tint.Err(err))
	}

	slog.Info("proto file written to storage", slog.String("output folder", cmd.OutputPath))
	return
}
