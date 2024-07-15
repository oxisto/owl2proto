package main

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/oxisto/owl2proto/internal/util"
	"github.com/oxisto/owl2proto/ontology"
	"github.com/oxisto/owl2proto/owl"

	"github.com/lmittmann/tint"
)

// prepareOntology extracts important information from the owl ontology file that is needed for the protobuf file creation
func (cmd *GenerateProtoCmd) prepareOntology() {
	for _, c := range cmd.owlOntology.Declarations {
		// Prepare ontology classes
		// We set the name extracted from the IRI and the IRI. If a name label exists we will change the name later.
		if c.Class.IRI != "" {
			cmd.preparedOntology.Resources[c.Class.IRI] = &ontology.Resource{
				Iri:  c.Class.IRI,
				Name: util.CleanString(util.GetNameFromIri(c.Class.IRI)),
			}
		}

		// Prepare ontology data properties
		if c.DataProperty.IRI != "" {
			cmd.preparedOntology.AnnotationAssertion[c.DataProperty.IRI] = &ontology.AnnotationAssertion{
				IRI:  c.DataProperty.IRI,
				Name: util.CleanString(util.GetNameFromIri(c.DataProperty.IRI)),
			}
		} else if c.DataProperty.AbbreviatedIRI != "" {
			cmd.preparedOntology.AnnotationAssertion[c.DataProperty.AbbreviatedIRI] = &ontology.AnnotationAssertion{
				IRI:  c.DataProperty.AbbreviatedIRI,
				Name: util.CleanString(util.GetDataPropertyAbbreviatedIriName(c.DataProperty.AbbreviatedIRI)),
			}
		}
	}

	// Prepare name and comment
	for _, aa := range cmd.owlOntology.AnnotationAssertion {
		if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:label" {
			if _, ok := cmd.preparedOntology.Resources[aa.IRI]; ok {
				cmd.preparedOntology.Resources[aa.IRI].Name = util.CleanString(aa.Literal)
			} else if _, ok := cmd.preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI]; ok {
				cmd.preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Name = util.CleanString(aa.Literal)
			}
		} else if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:comment" {
			if _, ok := cmd.preparedOntology.Resources[aa.IRI]; ok {
				c := cmd.preparedOntology.Resources[aa.IRI].Comment
				c = append(c, aa.Literal)
				cmd.preparedOntology.Resources[aa.IRI].Comment = c
			} else if _, ok := cmd.preparedOntology.AnnotationAssertion[aa.IRI]; ok {
				c := cmd.preparedOntology.AnnotationAssertion[aa.IRI].Comment
				c = append(c, aa.Literal)
				cmd.preparedOntology.AnnotationAssertion[aa.IRI].Comment = c
			} else if _, ok := cmd.preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI]; ok {
				c := cmd.preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Comment
				c = append(c, aa.Literal)
				cmd.preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Comment = c
			}
		}
	}

	// Prepare SubClasses
	// There are 5 different structures of SubClasses. All Class properties are IRIs:
	// - 2 Classes: The second Class is the parent of the first Class
	// - Class and DataSomeValuesFrom: Class is the current resource and DataSomeValuesFrom contains the Datatype (e.g., xsd:string) and the corresponding DataProperty/variable name as IRI or abbreviatedIRI (e.g., filename as IRI or prop:enabeld as abbreviatedIRI)
	// - Class and DataHasValue: Class is the current resource and DataHasValue is the same as DataSomeValuesFrom except, that no Datatype exists in the Ontology, but an Literal (Literal is a string, that is not used as an Ontology object/property).
	// - Class and ObjectSomeValuesFrom: Class is the current resource and ObjectSomeValuesFrom contains the ObjectProperty (e.g., prop:hasMultiple) and the linked resource (Class)
	// - Class and ObjectHasValue: Class is the current resource and ObjectHasValue contains the ObjectProperty IRI (e.g., "http://graph.clouditor.io/classes/scope") and a named individual (e.g., "http://graph.clouditor.io/classes/resourceId")
	for _, sc := range cmd.owlOntology.SubClasses {
		if len(sc.Class) == 2 {

			// "owl.Thing" is the root of the ontology and is not needed for the protobuf files
			if sc.Class[1].IRI != "owl.Thing" {
				// Create resource that has a parent. All resources directly under "owl.Thing" are alread created before (via the Declarations)
				r := &ontology.Resource{
					Iri:     sc.Class[0].IRI,
					Name:    cmd.preparedOntology.Resources[sc.Class[0].IRI].Name,
					Parent:  sc.Class[1].IRI,
					Comment: cmd.preparedOntology.Resources[sc.Class[0].IRI].Comment,
				}

				// Add subresources to the parent resource
				if val, ok := cmd.preparedOntology.Resources[sc.Class[1].IRI]; ok {
					if val.SubResources == nil {
						cmd.preparedOntology.Resources[sc.Class[1].IRI].SubResources = make([]*ontology.Resource, 0)
					}
					cmd.preparedOntology.Resources[sc.Class[1].IRI].SubResources = append(cmd.preparedOntology.Resources[sc.Class[1].IRI].SubResources, r)
				}

				// Add parent IRI to resource (not subresource!). We couldn't do this beforehand (Declarations) because we only get the information here,
				cmd.preparedOntology.Resources[sc.Class[0].IRI].Parent = sc.Class[1].IRI
			}
		} else if sc.DataSomeValuesFrom != nil {
			// Add data values, e.g. "enabled xsd:bool" ("enabled" is a data property and "xsd:bool" is a datatype) or
			for _, v := range sc.DataSomeValuesFrom {
				var (
					comment string
				)
				// Check if comment is available
				if val, ok := cmd.preparedOntology.AnnotationAssertion[v.DataProperty.IRI]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				} else if val, ok := cmd.preparedOntology.AnnotationAssertion[v.DataProperty.AbbreviatedIRI]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				// Get DataProperty name
				cmd.preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(cmd.preparedOntology.Resources[sc.Class[0].IRI].Relationship, &ontology.Relationship{
					IRI:     v.DataProperty.IRI,
					Typ:     util.GetProtoType(v.Datatype.AbbreviatedIRI),
					Value:   util.GetDataPropertyIRIName(v.DataProperty, cmd.preparedOntology),
					Comment: comment,
				})

			}
		} else if sc.DataHasValue != nil {
			// Add data values, e.g. "interval xsd:java.time.Duration" ("interval" is a data property and "xsd:java.time.Duration" is Literal/string)
			for _, v := range sc.DataHasValue {
				var (
					comment    string
					identifier string
				)

				// Get IRI or abbreviatedIRI from DataProperty
				if v.DataProperty.AbbreviatedIRI != "" {
					identifier = v.DataProperty.AbbreviatedIRI
				} else if v.DataProperty.IRI != "" {
					identifier = v.DataProperty.IRI
				}

				// Check if comment is available
				if val, ok := cmd.preparedOntology.AnnotationAssertion[identifier]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				cmd.preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(cmd.preparedOntology.Resources[sc.Class[0].IRI].Relationship, &ontology.Relationship{
					IRI:     identifier,
					Typ:     util.GetProtoType(v.Literal),
					Value:   util.GetDataPropertyIRIName(v.DataProperty, cmd.preparedOntology),
					Comment: comment,
				})

			}
		} else if sc.ObjectSomeValuesFrom != nil {
			// Add object values, e.g., "offers ResourceLogging"
			for _, v := range sc.ObjectSomeValuesFrom {
				if v.ObjectProperty.IRI != "" {
					cmd.preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship = append(cmd.preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship, &ontology.ObjectRelationship{
						ObjectProperty:     v.ObjectProperty.IRI,
						ObjectPropertyName: util.GetObjectPropertyIRIName(v.ObjectProperty, cmd.preparedOntology),
						Class:              v.Class.IRI,
						Name:               cmd.preparedOntology.Resources[v.Class.IRI].Name,
					})
				} else if v.ObjectProperty.AbbreviatedIRI != "" {
					cmd.preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship = append(cmd.preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship, &ontology.ObjectRelationship{
						ObjectProperty:     v.ObjectProperty.AbbreviatedIRI,
						ObjectPropertyName: util.GetObjectPropertyIRIName(v.ObjectProperty, cmd.preparedOntology),
						Class:              v.Class.IRI,
						Name:               cmd.preparedOntology.Resources[v.Class.IRI].Name,
					})
				}
			}
		} else if sc.ObjectHasValue != nil {
			for _, v := range sc.ObjectHasValue {
				// Add object value, e.g., "scope resourceId"
				var (
					comment         string
					identifier      string
					namedIndividual string
				)

				// Get IRI or abbreviatedIRI from ObjectProperty
				if v.ObjectProperty.AbbreviatedIRI != "" {
					identifier = v.ObjectProperty.AbbreviatedIRI
				} else if v.ObjectProperty.IRI != "" {
					identifier = v.ObjectProperty.IRI
				}

				// Get IRI or abbreviatedIRI from NamedIndividual
				if v.NamedIndividual.AbbreviatedIRI != "" {
					namedIndividual = v.NamedIndividual.AbbreviatedIRI
				} else if v.NamedIndividual.IRI != "" {
					namedIndividual = v.NamedIndividual.IRI
				}

				// Check if comment is available
				if val, ok := cmd.preparedOntology.AnnotationAssertion[identifier]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				cmd.preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(cmd.preparedOntology.Resources[sc.Class[0].IRI].Relationship, &ontology.Relationship{
					IRI:     identifier,
					Typ:     util.GetProtoType(namedIndividual),
					Value:   util.GetObjectPropertyIRIName(v.ObjectProperty, cmd.preparedOntology),
					Comment: comment,
				})
			}
		}
	}
}

