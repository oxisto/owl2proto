package ontology

import (
	"log/slog"
	"strings"

	"github.com/oxisto/owl2proto/internal/util"
	"github.com/oxisto/owl2proto/owl"
)

// OntologyPrepared contains an [owl.Ontology] in a way that is "prepared" for the translation to protobuf messages.
type OntologyPrepared struct {
	Resources           map[string]*Resource
	SubClasses          map[string]*owl.SubClassOf
	AnnotationAssertion map[string]*AnnotationAssertion
	NamedIndividual     map[string]*NamedIndividual

	Prefixes map[string]*owl.Prefix

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
	From    string // IRI
}

type ObjectRelationship struct {
	ObjectProperty     string
	ObjectPropertyName string
	From               string // IRI
	To                 string // IRI
	Name               string // Name of To IRI
	Comment            string // Comment of the property
}

type AnnotationAssertion struct {
	IRI     string
	Name    string
	Comment []string
}

type NamedIndividual struct {
	IRI  string
	Name string
	Type string
}

// FindAllDataProperties adds all object properties for the given entity and the parents
func (po *OntologyPrepared) FindAllDataProperties(iri string) []*Relationship {
	var (
		relationships []*Relationship
		parent        string
	)

	res, ok := po.Resources[iri]
	if !ok {
		slog.Error("Could not find entity", "iri", iri)
		return nil
	}

	relationships = append(relationships, res.Relationship...)

	parent = po.Resources[iri].Parent
	if parent == "" || iri == po.RootResourceName {
		return relationships
	} else {
		relationships = append(relationships, po.FindAllDataProperties(parent)...)
	}

	return relationships
}

