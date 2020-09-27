package jet

import (
	_ "net/http/pprof"
	"reflect"
	"testing"
)

func Test_parseArgs(t *testing.T) {
	type args struct {
		str []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{args: args{[]string{"asd", "asd", "asd", "--scraper", "rutgers", "-u", "NK", "-c"}}, want: []string{"-u", "NK", "-c"}},
	}
	for _, tt := range tests {
		if got := parseArgs(tt.args.str); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. parseArgs() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
