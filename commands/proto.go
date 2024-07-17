package commands

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

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

	// FullSemanticMode determines whether all semantic meta-data is emitted as options (see owl/owl.proto), such as all
	// IRIs, prefixes for the ontology, etc. If disabled, only a very condensed form of the class hierarchy is emitted
	// as protobuf options.
	FullSemanticMode bool `optional:"" default:"true"`

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

	output += cmd.emitOptionsHeader()

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

// emitOptionsHeader includes all semantic ontology metadata as options when full semantic mode is enabled; otherwise, it adds a streamlined version of the ontology class hierarchy.
func (cmd *GenerateProtoCmd) emitOptionsHeader() string {
	var output string

	if cmd.FullSemanticMode {
		// Add import
		output += `
import "owl/owl.proto";

`
		// Prepare prefix output
		var prefixOutputs []string

		for short, prefix := range cmd.preparedOntology.Prefixes {
			prefixOutputs = append(prefixOutputs, `{
	prefix: "`+short+`"
	iri: "`+prefix.IRI+`"
}`)
		}

		// Add ontology meta-data
		output += `
option (owl.meta) = {
	prefixes: [` + strings.Join(prefixOutputs, ",") + `]};
`

	} else {
		// Add EnumValueOptions
		output += `
extend google.protobuf.MessageOptions {
	repeated string resource_type_names = 60000;
}`

	}

	return output
}

// emitClassOptions adds the class options IRI and parent when full semantic mode is enabled, otherwise it adds only the resource type name.
func (cmd *GenerateProtoCmd) emitClassOptions(iri string) string {
	var (
		output string
		class  *ontology.Resource
	)

	class = cmd.preparedOntology.Resources[iri]

	if cmd.FullSemanticMode {
		output += fmt.Sprintf("\toption (owl.class).iri = \"%s\";\n", cmd.preparedOntology.AbbreviateIRI(iri))
		for _, parentIri := range cmd.getParents(class) {
			output += fmt.Sprintf("\toption (owl.class).parent = \"%s\";\n", cmd.preparedOntology.AbbreviateIRI(parentIri))
		}
	} else {
		for _, typ := range cmd.getResourceTypeList(class) {
			output += fmt.Sprintf("\toption (resource_type_names) = \"%s\";\n", typ)
		}
	}

	return output
}

// emitPropertyOptions adds the property options IRI, parent and class IRI when full semantic mode is enabled.
func (cmd *GenerateProtoCmd) emitPropertyOptions(r *ontology.Relationship) string {
	var (
		opts       []string
		optsOutput string
	)
	// Make name and id mandatory
	// TODO(oxisto): somehow extract this out of the ontology file itself which fields have constraints
	if r.Value == "name" || r.Value == "id" {
		opts = append(opts, "(buf.validate.field).required = true")
	}

	if cmd.FullSemanticMode {
		opts = append(opts, fmt.Sprintf("(owl.property).iri = \"%s\"", cmd.preparedOntology.AbbreviateIRI(r.IRI)))
		// TODO(oxisto): Emit all the real property parents
		opts = append(opts, fmt.Sprintf("(owl.property).parent = \"%s\"", "owl:topDataProperty"))
		opts = append(opts, fmt.Sprintf("(owl.property).class_iri = \"%s\"", cmd.preparedOntology.AbbreviateIRI(r.From)))
	}

	if len(opts) > 0 {
		optsOutput = fmt.Sprintf(" [ %s ]", strings.Join(opts, ",\n\t"))
	}

	return optsOutput
}

func (cmd *GenerateProtoCmd) emitObjectPropertyOptions(r *ontology.ObjectRelationship) string {
	var (
		opts       []string
		optsOutput string
	)

	if cmd.FullSemanticMode {
		opts = append(opts, fmt.Sprintf("(owl.property).iri = \"%s\"", cmd.preparedOntology.AbbreviateIRI(r.ObjectProperty)))
		// TODO(oxisto): Emit all the real property parents
		opts = append(opts, fmt.Sprintf("(owl.property).parent = \"%s\"", "owl:topObjectProperty"))
		opts = append(opts, fmt.Sprintf("(owl.property).class_iri = \"%s\"", cmd.preparedOntology.AbbreviateIRI(r.From)))
	}

	if len(opts) > 0 {
		optsOutput = fmt.Sprintf(" [ %s ]", strings.Join(opts, ",\n\t"))
	}

	return optsOutput
}

// addObjectProperties adds all object properties for the given resource to the output string
// Object properties (e.g., "AccessRestriction access_restriction", "HttpEndpoint http_endpoint", "TransportEncryption transport_encryption")
func (cmd *GenerateProtoCmd) addObjectProperties(output, rmk string) string {
	var (
		fieldNumber = 0
		optsOutput  string
	)

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

		optsOutput = cmd.emitObjectPropertyOptions(o)

		if o.Name != "" && o.ObjectProperty != "" {
			value, typ, name := cmd.preparedOntology.GetObjectDetail(o.ObjectPropertyName, cmd.preparedOntology.Resources[o.To])
			if value != "" && typ != "" {
				output += fmt.Sprintf("\n\t%s%s %s  = %d%s;", value, typ, util.ToSnakeCase(name), fieldNumber, optsOutput)
			} else if typ != "" && name != "" {
				output += fmt.Sprintf("\n\t%s %s = %d%s;", typ, util.ToSnakeCase(name), fieldNumber, optsOutput)
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
func (cmd *GenerateProtoCmd) findAllObjectProperties(iri string) []*ontology.ObjectRelationship {
	var (
		objectRelationships []*ontology.ObjectRelationship
		parent              string
	)

	res, ok := cmd.preparedOntology.Resources[iri]
	if !ok {
		slog.Error("Could not find entity", "iri", iri)
		return nil
	}

	objectRelationships = append(objectRelationships, res.ObjectRelationship...)

	parent = cmd.preparedOntology.Resources[iri].Parent
	if parent == "" || iri == cmd.preparedOntology.RootResourceName {
		return objectRelationships
	} else {
		objectRelationships = append(objectRelationships, cmd.findAllObjectProperties(parent)...)
	}

	return objectRelationships
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
			var (
				optsOutput  string
				fieldNumber = 0
			)

			// Get list of resource types for given  object
			resourceTypeList := cmd.getResourceTypeList(cmd.preparedOntology.Resources[rmk])

			// Get field number
			resourceTypeList = append(resourceTypeList, r.Value)
			fieldNumber, cmd.i = util.GetFieldNumber(cmd.DeterministicFieldNumbers, cmd.i, resourceTypeList...)

			optsOutput = cmd.emitPropertyOptions(r)

			// Add data property comment if available
			if r.Comment != "" {
				output += fmt.Sprintf("\n\t// %s", r.Comment)
			}

			output += fmt.Sprintf("\n\t%s %s = %d%s;", r.Typ, util.ToSnakeCase(r.Value), fieldNumber, optsOutput)
		}
	}

	return output
}

func (cmd *GenerateProtoCmd) addClassHierarchy(output, iri string) string {
	output += cmd.emitClassOptions(iri)

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

// getParents returns a list of all parent IRIs
func (cmd *GenerateProtoCmd) getParents(resource *ontology.Resource) []string {
	var iris []string

	if resource.Parent == "" {
		return []string{"owl:Thing"}
	} else {
		iris = append(iris, resource.Parent)
		iris = append(iris, cmd.getParents(cmd.preparedOntology.Resources[resource.Parent])...)
	}

	return iris
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
