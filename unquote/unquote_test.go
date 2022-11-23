package unquote

import (
	"strings"
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
			in:      `"foo\'\a\b\f\n\r\t\vbar"`,
			want:    "foo'\a\b\f\n\r\t\vbar", // Double-quoted; string contains real control characters.
			wantRaw: `foo\'\a\b\f\n\r\t\vbar`,
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
			t.Logf("want: %q", input.want)
			t.Logf("got:  %q", got)
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

func TestErrorHandling(t *testing.T) {
	inputs := []struct {
		in      string
		wantErr string
	}{
		{
			in:      `"value`,
			wantErr: "unmatched quote",
		},
		{
			in:      `"`,
			wantErr: "not a quoted string",
		},
		{
			in:      "`foo`",
			wantErr: "invalid quote character `",
		},
	}

	for _, input := range inputs {
		node := &ast.Node{Name: "name", Values: []*ast.Value{{Value: input.in}}}

		_, err := Unquote(node)
		if err == nil || !strings.Contains(err.Error(), input.wantErr) {
			t.Errorf("Unquote(%s) got %v, want err to contain %q", input.in, err, input.wantErr)
		}

		_, err = Raw(node)
		if err == nil || !strings.Contains(err.Error(), input.wantErr) {
			t.Errorf("Raw(%s) got %v, want err to contain %q", input.in, err, input.wantErr)
		}
	}
}
