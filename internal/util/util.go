package util

import (
	"strings"

	"github.com/oxisto/owl2proto/ontology"
	"github.com/oxisto/owl2proto/owl"
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
	case "prop:proxyTarget":
		return "string", ""
	case "prop:parent":
		return "", ""
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

// GetProtoType converts Ontology type to golang type
func GetProtoType(s string) string {
	switch s {
	case "xsd:boolean":
		return "bool"
	case "xsd:String", "xsd:string", "xsd:de.fraunhofer.aisec.cpg.graph.Node", "xsd:de.fraunhofer.aisec.cpg.graph.statements.expressions.CallExpression", "xsd:de.fraunhofer.aisec.cpg.graph.statements.expressions.Expression", "xsd:de.fraunhofer.aisec.cpg.graph.declarations.FunctionDeclaration":
		return "string"
	case "xsd:java.util.ArrayList<String>", "java.util.List<de.fraunhofer.aisec.cpg.graph.declarations.TranslationUnitDeclaration>", "java.util.List<de.fraunhofer.aisec.cpg.graph.statements.expressions.CallExpression>", "xsd:java.util.List<de.fraunhofer.aisec.cpg.graph.statements.expressions.CallExpression>", "xsd:java.util.List<de.fraunhofer.aisec.cpg.graph.declarations.TranslationUnitDeclaration>":
		return "repeated string"
	case "xsd:integer", "xsd:int":
		return "int32"
	case "xsd:Short":
		return "uint32"
	case "xsd:float":
		return "float"
	case "xsd:java.time.Duration", "xsd:dateTime":
		return "int64"
	case "xsd:java.time.ZonedDateTime":
		return "google.protobuf.Timestamp"
	case "xsd:java.util.ArrayList<Short>":
		return "repeated uint32"
	case "xsd:java.util.Map<String, String>": // TODO(oxisto): Do we want to use here maps, as in the CPG?
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
