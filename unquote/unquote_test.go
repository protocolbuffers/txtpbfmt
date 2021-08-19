package unquote

import (
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/protocolbuffers/txtpbfmt/ast"
)

func TestUnquote(t *testing.T) {
	inputs := []struct {
		in      string
		want    string
		wantRaw string
	}{
		{
			in:      `"value"`,
			want:    `value`,
			wantRaw: `value`,
		},
		{
			in:      `'value'`,
			want:    `value`,
			wantRaw: `value`,
		},
		{
			in:      `"foo\'bar"`,
			want:    `foo'bar`,
			wantRaw: `foo\'bar`,
		},
		{
			in:      `'foo\"bar'`,
			want:    `foo"bar`,
			wantRaw: `foo\"bar`,
		},
	}
	for _, input := range inputs {
		node := &ast.Node{Name: "name", Values: []*ast.Value{{Value: input.in}}}

		got, err := Unquote(node)
		if err != nil {
			t.Errorf("Unquote(%v) returned err %v", input.in, err)
			continue
		}
		if diff := diff.Diff(input.want, got); diff != "" {
			t.Errorf("Unquote(%v) returned diff (-want, +got):\n%s", input.in, diff)
		}

		got, err = Raw(node)
		if err != nil {
			t.Errorf("unquote.Raw(%v) returned err %v", input.in, err)
			continue
		}
		if diff := diff.Diff(input.wantRaw, got); diff != "" {
			t.Errorf("unquote.Raw(%v) returned diff (-wantRaw, +got):\n%s", input.in, diff)
		}
	}
}
