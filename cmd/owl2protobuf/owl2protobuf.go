package main

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/oxisto/owl2protobuf/pkg/owl"

	"github.com/lmittmann/tint"
)

// TODOs
// - get label instead of iri for the name fields
// - add comments
// - add relationsships

var (
	file string
)

func prepareOntology(o owl.Ontology) owl.OntologyPrepared {
	preparedOntology := owl.OntologyPrepared{
		Resources:           make(map[string]*owl.Resource),
		SubClasses:          make(map[string]owl.SubClassOf),
		AnnotationAssertion: make(map[string]owl.AnnotationAssertion),
	}

	// Add classes
	// We set the name extracted from the IRI and the IRI and if a name label exists we will change the name later
	for _, c := range o.Declarations {
		if c.Class.IRI != "" {
			preparedOntology.Resources[c.Class.IRI] = &owl.Resource{
				Iri:  c.Class.IRI,
				Name: getNameFromIri(c.Class.IRI),
			}
		}
	}

	// prepare resources map with  and name
	for _, aa := range o.AnnotationAssertion {
		if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:label" {
			if _, ok := preparedOntology.Resources[aa.IRI]; ok {
				preparedOntology.Resources[aa.IRI].Name = cleanString(aa.Literal)

			}
		} else if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:comment" {
			if _, ok := preparedOntology.Resources[aa.IRI]; ok {
				c := preparedOntology.Resources[aa.IRI].Comment
				c = append(c, aa.Literal)
				preparedOntology.Resources[aa.IRI].Comment = c

			}
		}
	}

	// Prepare SubClasses
	for _, sc := range o.SubClasses {
		if len(sc.Class) == 2 {

			if sc.Class[1].IRI != "owl.Thing" {
				r := &owl.Resource{
					Iri:     sc.Class[0].IRI,
					Name:    preparedOntology.Resources[sc.Class[0].IRI].Name,
					Parent:  sc.Class[1].IRI,
					Comment: preparedOntology.Resources[sc.Class[0].IRI].Comment,
				}

				if val, ok := preparedOntology.Resources[sc.Class[1].IRI]; ok {
					if val.SubResources == nil {
						preparedOntology.Resources[sc.Class[1].IRI].SubResources = make([]*owl.Resource, 0)
					}
					preparedOntology.Resources[sc.Class[1].IRI].SubResources = append(preparedOntology.Resources[sc.Class[1].IRI].SubResources, r)
				}
			}
		} else if sc.DataSomeValuesFrom != nil {
			for _, v := range sc.DataSomeValuesFrom {
				preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(preparedOntology.Resources[sc.Class[0].IRI].Relationship, &owl.Relationship{
					Typ:   v.DataType,
					Value: getDataPropertyName(v.DataProperty),
				})
			}
		}
	}

	return preparedOntology
}

// getNameFromIri gets the last part of the IRI
func getNameFromIri(s string) string {
	if s == "" {
		return ""
	}
	split := strings.Split(s, "/")

	return split[4]
}

func getDataPropertyName(s string) string {
	if s == "" {
		return ""
	}

	split := strings.Split(s, ":")

	return split[1]
}

// cleanString deletes spaces and /.
func cleanString(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")

	return s
}

// toSnakeCase converts camel case to snake case and deletes spaces
// TODO(all): FIx "CI/CD Service" to CICDService and cicd_service
func toSnakeCase(s string) string {
	var result string

	s = cleanString(s)

	for i, char := range s {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result += "_"
		}

		result += string(char)
	}

	return strings.ToLower(result)
}

func createProtoFile(preparedOntology owl.OntologyPrepared) string {
	output := ""

	//Add header and imports
	output = `
// Copyright 2024 Fraunhofer AISEC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//           $$\                           $$\ $$\   $$\
//           $$ |                          $$ |\__|  $$ |
//  $$$$$$$\ $$ | $$$$$$\  $$\   $$\  $$$$$$$ |$$\ $$$$$$\    $$$$$$\   $$$$$$\
// $$  _____|$$ |$$  __$$\ $$ |  $$ |$$  __$$ |$$ |\_$$  _|  $$  __$$\ $$  __$$\
// $$ /      $$ |$$ /  $$ |$$ |  $$ |$$ /  $$ |$$ |  $$ |    $$ /  $$ |$$ | \__|
// $$ |      $$ |$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |  $$ |$$\ $$ |  $$ |$$ |
// \$$$$$$\  $$ |\$$$$$   |\$$$$$   |\$$$$$$  |$$ |  \$$$   |\$$$$$   |$$ |
//  \_______|\__| \______/  \______/  \_______|\__|   \____/  \______/ \__|
//
// This file is part of Clouditor Community Edition.

syntax = "proto3";

package clouditor.discovery.v1;

import "google/api/annotations.proto";
import "google/protobuf/struct.proto";
import "tagger/tagger.proto";
import "validate/validate.proto";

option go_package = "api/discovery";
`

	// Add proto messages
	for k, v := range preparedOntology.Resources {
		// The resources with owl:Thing as parent can be skipped
		// We decided to let the key empty instead of adding "owl:Thing"
		if k == "" {
			continue
		}

		// Add comment
		for _, v := range v.Comment {
			output += "\n// " + v
		}

		// Create message
		output += fmt.Sprintf("\nmessage %s {", v.Name)

		// Add properties
		for _, r := range v.Relationship {
			if r.Typ != "" && r.Value != "" {
				output += fmt.Sprintf("\n\t%s %s"+r.Value, r.Typ)
			}
		}

		// Add subresources to proto resource message if present
		i := 100
		if len(v.SubResources) > 0 {
			output += "\n\toneof type {"
			for _, v2 := range v.SubResources {
				i += 1

				output += fmt.Sprintf("\n\t\t%s %s = %d;", v2.Name, toSnakeCase(v2.Name), i)

			}
			output += "\n\t}"
		}

		output += "\n}\n"
	}

	return output

}

func main() {
	var (
		b   []byte
		err error
		o   owl.Ontology
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
	// fmt.Printf("%+v", o)

	// prepareOntology
	preparedOntology := prepareOntology(o)

	// Generate protobuf file
	output := createProtoFile(preparedOntology)

	// Write protobuf file
	f, err := os.Create("output/ontology.proto")
	if err != nil {
		slog.Error("error creating file: %v", err)
	}

	_, err = f.WriteString(output)
	if err != nil {
		slog.Error("error writing output to file: %v", err)
		f.Close()
		return
	}
	err = f.Close()
	if err != nil {
		slog.Error("", tint.Err(err))
		return
	}
}