// Prepare extracts important information from the owl ontology file that is needed for the protobuf file creation.
func Prepare(src *owl.Ontology, rootIRI string) *OntologyPrepared {
	preparedOntology := &OntologyPrepared{
		Prefixes:            map[string]*owl.Prefix{},
		Resources:           map[string]*Resource{},
		SubClasses:          map[string]*owl.SubClassOf{},
		AnnotationAssertion: map[string]*AnnotationAssertion{},
		NamedIndividual:     map[string]*NamedIndividual{},
		RootResourceName:    rootIRI,
	}

	for idx := range src.Prefixes {
		p := &src.Prefixes[idx]
		preparedOntology.Prefixes[p.Name] = p
	}

	// Make sure our root resource name is definitely not an abbreviated IRI anymore
	preparedOntology.RootResourceName = preparedOntology.normalizeAbbreviatedIRI(preparedOntology.RootResourceName)

	for _, c := range src.Declarations {
		iri := NormalizedIRI(preparedOntology, &c.Class.Entity)

		// Prepare ontology classes
		// We set the name extracted from the IRI and the IRI. If a name label exists we will change the name later.
		if iri != "" {
			preparedOntology.Resources[iri] = &Resource{
				Iri:  iri,
				Name: util.CleanString(GetNameFromIri(iri)),
			}
		}

		// Prepare ontology data properties
		if c.DataProperty.IRI != "" {
			preparedOntology.AnnotationAssertion[c.DataProperty.IRI] = &AnnotationAssertion{
				IRI:  c.DataProperty.IRI,
				Name: util.CleanString(GetNameFromIri(c.DataProperty.IRI)),
			}
		} else if c.DataProperty.AbbreviatedIRI != "" {
			preparedOntology.AnnotationAssertion[c.DataProperty.AbbreviatedIRI] = &AnnotationAssertion{
				IRI:  c.DataProperty.AbbreviatedIRI,
				Name: util.CleanString(GetDataPropertyAbbreviatedIriName(c.DataProperty.AbbreviatedIRI)),
			}
		}

		// Prepare ontology named individuals
		if c.NamedIndividual.IRI != "" {
			preparedOntology.NamedIndividual[c.NamedIndividual.IRI] = &NamedIndividual{
				IRI:  c.NamedIndividual.IRI,
				Name: util.CleanString(GetNameFromIri(c.NamedIndividual.IRI)),
			}
		} else if c.NamedIndividual.AbbreviatedIRI != "" {
			preparedOntology.NamedIndividual[c.NamedIndividual.AbbreviatedIRI] = &NamedIndividual{
				IRI:  c.NamedIndividual.AbbreviatedIRI,
				Name: util.CleanString(GetDataPropertyAbbreviatedIriName(c.NamedIndividual.AbbreviatedIRI)),
			}
		}
	}

	// Prepare name and comment
	for _, aa := range src.AnnotationAssertion {
		// Prepare name from "rdfs:label"
		if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:label" {
			if _, ok := preparedOntology.Resources[NormalizedIRI(preparedOntology, aa)]; ok {
				preparedOntology.Resources[NormalizedIRI(preparedOntology, aa)].Name = util.CleanString(aa.Literal)
			} else if _, ok := preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI]; ok {
				preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Name = util.CleanString(aa.Literal)
			}
		}

		// Prepare comment from "rdfs:comment"
		if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:comment" { // Prepare comment from "rdfs:comment"
			if _, ok := preparedOntology.Resources[NormalizedIRI(preparedOntology, aa)]; ok {
				c := preparedOntology.Resources[NormalizedIRI(preparedOntology, aa)].Comment
				c = append(c, aa.Literal)
				preparedOntology.Resources[NormalizedIRI(preparedOntology, aa)].Comment = c
			} else if _, ok := preparedOntology.AnnotationAssertion[aa.IRI]; ok {
				c := preparedOntology.AnnotationAssertion[aa.IRI].Comment
				c = append(c, aa.Literal)
				preparedOntology.AnnotationAssertion[aa.IRI].Comment = c
			} else if _, ok := preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI]; ok {
				c := preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Comment
				c = append(c, aa.Literal)
				preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Comment = c
			}
		}

		// Prepare type for named individuals from "rdfs:seeAlso"
		if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:seeAlso" {
			if _, ok := preparedOntology.NamedIndividual[NormalizedIRI(preparedOntology, aa)]; ok {
				preparedOntology.NamedIndividual[NormalizedIRI(preparedOntology, aa)].Type = aa.Literal
			}
		}
	}

	// Prepare SubClasses There are 5 different structures of SubClasses. All Class properties are IRIs:
	//
	//  * 2 Classes: The second Class is the parent of the first Class
	//  * Class and DataSomeValuesFrom: Class is the current resource and DataSomeValuesFrom contains the
	//    Datatype (e.g., xsd:string) and the corresponding DataProperty/variable name as IRI or abbreviatedIRI
	//    (e.g., filename as IRI or prop:enabled as abbreviatedIRI)
	//  * Class and DataHasValue: Class is the current resource and DataHasValue is the same as DataSomeValuesFrom
	//    except, that no Datatype exists in the Ontology, but an Literal (Literal is a string, that is not used as
	//    an Ontology object/property).
	//  * Class and ObjectSomeValuesFrom: Class is the current resource and ObjectSomeValuesFrom contains the
	//    ObjectProperty (e.g., prop:hasMultiple) and the linked resource (Class)
	//  * Class and ObjectHasValue: Class is the current resource and ObjectHasValue contains the ObjectProperty
	//    IRI (e.g., "http://graph.clouditor.io/classes/scope") and a named individual
	//    (e.g., "http://graph.clouditor.io/classes/resourceId")
	for _, sc := range src.SubClasses {
		if len(sc.Class) == 2 {
			iri := NormalizedIRI(preparedOntology, &sc.Class[0].Entity)
			parentIri := NormalizedIRI(preparedOntology, &sc.Class[1].Entity)

			// "owl#Thing" is the root of the ontology and is not needed for the protobuf files
			if parentIri != "http://www.w3.org/2002/07/owl#Thing" {
				// Create resource that has a parent. All resources directly under "owl.Thing" are already created before
				// (via the Declarations)
				r := &Resource{
					Iri:     iri,
					Name:    preparedOntology.Resources[iri].Name,
					Parent:  parentIri,
					Comment: preparedOntology.Resources[iri].Comment,
				}

				// Add subresources to the parent resource
				if val, ok := preparedOntology.Resources[parentIri]; ok {
					if val.SubResources == nil {
						preparedOntology.Resources[parentIri].SubResources = make([]*Resource, 0)
					}
					preparedOntology.Resources[parentIri].SubResources = append(preparedOntology.Resources[parentIri].SubResources, r)
				}

				// Add parent IRI to resource (not subresource!). We couldn't do this beforehand (Declarations) because we only get the information here,
				preparedOntology.Resources[iri].Parent = parentIri
			}
		} else if sc.DataSomeValuesFrom != nil {
			// Add data values, e.g. "enabled xsd:bool" ("enabled" is a data property and "xsd:bool" is a datatype) or
			for _, v := range sc.DataSomeValuesFrom {
				fromIri := NormalizedIRI(preparedOntology, &sc.Class[0].Entity)
				var (
					comment string
				)
				// Check if comment is available
				if val, ok := preparedOntology.AnnotationAssertion[v.DataProperty.IRI]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				} else if val, ok := preparedOntology.AnnotationAssertion[v.DataProperty.AbbreviatedIRI]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				// Get DataProperty name
				preparedOntology.Resources[fromIri].Relationship = append(preparedOntology.Resources[fromIri].Relationship, &Relationship{
					IRI:     NormalizedIRI(preparedOntology, &v.DataProperty.Entity),
					Typ:     util.GetProtoType(v.Datatype.AbbreviatedIRI),
					Value:   preparedOntology.GetDataPropertyIRIName(v.DataProperty),
					From:    fromIri,
					Comment: comment,
				})

			}
		} else if sc.DataHasValue != nil {
			// Add data values, e.g. "interval xsd:java.time.Duration" ("interval" is a data property and "xsd:java.time.Duration" is Literal/string)
			for _, v := range sc.DataHasValue {
				var (
					comment string
				)

				fromIri := NormalizedIRI(preparedOntology, &sc.Class[0].Entity)
				relationshipIri := NormalizedIRI(preparedOntology, &v.DataProperty.Entity)

				// Check if comment is available
				if val, ok := preparedOntology.AnnotationAssertion[relationshipIri]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				preparedOntology.Resources[fromIri].Relationship = append(preparedOntology.Resources[NormalizedIRI(preparedOntology, &sc.Class[0].Entity)].Relationship, &Relationship{
					IRI:     relationshipIri,
					Typ:     util.GetProtoType(v.Literal),
					Value:   preparedOntology.GetDataPropertyIRIName(v.DataProperty),
					From:    fromIri,
					Comment: comment,
				})

			}
		} else if sc.ObjectSomeValuesFrom != nil {
			// Add object values, e.g., "offers ResourceLogging"
			for _, v := range sc.ObjectSomeValuesFrom {
				toIri := NormalizedIRI(preparedOntology, &v.Class.Entity)
				fromIri := NormalizedIRI(preparedOntology, &sc.Class[0].Entity)
				relationshipIri := NormalizedIRI(preparedOntology, &v.ObjectProperty.Entity)

				preparedOntology.Resources[fromIri].ObjectRelationship = append(preparedOntology.Resources[fromIri].ObjectRelationship, &ObjectRelationship{
					From:               fromIri,
					ObjectProperty:     relationshipIri,
					ObjectPropertyName: preparedOntology.GetObjectPropertyIRIName(v.ObjectProperty),
					To:                 toIri,
					Name:               preparedOntology.Resources[toIri].Name,
				})
			}
		} else if sc.ObjectHasValue != nil {
			for _, v := range sc.ObjectHasValue {
				// Add object value, e.g., "scope resourceId"
				var (
					comment string
				)

				fromIri := NormalizedIRI(preparedOntology, &sc.Class[0].Entity)
				relationshipIri := NormalizedIRI(preparedOntology, &v.ObjectProperty.Entity)
				typeIri := NormalizedIRI(preparedOntology, &v.NamedIndividual.Entity)

				// Check if comment is available
				if val, ok := preparedOntology.AnnotationAssertion[relationshipIri]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				preparedOntology.Resources[fromIri].Relationship = append(preparedOntology.Resources[fromIri].Relationship, &Relationship{
					IRI:     relationshipIri,
					Typ:     util.GetProtoType(preparedOntology.NamedIndividual[typeIri].Type),
					Value:   preparedOntology.GetObjectPropertyIRIName(v.ObjectProperty),
					From:    fromIri,
					Comment: comment,
				})
			}
		}
	}

	return preparedOntology
}

