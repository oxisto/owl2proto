package ontology

import (
	"testing"

	"github.com/oxisto/owl2proto/owl"
)

func TestOntologyPrepared_normalizeIRI(t *testing.T) {
	type fields struct {
		Resources           map[string]*Resource
		SubClasses          map[string]*owl.SubClassOf
		AnnotationAssertion map[string]*AnnotationAssertion
		Prefixes            map[string]*owl.Prefix
		RootResourceName    string
	}
	type args struct {
		iri string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "happy path",
			fields: fields{
				Prefixes: map[string]*owl.Prefix{
					"ex": {
						Name: "ex",
						IRI:  "http://example.com/",
					},
				},
			},
			args: args{
				iri: "ex:Resource",
			},
			want: "http://example.com/Resource",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ont := &OntologyPrepared{
				Resources:           tt.fields.Resources,
				SubClasses:          tt.fields.SubClasses,
				AnnotationAssertion: tt.fields.AnnotationAssertion,
				Prefixes:            tt.fields.Prefixes,
				RootResourceName:    tt.fields.RootResourceName,
			}
			if got := ont.normalizeIRI(tt.args.iri); got != tt.want {
				t.Errorf("OntologyPrepared.normalizeIRI() = %v, want %v", got, tt.want)
			}
		})
	}
}
