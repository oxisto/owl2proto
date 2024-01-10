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

var (
	file     string
	resource map[string][]*Resource
)

type Resource struct {
	iri          string
	name         string
	parent       string
	comment      []string
	relationship []*Relationship
}

type Relationship struct {
	property string
	value    string
}

func parseOntology(o owl.Ontology) map[string][]*Resource {
	resource = make(map[string][]*Resource)

	for _, sc := range o.SubClasses {
		if len(sc.Class) == 2 {
			if sc.Class[1].IRI == "owl.Thing" {
				resource["Thing"] = append(resource["owl.Thing"], &Resource{
					iri:  sc.Class[0].IRI,
					name: getNameFromIri(sc.Class[0].IRI),
				})
			} else {
				resource[sc.Class[1].IRI] = append(resource[sc.Class[1].IRI], &Resource{
					iri:    sc.Class[0].IRI,
					name:   getNameFromIri(sc.Class[0].IRI),
					parent: sc.Class[1].IRI,
				})
			}
		}
	}

	return resource
}

func getNameFromIri(s string) string {
	if s == "" {
		return ""
	}
	split := strings.Split(s, "/")

	return split[4]
}

func toSnakeCase(s string) string {
	var result string

	for i, char := range s {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result += "_"
		}

		result += string(char)
	}

	return strings.ToLower(result)
}

func createProtoFile(input map[string][]*Resource) string {
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
	for k, v := range resource {
		output += fmt.Sprintf("message %s {", getNameFromIri(k))

		// Check if the resource has additional subresources
		i := 100
		if len(v) > 0 {
			output += "\n\toneof type {"
			for _, v2 := range v {
				i += 1

				output += fmt.Sprintf("\n\t\t%s %s = %d;", v2.name, toSnakeCase(v2.name), i)

			}
			output += "\n\t}"
		}

		output += "\n}\n\n"
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

	// Parse Ontology
	resource := parseOntology(o)

	// Generate protobuf file
	output := createProtoFile(resource)

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
