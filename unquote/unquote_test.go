package unquote

import (
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/protocolbuffers/txtpbfmt/ast"
)

func TestUnquote(t *testing.T) {
	inputs := []struct {
		in       string
		want     string
		wantRaw  string
		wantRune rune
	}{
		{
			in:       `"value"`,
			want:     `value`,
			wantRaw:  `value`,
			wantRune: rune('"'),
		},
		{
			in:       `'value'`,
			want:     `value`,
			wantRaw:  `value`,
			wantRune: rune('\''),
		},
		{
			in:       `"foo\'\a\b\f\n\r\t\vbar"`,
			want:     "foo'\a\b\f\n\r\t\vbar", // Double-quoted; string contains real control characters.
			wantRaw:  `foo\'\a\b\f\n\r\t\vbar`,
			wantRune: rune('"'),
		},
		{
			in:       `'foo\"bar'`,
			want:     `foo"bar`,
			wantRaw:  `foo\"bar`,
			wantRune: rune('\''),
		},
	}
	for _, input := range inputs {
		node := &ast.Node{Name: "name", Values: []*ast.Value{{Value: input.in}}}

		got, gotRune, err := Unquote(node)
		if err != nil {
			t.Errorf("Unquote(%v) returned err %v", input.in, err)
			continue
		}
		if diff := diff.Diff(input.want, got); diff != "" {
			t.Logf("want: %q", input.want)
			t.Logf("got:  %q", got)
			t.Errorf("Unquote(%v) returned diff (-want, +got):\n%s", input.in, diff)
		}
		if gotRune != input.wantRune {
			t.Errorf("Unquote(%v) returned rune %q, want %q", input.in, gotRune, input.wantRune)
		}

		got, gotRune, err = Raw(node)
		if err != nil {
			t.Errorf("unquote.Raw(%v) returned err %v", input.in, err)
			continue
		}
		if diff := diff.Diff(input.wantRaw, got); diff != "" {
			t.Errorf("unquote.Raw(%v) returned diff (-wantRaw, +got):\n%s", input.in, diff)
		}
		if gotRune != input.wantRune {
			t.Errorf("unquote.Raw(%v) returned rune %q, want %q", input.in, gotRune, input.wantRune)
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
		{
			in:      `"foo\x"`,
			wantErr: `\x requires 2 following digits`,
		},
		{
			in:      `"foo\xH"`,
			wantErr: `\x requires 2 following digits`,
		},
		{
			in:      `"foo\xHH"`,
			wantErr: `\xHH contains non-hexadecimal digits`,
		},
		{
			in:      `"foo\u"`,
			wantErr: `\u requires 4 following digits`,
		},
		{
			in:      `"foo\uH"`,
			wantErr: `\u requires 4 following digits`,
		},
		{
			in:      `"foo\uHHHH"`,
			wantErr: `\uHHHH contains non-hexadecimal digits`,
		},
		{
			in:      `"foo\U"`,
			wantErr: `\U requires 8 following digits`,
		},
		{
			in:      `"foo\UH"`,
			wantErr: `\U requires 8 following digits`,
		},
		{
			in:      `"foo\UHHHHHHHH"`,
			wantErr: `\UHHHHHHHH contains non-hexadecimal digits`,
		},
		{
			in:      `"foo\UFFFFFFFF"`,
			wantErr: `\UFFFFFFFF is not a valid Unicode code point`,
		},
		{
			in:      `"foo\0"`,
			wantErr: `\0 requires 2 following digits`,
		},
		{
			in:      `"foo\0H"`,
			wantErr: `\0 requires 2 following digits`,
		},
		{
			in:      `"foo\0HH"`,
			wantErr: `\0HH contains non-octal digits`,
		},
		{
			in:      `"foo\y"`,
			wantErr: `unknown escape \y`,
		},
	}

	for _, input := range inputs {
		node := &ast.Node{Name: "name", Values: []*ast.Value{{Value: input.in}}}

		_, _, err := Unquote(node)
		if err == nil || !strings.Contains(err.Error(), input.wantErr) {
			t.Errorf("Unquote(%s) got %v, want err to contain %q", input.in, err, input.wantErr)
		}
	}
}

func TestErrorHandlingRaw(t *testing.T) {
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

		_, _, err := Raw(node)
		if err == nil || !strings.Contains(err.Error(), input.wantErr) {
			t.Errorf("Raw(%s) got %v, want err to contain %q", input.in, err, input.wantErr)
		}
	}
}

func TestErrorHandlingMisc(t *testing.T) {
	inputs := []struct {
		in      string
		wantErr string
	}{
		{
			in:      `"value`,
			wantErr: "unmatched quote",
		},
	}

	for _, input := range inputs {
		node := &ast.Node{Name: "name", Values: []*ast.Value{{Value: input.in}}}

		_, _, err := Unquote(node)
		if err == nil || !strings.Contains(err.Error(), input.wantErr) {
			t.Errorf("Unquote(%s) got %v, want err to contain %q", input.in, err, input.wantErr)
		}
	}
}
