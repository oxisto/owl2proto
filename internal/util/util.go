package util

import (
	"regexp"
	"sort"
	"strings"

	"github.com/oxisto/owl2proto/ontology"
	"github.com/oxisto/owl2proto/owl"
)

const (
	Repeated = "repeated "
)

// GetObjectDetail returns the object type
func GetObjectDetail(s, rootResourceName string, resource *ontology.Resource, preparedOntology ontology.OntologyPrepared) (rep, typ, name string) {
	switch s {
	case "prop:hasMultiple", "prop:offersMultiple":
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
		return "string", "", resource.Name
	case "prop:parent":
		return "", "string", "parent_id"
	default:
		rep = ""
	}

	// If the object is a kind of the rootResourceName, the type is string and "_id" is added to the name to show that an ID is stored in the string.
	if isResourceAboveX(resource, preparedOntology, rootResourceName) && s != "prop:to" {
		return rep, "string", resource.Name + "_id"
	}

	return rep, resource.Name, resource.Name

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
		return "int64"
	case "xsd:dateTime":
		return "google.protobuf.Timestamp"
	case "xsd:java.time.ZonedDateTime":
		return "google.protobuf.Timestamp"
	case "xsd:java.util.ArrayList<Short>":
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
// TODO(all): FIx "CI/CD Service" to CICDService and cicd_service
func ToSnakeCase(s string) string {
	var (
		matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
		matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
	)

	s = CleanString(s)
	snake := matchFirstCap.ReplaceAllString(s, "${1}${2}")
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
