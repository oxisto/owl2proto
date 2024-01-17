package main

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/oxisto/owl2protobuf/internal/util"
	"github.com/oxisto/owl2protobuf/ontology"
	"github.com/oxisto/owl2protobuf/owl"

	"github.com/lmittmann/tint"
)

// TODOs
// - get label instead of iri for the name fields
// - add data/object property comments
// - add ObjectHasValue
// - add cli

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
				preparedOntology.Resources[aa.AbbreviatedIRI].Name = util.CleanString(aa.Literal)
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
			}
		}
	}

	// Prepare SubClasses
	// There are 4 different structures of SubClasses. All Class properties are IRIs:
	// - 2 Classes: The second Class is the parent of the first Class
	// - Class and DataSomeValuesFrom: Class is the current resource and DataSomeValuesFrom contains the Datatype (e.g., xsd:string) and the corresponding DataProperty/variable name as IRI or abbreviatedIRI (e.g., filename as IRI or prop:enabeld as abbreviatedIRI)
	// - Class and DataHasValue: Class is the current resource and DataHasValue is the same as DataSomeValuesFrom except, that no Datatype exists in the Ontology, but an Literal (Literal is a string, that is not used as an Ontology object/property).
	// - Class and ObjectSomeValuesFrom: Class is the current resource and ObjectSomeValuesFrom contains the ObjectProperty (e.g., prop:hasMultiple) and the linked resource (Class)
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
				preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship = append(preparedOntology.Resources[sc.Class[0].IRI].ObjectRelationship, &ontology.ObjectRelationship{
					ObjectProperty: v.ObjectProperty.AbbreviatedIRI,
					Class:          v.Class.IRI,
					Name:           preparedOntology.Resources[v.Class.IRI].Name,
				})
			}
		}
	}

	return preparedOntology
}

// createProtoFile creates the protobuf file
func createProtoFile(preparedOntology ontology.OntologyPrepared, header string) string {
	output := ""

	//Add header
	output += header

	// Create proto message for ResourceID
	output += `

message ResourceID {
	repeated string resource_id = 1;
}
`

	// Create proto messages with comments
	for _, v := range preparedOntology.Resources {
		if v.Iri == "http://graph.clouditor.io/classes/ApplicationLog" {
			fmt.Println("Stop here")
		}
		// is the counter for the message field numbers
		i := 0

		// Add comment
		for _, v := range v.Comment {
			output += "\n// " + v
		}

		// Start message
		output += fmt.Sprintf("\nmessage %s {", v.Name)

		// Add data properties
		for _, r := range v.Relationship {
			if r.Typ != "" && r.Value != "" {
				i += 1
				// Add data property comment if available
				if r.Comment != "" {
					output += fmt.Sprintf("\n\t// %s", r.Comment)
				}
				output += fmt.Sprintf("\n\t%s %s  = %d;", r.Typ, r.Value, i)
			}
		}

		// Add object properties
		for _, o := range v.ObjectRelationship {
			if o.Name != "" && o.ObjectProperty != "" {
				i += 1
				value, typ := util.GetObjectDetail(o.ObjectProperty, rootResourceName, preparedOntology.Resources[o.Class], preparedOntology)
				output += fmt.Sprintf("\n\t%s%s %s  = %d;", value, typ, o.Name, i)
			}
		}

		// Add subresources if present
		// Important
		if len(v.SubResources) > 0 {
			// j is the counter for the oneof field numbers
			j := 100
			output += "\n\n\toneof type {"
			for _, v2 := range v.SubResources {
				j += 1

				output += fmt.Sprintf("\n\t\t%s %s = %d;", v2.Name, util.ToSnakeCase(v2.Name), j)

			}
			output += "\n\t}"
		}

		// Close message
		output += "\n}\n"
	}

	return output

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
