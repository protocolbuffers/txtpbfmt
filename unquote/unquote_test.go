package unquote

import (
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/protocolbuffers/txtpbfmt/parser"
)

func TestUnquote(t *testing.T) {
	inputs := []struct {
		in   string
		want string
	}{
		{
			in:   `name: "value"`,
			want: "value",
		},
		{
			in:   `name: "foo\'bar"`,
			want: "foo'bar",
		},
	}
	for _, input := range inputs {
		nodes, err := parser.Parse([]byte(input.in))
		if err != nil {
			t.Errorf("Parse %v returned err %v", input.in, err)
			continue
		}
		if len(nodes) == 0 {
			t.Errorf("Parse %v returned no nodes", input.in)
			continue
		}
		got, err := Unquote(nodes[0])
		if err != nil {
			t.Errorf("Unquote %v returned err %v", input.in, err)
			continue
		}
		if diff := diff.Diff(input.want, got); diff != "" {
			t.Errorf("Unquote %v returned diff (-want, +got):\n%s", input.in, diff)
		}
	}
}
