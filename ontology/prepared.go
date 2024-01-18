package ontology

import "github.com/oxisto/owl2proto/owl"

type OntologyPrepared struct {
	Resources           map[string]*Resource
	SubClasses          map[string]owl.SubClassOf
	AnnotationAssertion map[string]owl.AnnotationAssertion
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
	Typ   string
	Value string
}

type ObjectRelationship struct {
	ObjectProperty string
	Class          string // IRI
	Name           string // Name of Class IRI
}
