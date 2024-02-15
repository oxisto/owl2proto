package util

import "testing"

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
