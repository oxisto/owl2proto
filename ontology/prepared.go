package ontology

import (
	"github.com/oxisto/owl2proto/owl"
)

type OntologyPrepared struct {
	Resources           map[string]*Resource
	SubClasses          map[string]owl.SubClassOf
	AnnotationAssertion map[string]*AnnotationAssertion

	RootResourceName string
}

type Resource struct {
	Iri                string
	Name               string
	Parent             string
	Comment            []string
	Relationship       []*Relationship
	ObjectRelationship []*ObjectRelationship
	SubResources       []*Resource
}

type Relationship struct {
	IRI     string
	Typ     string
	Value   string
	Comment string
}

type ObjectRelationship struct {
	ObjectProperty     string
	ObjectPropertyName string
	Class              string // IRI
	Name               string // Name of Class IRI
	Comment            string // Comment of the property
}

type AnnotationAssertion struct {
	IRI     string
	Name    string
	Comment []string
}

// FindAllDataProperties adds all object properties for the given entity and the parents
func (po *OntologyPrepared) FindAllDataProperties(key string) []*Relationship {
	var (
		relationships []*Relationship
		parent        string
	)

	relationships = append(relationships, po.Resources[key].Relationship...)

	parent = po.Resources[key].Parent
	if parent == "" || key == po.RootResourceName {
		return relationships
	} else {
		relationships = append(relationships, po.FindAllDataProperties(parent)...)
	}

	return relationships
}