// GetDataPropertyIRIName return the existing IRI (IRI vs. abbreviatedIRI) from the Data Property
func (ont *OntologyPrepared) GetDataPropertyIRIName(prop owl.DataProperty) string {
	// It is possible, that the IRI/abbreviatedIRI name is not correct, therefore we have to get the correct name from the preparedOntology. Otherwise, we get the name directly from the IRI/abbreviatedIRI
	if prop.AbbreviatedIRI != "" {
		if val, ok := ont.AnnotationAssertion[prop.AbbreviatedIRI]; ok {
			return val.Name
		} else {
			return GetDataPropertyAbbreviatedIriName(prop.AbbreviatedIRI)
		}
	} else if prop.IRI != "" {
		if val, ok := ont.Resources[prop.IRI]; ok {
			return val.Name
		} else {
			return GetNameFromIri(prop.IRI)
		}
	}

	return ""
}

// TODO(all): Use generic for GetObjectPropertyIRIName and GetDataPropertyIRIName
// GetObjectPropertyIRIName return the existing IRI (IRI vs. abbreviatedIRI) from the Data Property
func (ont *OntologyPrepared) GetObjectPropertyIRIName(prop owl.ObjectProperty) string {
	// It is possible, that the IRI/abbreviatedIRI name is not correct, therefore we have to get the correct name from the preparedOntology. Otherwise, we get the name directly from the IRI/abbreviatedIRI
	if prop.AbbreviatedIRI != "" {
		if val, ok := ont.AnnotationAssertion[prop.AbbreviatedIRI]; ok {
			return val.Name
		} else {
			return GetDataPropertyAbbreviatedIriName(prop.AbbreviatedIRI)
		}
	} else if prop.IRI != "" {
		if val, ok := ont.Resources[prop.IRI]; ok {
			return val.Name
		} else {
			return GetNameFromIri(prop.IRI)
		}
	}

	return ""
}
