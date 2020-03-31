package fmt_test

import (
	"bytes"
	"os/exec"
	"testing"

	"google3/base/go/runfiles"
	"github.com/kylelemons/godebug/pretty"
)

func TestFmtCommand(t *testing.T) {
	path := runfiles.Path("github.com/protocolbuffers/txtpbfmt/fmt")

	tests := []struct {
		flags []string
		input string
		want  string
	}{
		{
			input: `name: """value"""`,
			want: `name:
  ""
  "value"
  ""
`,
		},
		{
			flags: []string{"--allow_triple_quoted_strings"},
			input: `name:"""value"""`,
			want: `name: """value"""
`,
		},
	}

	for _, test := range tests {
		out := &bytes.Buffer{}
		cmd := &exec.Cmd{
			Path:   path,
			Args:   append([]string{path}, test.flags...),
			Stdin:  bytes.NewBufferString(test.input),
			Stdout: out,
		}
		if err := cmd.Run(); err != nil {
			t.Errorf("fmt %v with input %q unexpectedly returned error: %v", test.flags, test.input, err)
			continue
		}
		have := out.String()
		if diff := pretty.Compare(test.want, have); diff != "" {
			t.Errorf("fmt %v with input %q output diff (-want, +have):\n%s", test.flags, test.input, diff)
		}
	}

}
