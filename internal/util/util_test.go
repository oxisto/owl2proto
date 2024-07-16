package util

import (
	"reflect"
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy path: upper case",
			args: args{
				s: "Logging",
			},
			want: "logging",
		},
		{
			name: "Happy path: empty",
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "Happy path: OSLogging",
			args: args{
				s: "OSLogging",
			},
			want: "os_logging",
		},
		{
			name: "Happy path: CICDService",
			args: args{
				s: "CICDService",
			},
			want: "cicd_service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeCase(tt.args.s); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFieldNumber(t *testing.T) {
	type args struct {
		deterministicFieldNumbers bool
		counter                   int
		input                     []string
	}
	tests := []struct {
		name        string
		args        args
		fieldNumber int
		wantI       int
	}{
		{
			name: "Happy path: incremented",
			args: args{
				deterministicFieldNumbers: false,
				counter:                   0,
				input:                     []string{"Resource", "Compute", "VirtualMachine", "name"},
			},
			fieldNumber: 1,
			wantI:       1,
		},
		{
			name: "Happy path: consistent",
			args: args{
				deterministicFieldNumbers: true,
				counter:                   0,
				input:                     []string{"Resource", "Compute", "VirtualMachine", "name"},
			},
			fieldNumber: 4044,
			wantI:       0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetFieldNumber(tt.args.deterministicFieldNumbers, tt.args.counter, tt.args.input...)
			if got != tt.fieldNumber {
				t.Errorf("GetFieldNumber() got = %v, want %v", got, tt.fieldNumber)
			}
			if got1 != tt.wantI {
				t.Errorf("GetFieldNumber() got1 = %v, want %v", got1, tt.wantI)
			}
		})
	}
}

func TestToPlural(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy path: last character 'y'",
			args: args{
				s: "Availability",
			},
			want: "Availabilities",
		},
		{
			name: "Happy path: last character not 'y'",
			args: args{
				s: "Application",
			},
			want: "Applications",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToPlural(tt.args.s); got != tt.want {
				t.Errorf("ToPlural() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy path: space",
			args: args{
				s: "test word",
			},
			want: "testword",
		},
		{
			name: "Happy path: '-'",
			args: args{
				s: "test-word",
			},
			want: "testword",
		},
		{
			name: "Happy path: '/'",
			args: args{
				s: "test/word",
			},
			want: "testword",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanString(tt.args.s); got != tt.want {
				t.Errorf("CleanString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortMapKeys(t *testing.T) {
	type args struct {
		m map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Happy path",
			args: args{
				m: map[string]string{
					"citrus": "thirdValue",
					"banana": "secondValue",
					"apple":  "firstValue",
				},
			},
			want: []string{"apple", "banana", "citrus"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SortMapKeys(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortMapKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProtoType(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy path: unknown type",
			args: args{
				s: "unknown_type",
			},
			want: "unknown_type",
		},
		{
			name: "Happy path: bool",
			args: args{
				s: "xsd:boolean",
			},
			want: "bool",
		},
		{
			name: "Happy path: string",
			args: args{
				s: "xsd:String",
			},
			want: "string",
		},
		{
			name: "Happy path: listString",
			args: args{
				s: "xsd:listString",
			},
			want: "repeated string",
		},
		{
			name: "Happy path: int",
			args: args{
				s: "xsd:int",
			},
			want: "int32",
		},
		{
			name: "Happy path: Short",
			args: args{
				s: "xsd:Short",
			},
			want: "uint32",
		},
		{
			name: "Happy path: float",
			args: args{
				s: "xsd:float",
			},
			want: "float",
		},
		{
			name: "Happy path: xsd:java.time.Duration",
			args: args{
				s: "xsd:java.time.Duration",
			},
			want: "google.protobuf.Duration",
		},
		{
			name: "Happy path: xsd:dateTime",
			args: args{
				s: "xsd:dateTime",
			},
			want: "google.protobuf.Timestamp",
		},
		{
			name: "Happy path: xsd:java.util.ArrayList<Short>",
			args: args{
				s: "xsd:java.util.ArrayList<Short>",
			},
			want: "repeated uint32",
		},
		{
			name: "Happy path: xsd:java.util.Map<String, String>",
			args: args{
				s: "xsd:java.util.Map<String, String>",
			},
			want: "map<string, string>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetProtoType(tt.args.s); got != tt.want {
				t.Errorf("GetProtoType() = %v, want %v", got, tt.want)
			}
		})
	}
}
