package util

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/cespare/xxhash/v2"
)

const (
	Repeated = "repeated "
)

// ToPlural returns the plural of a string
func ToPlural(s string) string {
	// if last character is "y", change to "ies"
	if s[len(s)-1:] == "y" {
		return s[:len(s)-1] + "ies"
	} else {
		return s + "s"
	}

}

// GetProtoType converts Ontology type to golang type
func GetProtoType(s string) string {
	switch s {
	case "xsd:boolean":
		return "bool"
	case "xsd:String", "xsd:string", "xsd:de.fraunhofer.aisec.cpg.graph.Node", "xsd:de.fraunhofer.aisec.cpg.graph.statements.expressions.CallExpression", "xsd:de.fraunhofer.aisec.cpg.graph.statements.expressions.Expression", "xsd:de.fraunhofer.aisec.cpg.graph.declarations.FunctionDeclaration", "http://graph.clouditor.io/classes/resourceId":
		return "string"
	case "xsd:listString", "xsd:java.util.ArrayList<String>", "java.util.List<de.fraunhofer.aisec.cpg.graph.declarations.TranslationUnitDeclaration>", "java.util.List<de.fraunhofer.aisec.cpg.graph.statements.expressions.CallExpression>", "xsd:java.util.List<de.fraunhofer.aisec.cpg.graph.statements.expressions.CallExpression>", "xsd:java.util.List<de.fraunhofer.aisec.cpg.graph.declarations.TranslationUnitDeclaration>":
		return "repeated string"
	case "xsd:integer", "xsd:int":
		return "int32"
	case "xsd:Short":
		return "uint32"
	case "xsd:float":
		return "float"
	case "xsd:java.time.Duration":
		return "google.protobuf.Duration"
	case "xsd:dateTime", "xsd:java.time.ZonedDateTime":
		return "google.protobuf.Timestamp"
	case "xsd:java.util.ArrayList<Short>":
		// Note, there is no uint16 in protobuf, therefore we need to resort to uint32.
		return "repeated uint32"
	case "xsd:java.util.Map<String, String>":
		return "map<string, string>"
	default:
		return s
	}

}

// CleanString deletes spaces and /.
func CleanString(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")

	return s
}

// ToSnakeCase converts camel case to snake case and deletes spaces
func ToSnakeCase(s string) string {
	var (
		matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
		matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
	)

	s = CleanString(s)
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// SortMapKeys returns the keys of the map sorted ba [sort.Strings].
func SortMapKeys[V any](m map[string]V) []string {
	resources := make([]string, 0, len(m))

	for k := range m {
		resources = append(resources, k)
	}

	// Sort slice by key
	sort.Strings(resources)

	return resources
}

// GetFieldNumber returns a "consistent" field number for the proto field based on the input strings if
// deterministicFieldNumbers is true, otherwise it returns the incremented counter input (ascending field numbers). The
// maximum field number is 18999. The first return value is the field number and the second is the counter i
func GetFieldNumber(deterministicFieldNumbers bool, counter int, input ...string) (int, int) {
	if deterministicFieldNumbers {
		hash := xxhash.Sum64([]byte(strings.Join(input, "")))

		// the maximum field number is 18999, because the numbers 19000 to 19999 are reserved for the Protocol Buffers implementation
		number := int(hash%19000) + 1

		return number, counter
	} else {
		counter++
		if counter >= 19000 {
			slog.Error("field number '%s' is to high", slog.Int("counter", counter))
			os.Exit(1)
		}
		return counter, counter
	}
}

func WriteFile(outputFile, s string) error {
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
