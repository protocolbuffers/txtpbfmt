package quote

import (
	"testing"
)

func TestUnescapeQuotes(t *testing.T) {
	inputs := []struct {
		in   string
		want string
	}{
		{in: ``, want: ``},
		{in: `"`, want: `"`},
		{in: `\`, want: `\`},
		{in: `\\`, want: `\\`},
		{in: `\\\`, want: `\\\`},
		{in: `"\"`, want: `""`},
		{in: `"\\"`, want: `"\\"`},
		{in: `"\\\"`, want: `"\\"`},
		{in: `'\'`, want: `''`},
		{in: `'\\'`, want: `'\\'`},
		{in: `'\\\'`, want: `'\\'`},
		{in: `'\n'`, want: `'\n'`},
		{in: `\'\"\\\n\"\'`, want: `'"\\\n"'`},
	}
	for _, input := range inputs {
		got := unescapeQuotes(input.in)
		if got != input.want {
			t.Errorf("unescapeQuotes(`%s`): got `%s`, want `%s`", input.in, got, input.want)
		}
	}
}
