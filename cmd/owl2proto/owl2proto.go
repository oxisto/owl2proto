package main

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/oxisto/owl2proto/internal/util"
	"github.com/oxisto/owl2proto/ontology"
	"github.com/oxisto/owl2proto/owl"

	"github.com/lmittmann/tint"
)

var (
	owlFile          string
	headerFile       string
	outputPath       string
	rootResourceName string
)

const (
	DefaultOutputPath = "api/ontology.proto"
)

// prepareOntology extracts important information from the owl ontology file that is needed for the protobuf file creation
func prepareOntology(o owl.Ontology) ontology.OntologyPrepared {
	preparedOntology := ontology.OntologyPrepared{
		Resources:           make(map[string]*ontology.Resource),
		SubClasses:          make(map[string]owl.SubClassOf),
		AnnotationAssertion: make(map[string]*ontology.AnnotationAssertion),
	}

	for _, c := range o.Declarations {
		// Prepare ontology classes
		// We set the name extracted from the IRI and the IRI. If a name label exists we will change the name later.
		if c.Class.IRI != "" {
			preparedOntology.Resources[c.Class.IRI] = &ontology.Resource{
				Iri:  c.Class.IRI,
				Name: util.CleanString(util.GetNameFromIri(c.Class.IRI)),
			}
		}

		// Prepare ontology data properties
		if c.DataProperty.IRI != "" {
			preparedOntology.AnnotationAssertion[c.DataProperty.IRI] = &ontology.AnnotationAssertion{
				IRI:  c.DataProperty.IRI,
				Name: util.CleanString(util.GetNameFromIri(c.DataProperty.IRI)),
			}
		} else if c.DataProperty.AbbreviatedIRI != "" {
			preparedOntology.AnnotationAssertion[c.DataProperty.AbbreviatedIRI] = &ontology.AnnotationAssertion{
				IRI:  c.DataProperty.AbbreviatedIRI,
				Name: util.CleanString(util.GetDataPropertyAbbreviatedIriName(c.DataProperty.AbbreviatedIRI)),
			}
		}
	}

	// Prepare name and comment
	for _, aa := range o.AnnotationAssertion {
		if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:label" {
			if _, ok := preparedOntology.Resources[aa.IRI]; ok {
				preparedOntology.Resources[aa.IRI].Name = util.CleanString(aa.Literal)
			} else if _, ok := preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI]; ok {
				preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Name = util.CleanString(aa.Literal)
			}
		} else if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:comment" {
			if _, ok := preparedOntology.Resources[aa.IRI]; ok {
				c := preparedOntology.Resources[aa.IRI].Comment
				c = append(c, aa.Literal)
				preparedOntology.Resources[aa.IRI].Comment = c
			} else if _, ok := preparedOntology.AnnotationAssertion[aa.IRI]; ok {
				c := preparedOntology.AnnotationAssertion[aa.IRI].Comment
				c = append(c, aa.Literal)
				preparedOntology.AnnotationAssertion[aa.IRI].Comment = c
			} else if _, ok := preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI]; ok {
				c := preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Comment
				c = append(c, aa.Literal)
				preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Comment = c
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
	for _, sc := range o.SubClasses {
		if len(sc.Class) == 2 {

			// "owl.Thing" is the root of the ontology and is not needed for the protobuf files
			if sc.Class[1].IRI != "owl.Thing" {
				// Create resource that has a parent. All resources directly under "owl.Thing" are alread created before (via the Declarations)
				r := &ontology.Resource{
					Iri:     sc.Class[0].IRI,
					Name:    preparedOntology.Resources[sc.Class[0].IRI].Name,
					Parent:  sc.Class[1].IRI,
					Comment: preparedOntology.Resources[sc.Class[0].IRI].Comment,
				}

				// Add subresources to the parent resource
				if val, ok := preparedOntology.Resources[sc.Class[1].IRI]; ok {
					if val.SubResources == nil {
						preparedOntology.Resources[sc.Class[1].IRI].SubResources = make([]*ontology.Resource, 0)
					}
					preparedOntology.Resources[sc.Class[1].IRI].SubResources = append(preparedOntology.Resources[sc.Class[1].IRI].SubResources, r)
				}

				// Add parent IRI to resource (not subresource!). We couldn't do this beforehand (Declarations) because we only get the information here,
				preparedOntology.Resources[sc.Class[0].IRI].Parent = sc.Class[1].IRI
			}
		} else if sc.DataSomeValuesFrom != nil {
			// Add data values, e.g. "enabled xsd:bool" ("enabled" is a data property and "xsd:bool" is a datatype) or
			for _, v := range sc.DataSomeValuesFrom {
				var (
					comment string
				)
				// Check if comment is available
				if val, ok := preparedOntology.AnnotationAssertion[v.DataProperty.IRI]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				} else if val, ok := preparedOntology.AnnotationAssertion[v.DataProperty.AbbreviatedIRI]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				// Get DataProperty name
				preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(preparedOntology.Resources[sc.Class[0].IRI].Relationship, &ontology.Relationship{
					IRI:     v.DataProperty.IRI,
					Typ:     util.GetProtoType(v.Datatype.AbbreviatedIRI),
					Value:   util.GetDataPropertyIRIName(v.DataProperty, preparedOntology),
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
				if val, ok := preparedOntology.AnnotationAssertion[identifier]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(preparedOntology.Resources[sc.Class[0].IRI].Relationship, &ontology.Relationship{
					IRI:     identifier,
					Typ:     util.GetProtoType(v.Literal),
					Value:   util.GetDataPropertyIRIName(v.DataProperty, preparedOntology),
					Comment: comment,
				})

			}
		} else if sc.ObjectSomeValuesFrom != nil {
			// Add object values, e.g., "offers ResourceLogging"
			for _, v := range sc.ObjectSomeValuesFrom {
				if v.ObjectProperty.IRI != "" {
					preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship = append(preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship, &ontology.ObjectRelationship{
						ObjectProperty: v.ObjectProperty.IRI,
						Class:          v.Class.IRI,
						Name:           preparedOntology.Resources[v.Class.IRI].Name,
					})
				} else if v.ObjectProperty.AbbreviatedIRI != "" {
					preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship = append(preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship, &ontology.ObjectRelationship{
						ObjectProperty: v.ObjectProperty.AbbreviatedIRI,
						Class:          v.Class.IRI,
						Name:           preparedOntology.Resources[v.Class.IRI].Name,
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
				if val, ok := preparedOntology.AnnotationAssertion[identifier]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(preparedOntology.Resources[sc.Class[0].IRI].Relationship, &ontology.Relationship{
					IRI:     identifier,
					Typ:     util.GetProtoType(namedIndividual),
					Value:   util.GetObjectPropertyIRIName(v.ObjectProperty, preparedOntology),
					Comment: comment,
				})
			}
		}
	}

	return preparedOntology
}

// createProtoFile creates the protobuf file
func createProtoFile(preparedOntology ontology.OntologyPrepared, header string) string {
	output := ""

	// Add "auto-generated" header
	output += "// Auto-generated code by owl2proto (https://github.com/oxisto/owl2proto)"

	//Add header
	output += "\n\n" + header

	// Add EnumValueOptions
	output += `
extend google.protobuf.EnumValueOptions {
	optional string resource_type_name = 123456789;
}`

	// Sort preparedOntology.Resources map keys
	resourceMapKeys := util.SortMapKeys(preparedOntology.Resources)

	// Add ResourceType enum
	output += `
enum ResourceType {
	RESOURCE_TYPE_UNSPECIFIED = 0;`

	// Add all resource type entries
	// i is the counter for the enum field numbers
	i := 0
	for _, rmk := range resourceMapKeys {
		i += 1
		resourceTypeList := getResourceTypeList(preparedOntology.Resources[rmk], &preparedOntology)

		// For examle, ABAC has the resource types "ABAC,Authorization,SecurityFeature" and is presented as RESOURCE_ABAC_AUTHORIZATION_SECURITYFEATURE.
		// TODO(all): Or do we want instead RESOURCE_ABAC_AUTHORIZATION_SECURITY_FEATURE?
		output += fmt.Sprintf("\n\tRESOURCE_TYPE_%s = %d [(resource_type_name) = \"%s\"];", strings.ToUpper(strings.Join(resourceTypeList, "_")), i, strings.Join(resourceTypeList, ","))
	}

	// Close ResourceType enum
	output += "}\n"

	// Create proto messages with comments
	for _, rmk := range resourceMapKeys {
		class := preparedOntology.Resources[rmk]

		if len(class.SubResources) == 0 {
			// is the counter for the message field numbers
			i := 1

			// Add message comment
			output += fmt.Sprintf("\n// %s is an entity class in our ontology.", class.Name)

			// Add comment
			for _, w := range class.Comment {
				output += "\n// " + w
			}

			// Start message
			output += fmt.Sprintf("\nmessage %s {", class.Name)

			// We only add properties for "leaf" nodes
			if len(class.SubResources) == 0 {
				// Add data properties, e.g., "bool enabled", "int64 interval", "int64 retention_period"
				output, i = addDataProperties(output, rmk, i, preparedOntology)

				// Add object properties, e.g., "string compute_id", "ApplicationLogging application_logging", "TransportEncryption transport_encrypton"
				output, _ = addObjectProperties(output, rmk, i, preparedOntology)
			} else {
				// Otherwise, we add sub-classes

				// j is the counter for the oneof field numbers
				j := 100
				output += "\n\toneof type {"
				// Sort slice of sub-resources
				sort.Slice(class.SubResources, func(i, j int) bool {
					a := class.SubResources[i]
					b := class.SubResources[j]
					return a.Name < b.Name
				})
				for _, v2 := range class.SubResources {
					j += 1
					output += fmt.Sprintf("\n\t\t%s %s = %d;", v2.Name, util.ToSnakeCase(v2.Name), j)

				}
				output += "\n\t}"
			}

			// Close message
			output += "\n}\n"
		}
	}

	return output

}

// addObjectProperties adds all object properties for the given resource to the output string
// Object properties (e.g., "AccessRestriction access_restriction", "HttpEndpoing http_endpoint", "TransportEncryption transport_encryption")
func addObjectProperties(output, rmk string, i int, preparedOntology ontology.OntologyPrepared) (string, int) {
	// Get all data properties of the given resource (rmk) and the parent resources
	objectProperties := findAllObjectProperties(rmk, preparedOntology)

	// Sort slice of object properties
	sort.Slice(objectProperties, func(i, j int) bool {
		a := objectProperties[i]
		b := objectProperties[j]
		return a.Name < b.Name
	})

	// Create output for the object properties
	for _, o := range objectProperties {
		if o.Name != "" && o.ObjectProperty != "" {
			value, typ, name := util.GetObjectDetail(o.ObjectProperty, rootResourceName, preparedOntology.Resources[o.Class], preparedOntology)
			if value != "" && typ != "" {
				output += fmt.Sprintf("\n\t%s%s %s  = %d;", value, typ, util.ToSnakeCase(name), i)
				i += 1
			} else if typ != "" && name != "" {
				output += fmt.Sprintf("\n\t%s %s = %d;", typ, util.ToSnakeCase(name), i)
				i += 1
			}
		}
	}

	return output, i
}

// findAllObjectProperties adds all object properties for the given entity and the parents
func findAllObjectProperties(rmk string, preparedOntology ontology.OntologyPrepared) []*ontology.ObjectRelationship {
	var (
		objectRelationsships []*ontology.ObjectRelationship
		parent               string
	)

	objectRelationsships = append(objectRelationsships, preparedOntology.Resources[rmk].ObjectRelationship...)

	parent = preparedOntology.Resources[rmk].Parent
	if parent == "" || rmk == rootResourceName {
		return objectRelationsships
	} else {
		objectRelationsships = append(objectRelationsships, findAllObjectProperties(parent, preparedOntology)...)
	}

	return objectRelationsships
}

// addObjectProperties adds all data properties for the given resource to the output string
// Data properties (e.g., "bool enabled", "int64 interval", "int64 retention_period")
func addDataProperties(output, rmk string, i int, preparedOntology ontology.OntologyPrepared) (string, int) {
	// Get all data properties of the given resource (rmk) and the parent resources
	dataProperties := findAllDataProperties(rmk, preparedOntology)

	// Sort slice of data properties
	sort.Slice(dataProperties, func(i, j int) bool {
		a := dataProperties[i]
		b := dataProperties[j]
		return a.Value < b.Value
	})

	// Create output for the data properties
	for _, r := range dataProperties {
		if r.Typ != "" && r.Value != "" {
			// Add data property comment if available
			if r.Comment != "" {
				output += fmt.Sprintf("\n\t// %s", r.Comment)
			}
			output += fmt.Sprintf("\n\t%s %s = %d;", r.Typ, util.ToSnakeCase(r.Value), i)
			i += 1
		}
	}

	return output, i
}

// findAllDataProperties adds all object properties for the given entity and the parents
func findAllDataProperties(rmk string, preparedOntology ontology.OntologyPrepared) []*ontology.Relationship {
	var (
		relationships []*ontology.Relationship
		parent        string
	)

	relationships = append(relationships, preparedOntology.Resources[rmk].Relationship...)

	parent = preparedOntology.Resources[rmk].Parent
	if parent == "" || rmk == rootResourceName {
		return relationships
	} else {
		relationships = append(relationships, findAllDataProperties(parent, preparedOntology)...)
	}

	return relationships
}

func writeProtofileToStorage(outputFile, s string) error {
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
func getResourceTypeList(resource *ontology.Resource, preparedOntology *ontology.OntologyPrepared) []string {
	var resource_types []string

	if resource.Parent == "" {
		return []string{resource.Name}
	} else {
		resource_types = append(resource_types, resource.Name)
		resource_types = append(resource_types, getResourceTypeList(preparedOntology.Resources[resource.Parent], preparedOntology)...)
	}

	return resource_types
}

func main() {
	var (
		b   []byte
		err error
		o   owl.Ontology
	)

	if len(os.Args) < 4 {
		slog.Error("not enough command line arguments given",
			slog.String("arguments needed",
				"owl file location, header file location, root resource name from owl file (e.g., http://graph.clouditor.io/classes/CloudResource) and output path (optional, default is 'api/ontology.proto'"),
		)

		return
	}
	owlFile = os.Args[1]
	headerFile = os.Args[2]
	rootResourceName = os.Args[3]

	// Check if output path is given as argument
	if len(os.Args) >= 5 {
		outputPath = os.Args[4]
	} else {
		outputPath = DefaultOutputPath
	}

	// Set up logging
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	// Read Ontology XML
	b, err = os.ReadFile(owlFile)
	if err != nil {
		slog.Error("error reading ontology file", "location", owlFile, tint.Err(err))
		return
	}

	err = xml.Unmarshal(b, &o)
	if err != nil {
		slog.Error("error while unmarshalling XML", tint.Err(err))
		return
	}

	// Read header content from file
	b, err = os.ReadFile(headerFile)
	if err != nil {
		slog.Error("error reading header file", "location", headerFile, tint.Err(err))
		return
	}

	// prepareOntology
	preparedOntology := prepareOntology(o)

	// Generate proto content
	output := createProtoFile(preparedOntology, string(b))

	// Write proto content to file
	err = writeProtofileToStorage(outputPath, output)
	if err != nil {
		slog.Error("error writing proto file to storage", tint.Err(err))
	}

	slog.Info("proto file written to storage", slog.String("output folder", outputPath))
}
