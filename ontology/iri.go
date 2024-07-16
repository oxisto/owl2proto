package ontology

import (
	"strings"

	"github.com/oxisto/owl2proto/owl"
)

// NormalizedIRI returns the normalized (abbreviated) IRI
func (ont *OntologyPrepared) NormalizedIRI(c *owl.Class) string {
	if c.IRI != "" {
		return c.IRI
	} else if c.AbbreviatedIRI != "" {
		return ont.normalizeAbbreviatedIRI(c.AbbreviatedIRI)
	}

	return ""
}

// normalizeAbbreviatedIRI normalizes the abbreviated IRI, e.g., "ex:Storage" -> "http://example.com/cloud/Storage"
func (ont *OntologyPrepared) normalizeAbbreviatedIRI(iri string) string {
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
