package ontology

import (
	"strings"

	"github.com/oxisto/owl2proto/owl"
)

func (ont *OntologyPrepared) NormalizedIRI(c *owl.Class) string {
	if c.IRI != "" {
		return c.IRI
	} else if c.AbbreviatedIRI != "" {
		return ont.normalizeIRI(c.AbbreviatedIRI)
	}

	return ""
}

func (ont *OntologyPrepared) normalizeIRI(iri string) string {
	// We need to split the abbreviated IRI and look for the matching prefix
	prefix, name, found := strings.Cut(iri, ":")
	if !found {
		return ""
	}

	p, ok := ont.Prefixes[prefix]
	if ok {
		return p.IRI + name
	}

	return ""
}
