package util

import (
	"log/slog"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/cespare/xxhash"
	"github.com/oxisto/owl2proto/ontology"
	"github.com/oxisto/owl2proto/owl"
)

const (
	Repeated = "repeated "
)

// GetObjectDetail returns the object type
func GetObjectDetail(s, rootResourceName string, resource *ontology.Resource, preparedOntology *ontology.OntologyPrepared) (rep, typ, name string) {
	rName := resource.Name
	switch s {
	case "prop:hasMultiple", "prop:offersMultiple", "http://graph.clouditor.io/classes/offersMultiple":
		rep = Repeated
	case "prop:has", "prop:runsOn", "prop:offers", "prop:storage":
		rep = ""
	case "prop:to":
		rep = Repeated
	case "prop:collectionOf":
		rep = Repeated
	case "prop:offersInterface":
		rep = ""
	case "prop:proxyTarget":
		return "string", "", rName
	case "prop:parent":
		return "", "optional string", "parent_id"
	default:
		rep = ""
	}

	// If the object is a kind of the rootResourceName, the type is string and "_id" is added to the name to show that an ID is stored in the string.
	if isResourceAboveX(resource, preparedOntology, rootResourceName) {
		// if the property is repeated, than use "ids"
		if rep == "" {
			return rep, "optional string", rName + "_id"
		} else {
			return rep, "string", rName + "_ids"
		}
	}

	// if the property is repeated add "s" to the name
	if rep == Repeated {
		name = toPlural(rName)
	} else {
		name = rName
	}

	return rep, rName, name

}

// toPlural return the plural of a string
func toPlural(s string) string {
	// if last character is "y", change to "ies"
	if s[len(s)-1:] == "y" {
		return s[:len(s)-1] + "ies"
	} else {
		return s + "s"
	}

}

// isResourceAboveX checks if a resource above the given resource has the name of rootResourceName
func isResourceAboveX(resource *ontology.Resource, preparedOntology *ontology.OntologyPrepared, rootResourceName string) bool {
	if resource == nil {
		return false
	}
	if resource.Parent == "" {
		return false
	}

	if resource.Parent == rootResourceName {
		return true
	}

	if isResourceAboveX(preparedOntology.Resources[resource.Parent], preparedOntology, rootResourceName) {
		return true
	}

	return false
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

// GetNameFromIri gets the last part of the IRI
func GetNameFromIri(s string) string {
	if s == "" {
		return ""
	}
	split := strings.Split(s, "/")

	return split[4]
}

// GetDataPropertyIRIName return the existing IRI (IRI vs. abbreviatedIRI) from the Data Property
func GetDataPropertyIRIName(prop owl.DataProperty, preparedOntology ontology.OntologyPrepared) string {
	// It is possible, that the IRI/abbreviatedIRI name is not correct, therefore we have to get the correct name from the preparedOntology. Otherwise, we get the name directly from the IRI/abbreviatedIRI
	if prop.AbbreviatedIRI != "" {
		if val, ok := preparedOntology.AnnotationAssertion[prop.AbbreviatedIRI]; ok {
			return val.Name
		} else {
			return GetDataPropertyAbbreviatedIriName(prop.AbbreviatedIRI)
		}
	} else if prop.IRI != "" {
		if val, ok := preparedOntology.Resources[prop.IRI]; ok {
			return val.Name
		} else {
			return GetNameFromIri(prop.IRI)
		}
	}

	return ""
}

// TODO(all): Use generic for GetObjectPropertyIRIName and GetDataPropertyIRIName
// GetObjectPropertyIRIName return the existing IRI (IRI vs. abbreviatedIRI) from the Data Property
func GetObjectPropertyIRIName(prop owl.ObjectProperty, preparedOntology ontology.OntologyPrepared) string {
	// It is possible, that the IRI/abbreviatedIRI name is not correct, therefore we have to get the correct name from the preparedOntology. Otherwise, we get the name directly from the IRI/abbreviatedIRI
	if prop.AbbreviatedIRI != "" {
		if val, ok := preparedOntology.AnnotationAssertion[prop.AbbreviatedIRI]; ok {
			return val.Name
		} else {
			return GetDataPropertyAbbreviatedIriName(prop.AbbreviatedIRI)
		}
	} else if prop.IRI != "" {
		if val, ok := preparedOntology.Resources[prop.IRI]; ok {
			return val.Name
		} else {
			return GetNameFromIri(prop.IRI)
		}
	}

	return ""
}

// GetDataPropertyAbbreviatedIriName returns the abbreviatedIRI name, e.g. "prop:enabled" returns "enabled"
func GetDataPropertyAbbreviatedIriName(s string) string {
	if s == "" {
		return ""
	}

	split := strings.Split(s, ":")

	return split[1]
}

// CleanString deletes spaces and /.
func CleanString(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")

	return s
}

// ToSnakeCase converts camel case to snake case and deletes spaces
// TODO(all): Fix "OSLogging" to OSLogging and os_logging
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

// SortMapKeys sorts the keys of the map
func SortMapKeys[V *ontology.Resource](m map[string]V) []string {
	resources := make([]string, 0, len(m))

	for k := range m {
		resources = append(resources, k)
	}

	// Sort slice by key
	sort.Strings(resources)

	return resources
}

// GetFieldNumber returns a "consistent" field number for the proto field based on the input strings if fieldNumberOption is true, otherwise it returns the incremented counter input (ascending field numbers). The maximum field number is 18999.
// The first return value is the field number and the second is the counter i
func GetFieldNumber(fieldNumberOption bool, counter int, input ...string) (int, int) {

	if fieldNumberOption {
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
