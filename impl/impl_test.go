package impl

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestPreprocess(t *testing.T) {
	type testType int
	const (
		nonTripleQuotedTest testType = 1
		tripleQuotedTest    testType = 2
	)
	inputs := []struct {
		name     string
		in       string
		want     map[int]bool
		err      bool
		testType testType
	}{{
		name: "simple example",
		//   012
		in: `p {}`,
		want: map[int]bool{
			2: true,
		},
	}, {
		name: "multiple nested children in the same line",
		//   0123456
		in: `p { b { n: v } }`,
		want: map[int]bool{
			2: true,
			6: true,
		},
	}, {
		name: "only second line",
		//   0123
		in: `p {
b { n: v } }`,
		want: map[int]bool{
			6: true,
		},
	}, {
		name: "empty output",
		in: `p {
b {
	n: v } }`,
		want: map[int]bool{},
	}, {
		name: "comments and strings",
		in: `
# p      {}
s: "p   {}"
# s: "p {}"
s: "#p  {}"
p        {}`,
		want: map[int]bool{
			// (5 lines) * (10 chars) - 2
			58: true,
		},
	}, {
		name: "escaped char",
		in: `p { s="\"}"
	}`,
		want: map[int]bool{},
	}, {
		name: "missing '}'",
		in:   `p {`,
		want: map[int]bool{},
	}, {
		name: "too many '}'",
		in:   `p {}}`,
		err:  true,
	}, {
		name: "single quote",
		in:   `"`,
		err:  true,
	}, {
		name: "double quote",
		in:   `""`,
	}, {
		name: "two single quotes",
		in:   `''`,
	}, {
		name: "single single quote",
		in:   `'`,
		err:  true,
	}, {
		name: "naked single quote in double quotes",
		in:   `"'"`,
	}, {
		name: "escaped single quote in double quotes",
		in:   `"\'"`,
	}, {
		name: "invalid naked single quote in single quotes",
		in:   `'''`,
		err:  true,
	}, {
		name: "invalid standalone angled bracket",
		in:   `>`,
		err:  true,
	}, {
		name: "invalid angled bracket outside template",
		in:   `foo > bar`,
		err:  true,
	}, {
		name: "valid angled bracket inside string",
		in:   `"foo > bar"`,
	}, {
		name: "valid angled bracket inside template",
		in:   `% foo >= bar %`,
	}, {
		name: "valid angled bracket inside comment",
		in:   `# foo >= bar`,
	}, {
		name: "valid angled bracket inside if condition in template",
		in:   `%if (value > 0)%`,
	}, {
		name: "valid templated arg inside comment",
		in:   `# foo: %bar%`,
	}, {
		name: "valid templated arg inside string",
		in:   `foo: "%bar%"`,
	}, {
		name: "% delimiter inside commented lines",
		in: `
					# comment %
					{
						# comment %
					}
					`,
	}, {
		name: "% delimiter inside strings",
		in: `
					foo: "%"
					{
						bar: "%"
					}
					`,
	}, {
		name: "escaped single quote in single quotes",
		in:   `'\''`,
	}, {
		name: "two single quotes",
		in:   `''`,
	}, {
		name:     "triple quoted backlash",
		in:       `"""\"""`,
		err:      false,
		testType: tripleQuotedTest,
	}, {
		name:     "triple quoted backlash invalid",
		in:       `"""\"""`,
		err:      true,
		testType: nonTripleQuotedTest,
	}, {
		name:     "triple quoted and regular quotes backslash handling",
		in:       `"""text""" "\""`,
		err:      false,
		testType: tripleQuotedTest,
	}, {
		name: "txtpbfmt: off/on",
		in: `# txtpbfmt: off
      foo: "bar"
   {
     bar: "baz"
  }
# txtpbfmt: on
    `, err: false,
	}}
	for _, input := range inputs {
		bytes := []byte(input.in)
		// ensure capacity is equal to length to catch slice index out of bounds errors
		bytes = bytes[0:len(bytes):len(bytes)]
		if input.testType != tripleQuotedTest {
			have, err := sameLineBrackets(bytes, false)
			if (err != nil) != input.err {
				t.Errorf("sameLineBrackets[%s] allowTripleQuotedStrings=false %v returned err %v", input.name, input.in, err)
				continue
			}
			if diff := pretty.Compare(input.want, have); diff != "" {
				t.Errorf("sameLineBrackets[%s] allowTripleQuotedStrings=false %v returned diff (-want, +have):\n%s", input.name, input.in, diff)
			}
		}

		if input.testType != nonTripleQuotedTest {
			have, err := sameLineBrackets(bytes, true)
			if (err != nil) != input.err {
				t.Errorf("sameLineBrackets[%s] allowTripleQuotedStrings=true %v returned err %v", input.name, input.in, err)
				continue
			}
			if diff := pretty.Compare(input.want, have); diff != "" {
				t.Errorf("sameLineBrackets[%s] allowTripleQuotedStrings=true %v returned diff (-want, +have):\n%s", input.name, input.in, diff)
			}
		}
	}
}

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
