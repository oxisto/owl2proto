package util

import (
	"strings"

	"github.com/oxisto/owl2protobuf/ontology"
)

const (
	Repeated = "repeated "
)

// GetObjectDetail returns the object type
func GetObjectDetail(s, rootResourceName string, resource *ontology.Resource, preparedOntology ontology.OntologyPrepared) (string, string) {
	var (
		value string
	)

	switch s {
	case "prop:hasMultiple", "prop:offersMultiple":
		value = Repeated
	case "prop:has", "prop:runsOn", "prop:to", "prop:offers", "prop:storage":
		value = ""
	case "prop:collectionOf":
		value = Repeated
	case "prop:offersInterface":
		value = ""
	case "prop:proxyTarget", "prop:parent":
		return "string", ""
	default:
		return s, ""
	}

	if isResourceAboveX(resource, preparedOntology, rootResourceName) {
		return value, "ResourceID"
	}

	return value, resource.Name

}

// isResourceAboveX checks if a resource above the given resource has the name of rootResourceName
func isResourceAboveX(resource *ontology.Resource, preparedOntology ontology.OntologyPrepared, rootResourceName string) bool {
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

// GetGoType converts Ontology type to golang type
func GetGoType(s string) string {
	switch s {
	case "xsd:string":
		return "string"
	case "xsd:boolean":
		return "bool"
	case "xsd:dateTime":
		return "time.duration"
	case "xsd:integer":
		return "int"
	case "xsd:float":
		return "float32"
	case "xsd:java.time.Duration":
		return "time.duration"
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

func GetDataPropertyName(s string) string {
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
// TODO(all): FIx "CI/CD Service" to CICDService and cicd_service
func ToSnakeCase(s string) string {
	var result string

	s = CleanString(s)

	for i, char := range s {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result += "_"
		}

		result += string(char)
	}

	return strings.ToLower(result)
}
