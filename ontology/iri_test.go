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
			name: "abbreviatedIRI not valid",
			fields: fields{
				Prefixes: map[string]*owl.Prefix{
					"ex": {
						Name: "ex",
						IRI:  "https://example.com/",
					},
				},
			},
			args: args{
				iri: "Resource",
			},
			want: "",
		},
		{
			name: "prefix not available",
			fields: fields{
				Prefixes: map[string]*owl.Prefix{
					"ex": {
						Name: "ex",
						IRI:  "https://example.com/",
					},
				},
			},
			args: args{
				iri: "x:Resource",
			},
			want: "",
		},
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
			if got := ont.normalizeAbbreviatedIRI(tt.args.iri); got != tt.want {
				t.Errorf("OntologyPrepared.normalizeIRI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOntologyPrepared_NormalizedIRI(t *testing.T) {
	type fields struct {
		Resources           map[string]*Resource
		SubClasses          map[string]*owl.SubClassOf
		AnnotationAssertion map[string]*AnnotationAssertion
		Prefixes            map[string]*owl.Prefix
		RootResourceName    string
	}
	type args struct {
		c *owl.Entity
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "No IRI available",
			args: args{
				c: &owl.Entity{},
			},
			want: "",
		},
		{
			name: "Happy path: abbreviated IRI available",
			fields: fields{
				Prefixes: map[string]*owl.Prefix{
					"ex": {
						Name: "ex",
						IRI:  "http://example.com/cloud/",
					},
				},
			},
			args: args{
				&owl.Entity{
					AbbreviatedIRI: "ex:Storage",
				},
			},
			want: "http://example.com/cloud/Storage",
		},
		{
			name: "Happy path: IRI available",
			args: args{
				&owl.Entity{
					IRI: "http://example.com/cloud/",
				},
			},
			want: "http://example.com/cloud/",
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
			if got := ont.NormalizedIRI(tt.args.c); got != tt.want {
				t.Errorf("OntologyPrepared.NormalizedIRI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOntologyPrepared_AbbreviateIRI(t *testing.T) {
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
				iri: "http://example.com/Resource",
			},
			want: "ex:Resource",
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
			if got := ont.AbbreviateIRI(tt.args.iri); got != tt.want {
				t.Errorf("OntologyPrepared.AbbreviateIRI() = %v, want %v", got, tt.want)
			}
		})
	}
}
