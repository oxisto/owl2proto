package ontology

import (
	"testing"
)

func TestGetDataPropertyAbbreviatedIriName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty input",
			args: args{},
			want: "",
		},
		{
			name: "Happy path",
			args: args{
				s: "prop:has",
			},
			want: "has",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNameWithoutPrefix(tt.args.s); got != tt.want {
				t.Errorf("GetDataPropertyAbbreviatedIriName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNameFromIri(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty input",
			args: args{},
			want: "",
		},
		{
			name: "Happy path",
			args: args{
				s: "https://example.com/Resource",
			},
			want: "Resource",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNameFromIri(tt.args.s); got != tt.want {
				t.Errorf("GetNameFromIri() = %v, want %v", got, tt.want)
			}
		})
	}
}
