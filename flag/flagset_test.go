package flag

import (
	"reflect"
	"testing"
)

func TestParseSet(t *testing.T) {
	type Cfg struct {
		Addr string
	}

	tests := []struct {
		Args    []string
		Environ []string
		Got     Cfg
		Want    Cfg
	}{
		{
			Args: []string{"-listen.addr=blah"},
			Want: Cfg{
				Addr: "blah",
			},
		},
		{
			Environ: []string{"APPD_LISTEN_ADDR=blah"},
			Want: Cfg{
				Addr: "blah",
			},
		},
		{
			Args:    []string{"-listen.addr=derp"},
			Environ: []string{"APPD_LISTEN_ADDR=blah"},
			Want: Cfg{
				Addr: "derp",
			},
		},
	}

	for _, tt := range tests {
		err := parseSet("appd", tt.Args, tt.Environ, func(fs FlagSet) {
			fs.String(&tt.Got.Addr, "address on which to listen", "listen", "addr")
		})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(tt.Want, tt.Got) {
			t.Errorf("want=%#v", tt.Want)
			t.Fatalf(" got=%#v", tt.Got)
		}
	}
}