// createProto creates the protobuf file
func (cmd *GenerateProtoCmd) createProto(header string) string {
	output := ""

	// Add "auto-generated" header
	output += "// Auto-generated code by owl2proto (https://github.com/oxisto/owl2proto)"

	//Add header
	output += "\n\n" + header

	// Add EnumValueOptions
	output += `
extend google.protobuf.MessageOptions {
	repeated string resource_type_names = 50000;
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
				fieldNumber, cmd.i = util.GetFieldNumber(cmd.FieldNumberOption, cmd.i, cmd.getResourceTypeList(v)...)
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
// Object properties (e.g., "AccessRestriction access_restriction", "HttpEndpoing http_endpoint", "TransportEncryption transport_encryption")
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
		fieldNumber, cmd.i = util.GetFieldNumber(cmd.FieldNumberOption, cmd.i, resourceTypeList...)

		if o.Name != "" && o.ObjectProperty != "" {
			value, typ, name := util.GetObjectDetail(o.ObjectProperty, cmd.preparedOntology.RootResourceName, cmd.preparedOntology.Resources[o.Class], cmd.preparedOntology)
			if value != "" && typ != "" {
				output += fmt.Sprintf("\n\t%s%s %s  = %d;", value, typ, util.ToSnakeCase(name), fieldNumber)
			} else if typ != "" && name != "" {
				output += fmt.Sprintf("\n\t%s %s = %d;", typ, util.ToSnakeCase(name), fieldNumber)
			}
		}
	}

	return output
}

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

	objectRelationsships = append(objectRelationsships, cmd.preparedOntology.Resources[rmk].ObjectRelationship...)

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
			fieldNumber, cmd.i = util.GetFieldNumber(cmd.FieldNumberOption, cmd.i, resourceTypeList...)

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

func writeFile(outputFile, s string) error {
	var err error

	// TODO(all):Create folder if not exists
	// Create storage file
	f, err := os.Create(outputFile)
	if err != nil {
		err = fmt.Errorf("error creating file: %v", err)
		slog.Error(err.Error())
	}

	// Write output string to file
	_, err = f.WriteString(s)
	if err != nil {
		err = fmt.Errorf("error writing output to file: %v", err)
		slog.Error(err.Error())
		f.Close()
		return err
	}

	// Close storage file
	err = f.Close()
	if err != nil {
		err = fmt.Errorf("error closing file: %v", err)
		slog.Error(err.Error())
		return err
	}

	return nil
}

// getResourceTypeList returns a list of all derived resources
func (cmd *GenerateProtoCmd) getResourceTypeList(resource *ontology.Resource) []string {
	var resource_types []string

	if resource.Parent == "" {
		return []string{resource.Name}
	} else {
		resource_types = append(resource_types, resource.Name)
		resource_types = append(resource_types, cmd.getResourceTypeList(cmd.preparedOntology.Resources[resource.Parent])...)
	}

	return resource_types
}

var cli struct {
	GenerateProto GenerateProtoCmd `cmd:"" help:"Generates proto files."`
	GenerateUML   GenerateUMLCmd   `cmd:"" help:"Generates proto files."`
}

type GenerateCmd struct {
	OwlFile          string `arg:""`
	preparedOntology *ontology.OntologyPrepared
	RootResourceName string `required:""`
	// true for deterministic field numbers, false for ascending field numbers
	FieldNumberOption bool `required:""`
}

type GenerateProtoCmd struct {
	GenerateCmd
	owlOntology *owl.Ontology
	HeaderFile  string
	OutputPath  string `optional:"" default:"api/ontology.proto"`
	// counter for generating the field number if ascending order is chosen
	i int
}

type GenerateUMLCmd struct {
	GenerateCmd
	OutputPath string `optional:"" default:"api/ontology.puml"`
}

func (cmd *GenerateProtoCmd) prepare() {
	var (
		b   []byte
		err error
	)

	cmd.FieldNumberOption = cmd.FieldNumberOption
	cmd.RootResourceName = cmd.RootResourceName

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
	err = xml.Unmarshal(b, &cmd.owlOntology)
	if err != nil {
		slog.Error("error while unmarshalling XML", tint.Err(err))
		return
	}

	// prepareOntology
	cmd.prepareOntology()
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
	err = writeFile(cmd.OutputPath, output)
	if err != nil {
		slog.Error("error writing proto file to storage", tint.Err(err))
	}

	slog.Info("proto file written to storage", slog.String("output folder", cmd.OutputPath))
	return
}

// func (cmd *GenerateUMLCmd) Run() (err error) {
// 	cmd.prepare()

// 	// Generate UML
// 	output := owl2proto.CreatePlantUMLFile(cmd.preparedOntology)

// 	// Write UML
// 	err = writeFile(cmd.OutputPath, output)
// 	if err != nil {
// 		slog.Error("error writing UML file to storage", tint.Err(err))
// 	}

// 	slog.Info("UML file written to storage", slog.String("output folder", cmd.OutputPath))
// 	return
// }

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError())
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
