package ontology

import (
	"strings"

	"github.com/oxisto/owl2proto/internal/util"
)

// isResourceAboveX checks if a resource above the given resource has the name of rootResourceName
func isResourceAboveX(resource *Resource, preparedOntology *OntologyPrepared, rootResourceName string) bool {
	if resource == nil {
		return false
	}
	if resource.Parent == "" {
		return false
	}

	if GetNameFromIri(resource.Parent) == rootResourceName {
		return true
	}

	if isResourceAboveX(preparedOntology.Resources[resource.Parent], preparedOntology, rootResourceName) {
		return true
	}

	return false
}

// GetNameFromIri gets the last part of the IRI
func GetNameFromIri(s string) string {
	if s == "" {
		return ""
	}
	split := strings.Split(s, "/")

	return split[len(split)-1]
}

// GetDataPropertyAbbreviatedIriName returns the abbreviatedIRI name, e.g. "prop:enabled" returns "enabled"
func GetDataPropertyAbbreviatedIriName(s string) string {
	if s == "" {
		return ""
	}

	split := strings.Split(s, ":")

	return split[1]
}

// GetObjectDetail returns the object type
func (ont *OntologyPrepared) GetObjectDetail(s string, resource *Resource) (rep, typ, name string) {
	rName := resource.Name
	switch s {
	case "hasMultiple", "offersMultiple":
		rep = util.Repeated
	case "has", "runsOn", "offers", "storage":
		rep = ""
	case "to":
		rep = util.Repeated
	case "collectionOf":
		rep = util.Repeated
	case "offersInterface":
		rep = ""
	case "proxyTarget":
		return "string", "", rName
	case "parent":
		return "", "optional string", "parent_id"
	default:
		rep = ""
	}

	// If the object is a kind of the rootResourceName, the type is string and "_id" is added to the name to show that an ID is stored in the string.
	if isResourceAboveX(resource, ont, ont.RootResourceName) {
		// if the property is repeated, than use "ids"
		if rep == "" {
			return rep, "optional string", rName + "_id"
		} else {
			return rep, "string", rName + "_ids"
		}
	}

	// if the property is repeated add "s" to the name
	if rep == util.Repeated {
		name = util.ToPlural(rName)
	} else {
		name = rName
	}

	return rep, rName, name
}
