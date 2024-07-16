package ontology

import (
	"strings"

	"github.com/oxisto/owl2proto/internal/util"
	"github.com/oxisto/owl2proto/owl"
)

// OntologyPrepared contains an [owl.Ontology] in a way that is "prepared" for the translation to protobuf messages.
type OntologyPrepared struct {
	Resources           map[string]*Resource
	SubClasses          map[string]*owl.SubClassOf
	AnnotationAssertion map[string]*AnnotationAssertion

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

	// "owl.Thing" is the root of the ontology and is not needed for the protobuf files. We do not have it in the prepared Ontology, so we have to skip, if we come to "owl.Thing".
	res, ok := po.Resources[key]
	if !ok {
		return nil
	}

	relationships = append(relationships, res.Relationship...)

	parent = po.Resources[key].Parent
	if parent == "" || key == po.RootResourceName {
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
		RootResourceName:    rootIRI,
	}

	for idx := range src.Prefixes {
		p := &src.Prefixes[idx]
		preparedOntology.Prefixes[p.Name] = p
	}

	for _, c := range src.Declarations {
		iri := preparedOntology.NormalizedIRI(&c.Class)

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
	}

	// Prepare name and comment
	for _, aa := range src.AnnotationAssertion {
		if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:label" {
			if _, ok := preparedOntology.Resources[aa.IRI]; ok {
				preparedOntology.Resources[aa.IRI].Name = util.CleanString(aa.Literal)
			} else if _, ok := preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI]; ok {
				preparedOntology.AnnotationAssertion[aa.AbbreviatedIRI].Name = util.CleanString(aa.Literal)
			}
		} else if aa.AnnotationProperty.AbbreviatedIRI == "rdfs:comment" {
			if _, ok := preparedOntology.Resources[aa.IRI]; ok {
				c := preparedOntology.Resources[aa.IRI].Comment
				c = append(c, aa.Literal)
				preparedOntology.Resources[aa.IRI].Comment = c
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
			iri := preparedOntology.NormalizedIRI(&sc.Class[0])
			parentIri := preparedOntology.NormalizedIRI(&sc.Class[1])

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
				fromIri := preparedOntology.NormalizedIRI(&sc.Class[0])
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
					IRI:     v.DataProperty.IRI,
					Typ:     util.GetProtoType(v.Datatype.AbbreviatedIRI),
					Value:   preparedOntology.GetDataPropertyIRIName(v.DataProperty),
					Comment: comment,
				})

			}
		} else if sc.DataHasValue != nil {
			// Add data values, e.g. "interval xsd:java.time.Duration" ("interval" is a data property and "xsd:java.time.Duration" is Literal/string)
			for _, v := range sc.DataHasValue {
				var (
					comment    string
					identifier string
				)

				// Get IRI or abbreviatedIRI from DataProperty
				if v.DataProperty.AbbreviatedIRI != "" {
					identifier = v.DataProperty.AbbreviatedIRI
				} else if v.DataProperty.IRI != "" {
					identifier = v.DataProperty.IRI
				}

				// Check if comment is available
				if val, ok := preparedOntology.AnnotationAssertion[identifier]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(preparedOntology.Resources[sc.Class[0].IRI].Relationship, &Relationship{
					IRI:     identifier,
					Typ:     util.GetProtoType(v.Literal),
					Value:   preparedOntology.GetDataPropertyIRIName(v.DataProperty),
					Comment: comment,
				})

			}
		} else if sc.ObjectSomeValuesFrom != nil {
			// Add object values, e.g., "offers ResourceLogging"
			for _, v := range sc.ObjectSomeValuesFrom {
				classIri := preparedOntology.NormalizedIRI(&v.Class)
				fromIri := preparedOntology.NormalizedIRI(&sc.Class[0])

				if v.ObjectProperty.IRI != "" {
					preparedOntology.Resources[fromIri].ObjectRelationship = append(preparedOntology.Resources[fromIri].ObjectRelationship, &ObjectRelationship{
						ObjectProperty:     v.ObjectProperty.IRI,
						ObjectPropertyName: preparedOntology.GetObjectPropertyIRIName(v.ObjectProperty),
						Class:              classIri,
						Name:               preparedOntology.Resources[classIri].Name,
					})
				} else if v.ObjectProperty.AbbreviatedIRI != "" {
					preparedOntology.Resources[fromIri].ObjectRelationship = append(preparedOntology.Resources[fromIri].ObjectRelationship, &ObjectRelationship{
						ObjectProperty:     v.ObjectProperty.AbbreviatedIRI,
						ObjectPropertyName: preparedOntology.GetObjectPropertyIRIName(v.ObjectProperty),
						Class:              classIri,
						Name:               preparedOntology.Resources[classIri].Name,
					})
				}
			}
		} else if sc.ObjectHasValue != nil {
			for _, v := range sc.ObjectHasValue {
				// Add object value, e.g., "scope resourceId"
				var (
					comment         string
					identifier      string
					namedIndividual string
				)

				// Get IRI or abbreviatedIRI from ObjectProperty
				if v.ObjectProperty.AbbreviatedIRI != "" {
					identifier = v.ObjectProperty.AbbreviatedIRI
				} else if v.ObjectProperty.IRI != "" {
					identifier = v.ObjectProperty.IRI
				}

				// Get IRI or abbreviatedIRI from NamedIndividual
				if v.NamedIndividual.AbbreviatedIRI != "" {
					namedIndividual = v.NamedIndividual.AbbreviatedIRI
				} else if v.NamedIndividual.IRI != "" {
					namedIndividual = v.NamedIndividual.IRI
				}

				// Check if comment is available
				if val, ok := preparedOntology.AnnotationAssertion[identifier]; ok {
					comment = strings.Join(val.Comment[:], "\n\t ")
				}

				preparedOntology.Resources[sc.Class[0].IRI].Relationship = append(preparedOntology.Resources[sc.Class[0].IRI].Relationship, &Relationship{
					IRI:     identifier,
					Typ:     util.GetProtoType(namedIndividual),
					Value:   preparedOntology.GetObjectPropertyIRIName(v.ObjectProperty),
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
