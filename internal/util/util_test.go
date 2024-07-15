package util

import (
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
		fieldNumberOption bool
		counter           int
		input             []string
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
				fieldNumberOption: false,
				counter:           0,
				input:             []string{"Resource", "Compute", "VirtualMachine", "name"},
			},
			fieldNumber: 1,
			wantI:       1,
		},
		{
			name: "Happy path: consistent",
			args: args{
				fieldNumberOption: true,
				counter:           0,
				input:             []string{"Resource", "Compute", "VirtualMachine", "name"},
			},
			fieldNumber: 4044,
			wantI:       0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetFieldNumber(tt.args.fieldNumberOption, tt.args.counter, tt.args.input...)
			if got != tt.fieldNumber {
				t.Errorf("GetFieldNumber() got = %v, want %v", got, tt.fieldNumber)
			}
			if got1 != tt.wantI {
				t.Errorf("GetFieldNumber() got1 = %v, want %v", got1, tt.wantI)
			}
		})
	}
}
