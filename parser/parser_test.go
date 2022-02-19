// Open source parser tests.
package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/kylelemons/godebug/pretty"
)

func TestError(t *testing.T) {
	inputs := []struct {
		in  string
		err string
	}{{
		in: "list_wrong_end: [one, two}",
	}, {
		in:  "multy_string_list: ['a' 'b', 'c']",
		err: "multiple-string value",
	}, {
		in: `mapping {{
  my_proto_field: "foo"
}}`, err: "Failed to find a FieldName",
	}, {
		in: `name: "string with literal new line
"`, err: "new line",
	}}
	for _, input := range inputs {
		out, err := Format([]byte(input.in))
		if err == nil {
			nodes, err := Parse([]byte(input.in))
			if err != nil {
				t.Errorf("Parse %v returned err %v", input.in, err)
				continue
			}
			t.Errorf("Expected a formatting error but none was raised while formatting:\n%v\nout\n%s\ntree\n%s\n", input.in, out, DebugFormat(nodes, 0))
			continue
		}
		if !strings.Contains(err.Error(), input.err) {
			t.Errorf(
				"Expected a formatting error containing \"%v\",\n  error was: \"%v\"\nwhile formatting:\n%v", input.err, err, input.in)
		}
	}
}

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

func TestDisable(t *testing.T) {
	inputs := []string{
		`#txtpbfmt:disable
some{random:field}`,
		`#a
#b

# txtpbfmt: disable
some{random:field}`,
	}
	for _, input := range inputs {
		output, err := Format([]byte(input))
		if err != nil {
			t.Errorf("%v: %v", err, input)
			continue
		}
		if diff := pretty.Compare(input, string(output)); diff != "" {
			t.Errorf("disable ineffective (-want, +have):\n%s", diff)
		}
	}
}

func TestFormat(t *testing.T) {
	inputs := []struct {
		name string
		in   string
		out  string
	}{{
		name: "file comment + block comment",
		in: `# file comment 1
    # file comment 2

    # presubmit comment 1
    # presubmit comment 2
    presubmit: {
      # review comment 1
      # review comment 2
      review_notify: "address"  # review same line comment 1

			# extra comment block 1 line 1
			# extra comment block 1 line 2

			# extra comment block 2 line 1
			# extra comment block 2 line 2
    }
`,
		out: `# file comment 1
# file comment 2

# presubmit comment 1
# presubmit comment 2
presubmit: {
  # review comment 1
  # review comment 2
  review_notify: "address"  # review same line comment 1

  # extra comment block 1 line 1
  # extra comment block 1 line 2

  # extra comment block 2 line 1
  # extra comment block 2 line 2
}
`}, {
		name: "2x file comment",
		in: `# file comment block 1 line 1
# file comment block 1 line 2

# file comment block 2 line 1
# file comment block 2 line 2

presubmit: {}
`,
		out: `# file comment block 1 line 1
# file comment block 1 line 2

# file comment block 2 line 1
# file comment block 2 line 2

presubmit: {}
`}, {
		name: "much blank space",
		in: `


      # presubmit comment 1
       # presubmit comment 2
   presubmit  :   {


      # review comment 1
    # review comment 2
     review_notify :   "address"     # review same line comment 2


}


`,
		out: `# presubmit comment 1
# presubmit comment 2
presubmit: {

  # review comment 1
  # review comment 2
  review_notify: "address"  # review same line comment 2

}

`}, {
		name: "empty children",
		in: `


  # file comment 1
    # file comment 2


   # presubmit comment 1
    # presubmit comment 2
presubmit: {
}


`,
		out: `# file comment 1
# file comment 2

# presubmit comment 1
# presubmit comment 2
presubmit: {
}

`}, {
		name: "list notation with []",
		in: `
presubmit: {
  check_tests: {
    action: [ MAIL, REVIEW, SUBMIT ]
  }
}
`,
		out: `presubmit: {
  check_tests: {
    action: [MAIL, REVIEW, SUBMIT]
  }
}
`}, {
		name: "list notation with [{}] with lots of comments",
		in: `# list comment
list: [
  # first comment
  {},  # first inline comment
  # second comment
  {} # second inline comment
  # last comment
] # list inline comment
# other comment`,
		out: `# list comment
list: [
  # first comment
  {},  # first inline comment
  # second comment
  {}  # second inline comment
  # last comment
]  # list inline comment
# other comment
`}, {
		name: "list notation with [{}] with inline children",
		in: `children: [           { name: "node_2.1" }           ,       {         name: "node_2.2"          },{name:"node_2.3"}]
`,
		out: `children: [ { name: "node_2.1" }, { name: "node_2.2" }, { name: "node_2.3" } ]
`}, {
		name: "list notation with [{}] without comma separators between multiline children",
		in: `children: [
  {
    name: "node_2.1"
  }
  {
    name: "node_2.2"
  }
  {
    name: "node_2.3"
  }
]
`,
		out: `children: [
  {
    name: "node_2.1"
  },
  {
    name: "node_2.2"
  },
  {
    name: "node_2.3"
  }
]
`}, {
		name: "list notation with [{}]",
		in: `children: [


  # Line 1



  # Line 2



  # Line 3



  {
    name: "node_1"
  },
  {


    name: "node_2"
    children: [  {            name:             "node_2.1" },         {name:"node_2.2"},{name:"node_2.3"        }]


  },
  {
    name: "node_3"
children    : [

      {
        name: "node_3.1"
      }, # after-node comment.



  # Line 1



  # Line 2



      {
        name: "node_3.2",
      },




      {
        name: "node_3.3"
      }
    ]
  }
]
`,
		out: `children: [
  # Line 1

  # Line 2

  # Line 3
  {
    name: "node_1"
  },
  {

    name: "node_2"
    children: [ { name: "node_2.1" }, { name: "node_2.2" }, { name: "node_2.3" } ]

  },
  {
    name: "node_3"
    children: [
      {
        name: "node_3.1"
      },  # after-node comment.

      # Line 1

      # Line 2

      {
        name: "node_3.2"
      },

      {
        name: "node_3.3"
      }
    ]
  }
]
`}, {
		name: "multiline string",
		in: `
name: "Foo"
description:
  "Foo is an open-source, scalable, and efficient storage solution "
  "for the web. It is based on MySQL—so it supports major MySQL features "
  "like transactions, indexes, and joins—but it also provides the scalability "
  "of NoSQL. As such, Foo offers the best of both the RDBMS and NoSQL "
  "worlds."
`,
		out: `name: "Foo"
description:
  "Foo is an open-source, scalable, and efficient storage solution "
  "for the web. It is based on MySQL—so it supports major MySQL features "
  "like transactions, indexes, and joins—but it also provides the scalability "
  "of NoSQL. As such, Foo offers the best of both the RDBMS and NoSQL "
  "worlds."
`}, {
		name: "escaped \"",
		in: `
  check_contents: {
    required_regexp: "\\s*syntax\\s*=\\s*\".*\""
  }
`,
		out: `check_contents: {
  required_regexp: "\\s*syntax\\s*=\\s*\".*\""
}
`}, {
		name: "single-quote inside a double-quote-delimited string (and vice-versa)",
		in: `
description:
    "foo's fork of mod_python's Cookie submodule. "
    "as well as \"marshalled\" cookies -- cookies that contain marshalled python "
    'double quote " inside single-quote-delimited string'
`,
		out: `description:
  "foo's fork of mod_python's Cookie submodule. "
  "as well as \"marshalled\" cookies -- cookies that contain marshalled python "
  "double quote \" inside single-quote-delimited string"
`}, {
		name: "list with inline comments; comments after list and after last child",
		in: `
tricorder: {
  options: {
        build_args: [
						# Other build_args comment.
            # LT.IfChange
            "first line",
            # LT.ThenChange(//foo)
            "--config=android_x86",  # Inline comment for android_x86.
            "--config=android_release"  # Inline comment for last child.
            # Comment after list.
        ]
      }
    }
`,
		out: `tricorder: {
  options: {
    build_args: [
      # Other build_args comment.
      # LT.IfChange
      "first line",
      # LT.ThenChange(//foo)
      "--config=android_x86",  # Inline comment for android_x86.
      "--config=android_release"  # Inline comment for last child.
      # Comment after list.
    ]
  }
}
`}, {
		name: "';' at end of value",
		in:   `name: "value";`,
		out: `name: "value"
`}, {
		name: "multi-line string with inline comments",
		in: `# cm
      options:
        "first line" # first comment
        "second line" # second comment
    `,
		out: `# cm
options:
  "first line"  # first comment
  "second line"  # second comment

`}, {
		name: "all kinds of inline comments",
		in: `# presubmit pre comment 1
# presubmit pre comment 2
presubmit: {
  # review pre comment 1
  # review pre comment 2
  review_notify: "review_notify_value" # review inline comment
	# comment for project
  project: [
    # project1 pre comment 1
    # project1 pre comment 2
    "project1", # project1 inline comment
    # project2 pre comment 1
    # project2 pre comment 2
    "project2" # project2 inline comment
    # after comment 1
    # after comment 2
  ]
  # description pre comment 1
  # description pre comment 2
  description: "line1" # line1 inline comment
    # line2 pre comment 1
    # line2 pre comment 2
    "line2" # line2 inline comment
  # after comment 1
  # after comment 2
	name { name: value } # inline comment
} # inline comment
`,
		out: `# presubmit pre comment 1
# presubmit pre comment 2
presubmit: {
  # review pre comment 1
  # review pre comment 2
  review_notify: "review_notify_value"  # review inline comment
  # comment for project
  project: [
    # project1 pre comment 1
    # project1 pre comment 2
    "project1",  # project1 inline comment
    # project2 pre comment 1
    # project2 pre comment 2
    "project2"  # project2 inline comment
    # after comment 1
    # after comment 2
  ]
  # description pre comment 1
  # description pre comment 2
  description:
    "line1"  # line1 inline comment
    # line2 pre comment 1
    # line2 pre comment 2
    "line2"  # line2 inline comment
  # after comment 1
  # after comment 2
  name { name: value }  # inline comment
}  # inline comment
`}, {
		name: "more ';'",
		in: `list_with_semicolon: [one, two];
string_with_semicolon: "str one";
multi_line_with_semicolon: "line 1"
  "line 2";
other_name: other_value`,
		out: `list_with_semicolon: [one, two]
string_with_semicolon: "str one"
multi_line_with_semicolon:
  "line 1"
  "line 2"
other_name: other_value
`}, {
		name: "keep lists",
		in: `list_two_items_inline: [one, two];
 list_two_items: [one,
 two];
list_one_item: [one]
list_one_item_multiline: [
one]
list_one_item_inline_comment: [one # with inline comment
]
list_one_item_pre_comment: [
# one item comment
one
]
list_one_item_post_comment: [
one
# post comment
]
list_no_item: []
list_no_item: [
]
# comment
list_no_item_comment: [
# as you can see there are no items
]
list_no_item_inline_comment: [] # Nothing here`,
		out: `list_two_items_inline: [one, two]
list_two_items: [
  one,
  two
]
list_one_item: [one]
list_one_item_multiline: [
  one
]
list_one_item_inline_comment: [
  one  # with inline comment
]
list_one_item_pre_comment: [
  # one item comment
  one
]
list_one_item_post_comment: [
  one
  # post comment
]
list_no_item: []
list_no_item: [
]
# comment
list_no_item_comment: [
  # as you can see there are no items
]
list_no_item_inline_comment: []  # Nothing here
`}, {
		name: "',' as field separator",
		in: `# cm

presubmit: {
  path_expression:  "...",
	other: "other"
}`,
		out: `# cm

presubmit: {
  path_expression: "..."
  other: "other"
}
`}, {
		name: "comment between name and value",
		in: `
    address: "address"
    options:
      # LT.IfChange
      "first line"
      # LT.ThenChange(//foo)
    other: OTHER
`,
		out: `address: "address"
options:
  # LT.IfChange
  "first line"
# LT.ThenChange(//foo)
other: OTHER
`}, {
		name: "another example of comment between name and value",
		in: `
    address: # comment
      "value"
    options: # comment
      "line 1"
      "line 2"
`,
		out: `address:
  # comment
  "value"
options:
  # comment
  "line 1"
  "line 2"
`}, {
		name: "new line with spaces between comment and value",
		in: `
  # comment
  
  check_tests: {
  }
`,
		out: `# comment
check_tests: {
}
`}, {
		name: "proto extension",
		in: `[foo.bar.Baz] {
}`,
		out: `[foo.bar.Baz] {
}
`}, {
		name: "multiple nested in the same line",
		in: `
    expr {
      union { userset { ref { relation: "_this" } } }
    }
`,
		out: `expr {
  union { userset { ref { relation: "_this" } } }
}
`}, {
		name: "comment on the last line without new line at the end of the file",
		in:   `name: "value"  # comment`,
		out: `name: "value"  # comment
`}, {
		name: "white space inside extension name",
		in: `[foo.Bar.
		Baz] {
	name: "value"
}
`,
		out: `[foo.Bar.Baz] {
  name: "value"
}
`}, {
		name: "one blank line at the end",
		in: `presubmit {
}

`,
		out: `presubmit {
}

`}, {
		name: "template directive",
		in: `[ext]: {
    offset: %offset%
    %if (offset < 0)% %for i : offset_count%
    # directive comment
		%if enabled%

    # innermost comment
    # innermost comment

    offset_type: PACKETS
		%end%
    %end% %end%

    # my comment
    # my comment

    %for (leading_timestamps : leading_timestamps_array)%
    leading_timestamps: %leading_timestamps.timestamp%
    %end%
    }
`,
		out: `[ext]: {
  offset: %offset%
  %if (offset < 0)%
  %for i : offset_count%
  # directive comment
  %if enabled%

  # innermost comment
  # innermost comment

  offset_type: PACKETS
  %end%
  %end%
  %end%

  # my comment
  # my comment

  %for (leading_timestamps : leading_timestamps_array)%
  leading_timestamps: %leading_timestamps.timestamp%
  %end%
}
`}, {
		name: "template directive with >",
		in: `[ext]: {
    offset: %offset%
    %if (offset > 0)% %for i : offset_count%
    # directive comment
		%if enabled%

    # innermost comment
    # innermost comment

    offset_type: PACKETS
		%end%
    %end% %end%

    # my comment
    # my comment

    %for (leading_timestamps : leading_timestamps_array)%
    leading_timestamps: %leading_timestamps.timestamp%
    %end%
    }
`,
		out: `[ext]: {
  offset: %offset%
  %if (offset > 0)%
  %for i : offset_count%
  # directive comment
  %if enabled%

  # innermost comment
  # innermost comment

  offset_type: PACKETS
  %end%
  %end%
  %end%

  # my comment
  # my comment

  %for (leading_timestamps : leading_timestamps_array)%
  leading_timestamps: %leading_timestamps.timestamp%
  %end%
}
`}, {
		name: "template directive with >=",
		in: `[ext]: {
    offset: %offset%
    %if (offset >= 0)% %for i : offset_count%
    # directive comment
		%if enabled%

    # innermost comment
    # innermost comment

    offset_type: PACKETS
		%end%
    %end% %end%

    # my comment
    # my comment

    %for (leading_timestamps : leading_timestamps_array)%
    leading_timestamps: %leading_timestamps.timestamp%
    %end%
    }
`,
		out: `[ext]: {
  offset: %offset%
  %if (offset >= 0)%
  %for i : offset_count%
  # directive comment
  %if enabled%

  # innermost comment
  # innermost comment

  offset_type: PACKETS
  %end%
  %end%
  %end%

  # my comment
  # my comment

  %for (leading_timestamps : leading_timestamps_array)%
  leading_timestamps: %leading_timestamps.timestamp%
  %end%
}
`}, {
		name: "template directive escaped",
		in: `node {
			name: %"value \"% value"%
		}
`,
		out: `node {
  name: %"value \"% value"%
}
`}, {
		name: "no_directives (as opposed to next tests)",
		in: `presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
		out: `presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`}, {
		name: "expand_all_children",
		in: `# txtpbfmt: expand_all_children
presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
		out: `# txtpbfmt: expand_all_children
presubmit: {
  check_presubmit_service: {
    address: "address"
    failure_status: WARNING
    options: "options"
  }
}
`}, {
		name: "skip_all_colons",
		in: `# txtpbfmt: skip_all_colons
presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
		out: `# txtpbfmt: skip_all_colons
presubmit { check_presubmit_service { address: "address" failure_status: WARNING options: "options" } }
`}, {
		name: "separate_directives",
		in: `# txtpbfmt: expand_all_children
# txtpbfmt: skip_all_colons
presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
		out: `# txtpbfmt: expand_all_children
# txtpbfmt: skip_all_colons
presubmit {
  check_presubmit_service {
    address: "address"
    failure_status: WARNING
    options: "options"
  }
}
`}, {
		name: "combined_directives",
		in: `# txtpbfmt: expand_all_children, skip_all_colons
presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
		out: `# txtpbfmt: expand_all_children, skip_all_colons
presubmit {
  check_presubmit_service {
    address: "address"
    failure_status: WARNING
    options: "options"
  }
}
`}, {
		name: "repeated proto format",
		in: `{
  		a: 1;
  		b: 2}
  	{ a: 1;
  	b: 4
  	nested: {
    	 # nested_string is optional
    	nested_string: "foo"
  	}
  	}
  	{ a: 1;
  	b:4,
  	c: 5}`,
		out: `{
  a: 1
  b: 2
}
{
  a: 1
  b: 4
  nested: {
    # nested_string is optional
    nested_string: "foo"
  }
}
{
  a: 1
  b: 4
  c: 5
}
`}, {
		name: "repeated proto format with short messages",
		in: `{  		a: 1}
  	{ a: 2  }
  	{ a: 3}`,
		out: `{ a: 1 }
{ a: 2 }
{ a: 3 }
`}, {
		name: "allow unnamed nodes everywhere",
		in: `
# txtpbfmt: allow_unnamed_nodes_everywhere
mapping {{
  my_proto_field: "foo"
}}`,
		out: `# txtpbfmt: allow_unnamed_nodes_everywhere
mapping {
  {
    my_proto_field: "foo"
  }
}
`}, {
		name: "sort fields and values",
		in: `# txtpbfmt: sort_fields_by_field_name
# txtpbfmt: sort_repeated_fields_by_content
presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    # Should go after ADD
    operation: EDIT
    operation: ADD
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should go before reviewerB
  auto_reviewers: "reviewerA"
}
`,
		out: `# txtpbfmt: sort_fields_by_field_name
# txtpbfmt: sort_repeated_fields_by_content
presubmit: {
  # Should go before reviewerB
  auto_reviewers: "reviewerA"
  auto_reviewers: "reviewerB"
  check_contents: {
    check_delta_only: true
    operation: ADD
    # Should go after ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
  }
}
`}, {
		name: "sort and remove duplicates",
		in: `# txtpbfmt: sort_fields_by_field_name
# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: remove_duplicate_values_for_repeated_fields
presubmit: {
  auto_reviewers: "reviewerB"
  # Should go before reviewerB
  auto_reviewers: "reviewerA"
  check_contents: {
    operation: EDIT
    operation: ADD
    # Should be removed
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
    # Should go before operation: ADD
    check_delta_only: true
  }
  # Should be removed
  auto_reviewers: "reviewerA"
}
`,
		out: `# txtpbfmt: sort_fields_by_field_name
# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: remove_duplicate_values_for_repeated_fields
presubmit: {
  # Should go before reviewerB
  auto_reviewers: "reviewerA"
  auto_reviewers: "reviewerB"
  check_contents: {
    # Should go before operation: ADD
    check_delta_only: true
    operation: ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
  }
}
`}, {
		name: "trailing comma / semicolon",
		in: `dict: {
	arg: {
		key: "first_value"
		value: { num: 0 }
	},
	arg: {
		key: "second_value"
		value: { num: 1 }
	};
}
`,
		out: `dict: {
  arg: {
    key: "first_value"
    value: { num: 0 }
  }
  arg: {
    key: "second_value"
    value: { num: 1 }
  }
}
`}}
	for _, input := range inputs {
		out, err := Format([]byte(input.in))
		if err != nil {
			t.Errorf("Format[%s] %v returned err %v", input.name, input.in, err)
			continue
		}
		if diff := diff.Diff(input.out, string(out)); diff != "" {
			nodes, err := Parse([]byte(input.in))
			if err != nil {
				t.Errorf("Parse[%s] %v returned err %v", input.name, input.in, err)
				continue
			}
			t.Errorf("Format[%s](\n%s\n)\nparsed tree\n%s\n\nreturned diff (-want, +got):\n%s", input.name, input.in, DebugFormat(nodes, 0), diff)
		}
	}
}

func TestParserConfigs(t *testing.T) {
	inputs := []struct {
		name    string
		in      string
		config  Config
		out     string
		wantErr string
	}{{
		name: "AlreadyExpandedConfigOff",
		in: `presubmit: {
  check_presubmit_service: {
    address: "address"
    failure_status: WARNING
    options: "options"
  }
}
`,
		config: Config{ExpandAllChildren: false},
		out: `presubmit: {
  check_presubmit_service: {
    address: "address"
    failure_status: WARNING
    options: "options"
  }
}
`,
	}, {
		name: "AlreadyExpandedConfigOn",
		in: `presubmit: {
  check_presubmit_service: {
    address: "address"
    failure_status: WARNING
    options: "options"
  }
}
`,
		config: Config{ExpandAllChildren: true},
		out: `presubmit: {
  check_presubmit_service: {
    address: "address"
    failure_status: WARNING
    options: "options"
  }
}
`,
	}, {
		name: "NotExpandedOnFix",
		in: `presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
		config: Config{ExpandAllChildren: false},
		out: `presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
	}, {
		name: "ExpandedOnFix",
		in: `presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
		config: Config{ExpandAllChildren: true},
		out: `presubmit: {
  check_presubmit_service: {
    address: "address"
    failure_status: WARNING
    options: "options"
  }
}
`,
	}, {
		name: "SortFieldNames",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    operation: EDIT
    operation: ADD
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should remain below reviewerB
  auto_reviewers: "reviewerA"
}
`,
		config: Config{SortFieldsByFieldName: true},
		out: `presubmit: {
  auto_reviewers: "reviewerB"
  # Should remain below reviewerB
  auto_reviewers: "reviewerA"
  check_contents: {
    check_delta_only: true
    operation: EDIT
    operation: ADD
    prohibited_regexp: "UnsafeFunction"
  }
}
`,
	}, {
		name: "SortFieldContents",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    # Should go after ADD
    operation: EDIT
    operation: ADD
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should remain below
  auto_reviewers: "reviewerA"
}
`,
		config: Config{SortRepeatedFieldsByContent: true},
		out: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    operation: ADD
    # Should go after ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should remain below
  auto_reviewers: "reviewerA"
}
`,
	}, {
		name: "SortFieldNamesAndContents",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    # Should go after ADD
    operation: EDIT
    operation: ADD
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should go before reviewerB
  auto_reviewers: "reviewerA"
}
`,
		config: Config{SortFieldsByFieldName: true, SortRepeatedFieldsByContent: true},
		out: `presubmit: {
  # Should go before reviewerB
  auto_reviewers: "reviewerA"
  auto_reviewers: "reviewerB"
  check_contents: {
    check_delta_only: true
    operation: ADD
    # Should go after ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
  }
}
`,
	}, {
		name: "SortBySpecifiedFieldOrder",
		in: `presubmit: {
  unmoved_not_in_config: "foo"
  check_contents: {
    # Not really at the top; attached to x_third
    x_third: "3"
    # attached to EDIT
    z_first: EDIT
    unknown_bubbles_to_top: true
    x_second: true
    z_first: ADD
    also_unknown_bubbles_to_top: true
    # Trailing comment is on different node; should not confuse ordering logic.
    # These always sort below fields in the sorting config, and thus stay at bottom.
  }
  # Should also not move
  unmoved_not_in_config: "bar"
}
`,
		config: Config{
			fieldSortOrder: map[string]map[string]int{
				"check_contents": makePriorities("z_first", "x_second", "x_third"),
			},
		},
		// Nodes are sorted by the specified order, else left untouched.
		out: `presubmit: {
  unmoved_not_in_config: "foo"
  check_contents: {
    unknown_bubbles_to_top: true
    also_unknown_bubbles_to_top: true
    # attached to EDIT
    z_first: EDIT
    z_first: ADD
    x_second: true
    # Not really at the top; attached to x_third
    x_third: "3"
    # Trailing comment is on different node; should not confuse ordering logic.
    # These always sort below fields in the sorting config, and thus stay at bottom.
  }
  # Should also not move
  unmoved_not_in_config: "bar"
}
`,
	}, {
		name: "SortBySpecifiedFieldOrderAndNameAndValue",
		in: `presubmit: {
  # attached to bar
  sort_by_name_and_value: "bar"
  check_contents: {
    x_third: "3"
    # attached to EDIT
    z_first: EDIT
    unknown_bubbles_to_top: true
    x_second: true
    z_first: ADD
    also_unknown_bubbles_to_top: true
    # The nested check_contents bubbles to the top, since it's not in the fieldSortOrder.
    check_contents: {
      x_second: true
      z_first: ADD
    }
  }
  sort_by_name_and_value: "foo"
}
`,
		config: Config{
			fieldSortOrder: map[string]map[string]int{
				"check_contents": makePriorities("z_first", "x_second", "x_third", "not_required"),
			},
			SortFieldsByFieldName:       true,
			SortRepeatedFieldsByContent: true,
		},
		// Nodes are sorted by name/value first, then by the specified order. Hence the specified
		// repeated fields (z_first) is also sorted by value rather than in original order.
		out: `presubmit: {
  check_contents: {
    also_unknown_bubbles_to_top: true
    # The nested check_contents bubbles to the top, since it's not in the fieldSortOrder.
    check_contents: {
      z_first: ADD
      x_second: true
    }
    unknown_bubbles_to_top: true
    z_first: ADD
    # attached to EDIT
    z_first: EDIT
    x_second: true
    x_third: "3"
  }
  # attached to bar
  sort_by_name_and_value: "bar"
  sort_by_name_and_value: "foo"
}
`,
	}, {
		name: "SortBySpecifiedFieldOrderErrorHandling",
		in: `presubmit: {
  node_not_in_config_will_not_trigger_error: true
  check_contents: {
    x_third: "3"
    z_first: EDIT
    unknown_field_triggers_error: true
    x_second: true
    z_first: ADD
  }
}
`,
		config: Config{
			fieldSortOrder: map[string]map[string]int{
				"check_contents": makePriorities("z_first", "x_second", "x_third"),
			},
			RequireFieldSortOrderToMatchAllFieldsInNode: true,
		},
		wantErr: `parent field: "check_contents", unsorted field: "unknown_field_triggers_error"`,
	}, {
		name: "RemoveRepeats",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  auto_reviewers: "reviewerA"
  check_contents: {
    operation: EDIT
    operation: ADD
    # Should be removed
		operation: EDIT
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should be removed
  auto_reviewers: "reviewerA"
}
`,
		config: Config{RemoveDuplicateValuesForRepeatedFields: true},
		out: `presubmit: {
  auto_reviewers: "reviewerB"
  auto_reviewers: "reviewerA"
  check_contents: {
    operation: EDIT
    operation: ADD
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
}
`,
	}, {
		name: "SortEverythingAndRemoveRepeats",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  # Should go before reviewerB
  auto_reviewers: "reviewerA"
  check_contents: {
    operation: EDIT
    operation: ADD
    # Should be removed
		operation: EDIT
    prohibited_regexp: "UnsafeFunction"
    # Should go before operation: ADD
    check_delta_only: true
  }
  # Should be removed
  auto_reviewers: "reviewerA"
}
`,
		config: Config{
			SortFieldsByFieldName:                  true,
			SortRepeatedFieldsByContent:            true,
			RemoveDuplicateValuesForRepeatedFields: true},
		out: `presubmit: {
  # Should go before reviewerB
  auto_reviewers: "reviewerA"
  auto_reviewers: "reviewerB"
  check_contents: {
    # Should go before operation: ADD
    check_delta_only: true
    operation: ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
  }
}
`,
	}, {
		name: "TripleQuotedStrings",
		config: Config{
			AllowTripleQuotedStrings: true,
		},
		in: `foo: """bar"""
`,
		out: `foo: """bar"""
`,
	}, {
		name: "TripleQuotedStrings_multiLine",
		config: Config{
			AllowTripleQuotedStrings: true,
		},
		in: `foo: """
  bar
"""
`,
		out: `foo: """
  bar
"""
`,
	}, {
		name: "TripleQuotedStrings_singleQuotes",
		config: Config{
			AllowTripleQuotedStrings: true,
		},
		in: `foo: '''
  bar
'''
`,
		out: `foo: '''
  bar
'''
`,
	}, {
		name: "TripleQuotedStrings_brackets",
		config: Config{
			AllowTripleQuotedStrings: true,
		},
		in: `s: """ "}" """
`,
		out: `s: """ "}" """
`,
	}, {
		name: "TripleQuotedStrings_metaComment",
		in: `# txtpbfmt: allow_triple_quoted_strings
foo: """
  bar
"""
`,
		out: `# txtpbfmt: allow_triple_quoted_strings
foo: """
  bar
"""
`,
	}, {
		name: "WrapStrings",
		config: Config{
			WrapStringsAtColumn: 15,
		},
		in: `# Comments are not wrapped
s: "one two three four five"
`,
		out: `# Comments are not wrapped
s:
  "one two "
  "three four "
  "five"
`,
	}, {
		name: "WrapStrings_inlineChildren",
		config: Config{
			WrapStringsAtColumn: 14,
		},
		in: `root { inner { s: "89 1234" } }
`,
		out: `root { inner { s: "89 1234" } }
`,
	}, {
		name: "WrapStrings_exactlyNumColumnsDoesNotWrap",
		config: Config{
			WrapStringsAtColumn: 14,
		},
		in: `root {
  inner {
    s: "89 123"
  }
}
`,
		out: `root {
  inner {
    s: "89 123"
  }
}
`,
	}, {
		name: "WrapStrings_numColumnsPlus1Wraps",
		config: Config{
			WrapStringsAtColumn: 14,
		},
		in: `root {
  inner {
    s:
      "89 123 "
      "123 56"
  }
}
`,
		out: `root {
  inner {
    s:
      "89 "
      "123 "
      "123 "
      "56"
  }
}
`,
	}, {
		name: "WrapStrings_commentKeptWhenLinesReduced",
		config: Config{
			WrapStringsAtColumn: 15,
		},
		in: `root {
  label:
    # before
    "56789 next-line"  # trailing
  label:
    # inside top
    "56789 next-line "  # trailing line1
    "v "
    "v "
    "v "
    "v "
    # inside in-between
    "straggler"  # trailing line2
    # next-node comment
}
`,
		out: `root {
  label:
    # before
    "56789 "  # trailing
    "next-line"
  label:
    # inside top
    "56789 "  # trailing line1
    "next-line "
    "v v v v "
    "straggler"
    # inside in-between
    # trailing line2
  # next-node comment
}
`,
	}, {
		name: "WrapStrings_doNotBreakLongWords",
		config: Config{
			WrapStringsAtColumn: 15,
		},
		in: `s: "one@two_three-four&five"
`,
		out: `s: "one@two_three-four&five"
`,
	}, {
		name: "WrapStrings_wrapHtml",
		config: Config{
			WrapStringsAtColumn: 15,
			WrapHTMLStrings:     true,
		},
		in: `s: "one two three <four>"
`,
		out: `s:
  "one two "
  "three "
  "<four>"
`,
	}, {
		name: "WrapStrings_empty",
		config: Config{
			WrapStringsAtColumn: 15,
		},
		in: `s: 
`,
		out: `s: 
`,
	}, {
		name: "WrapStrings_doNoWrapHtmlByDefault",
		config: Config{
			WrapStringsAtColumn: 15,
		},
		in: `s: "one two three <four>"
`,
		out: `s: "one two three <four>"
`,
	}, {
		name: "WrapStrings_metaComment",
		in: `# txtpbfmt: wrap_strings_at_column=15
# txtpbfmt: wrap_html_strings
s: "one two three <four>"
`,
		out: `# txtpbfmt: wrap_strings_at_column=15
# txtpbfmt: wrap_html_strings
s:
  "one two "
  "three "
  "<four>"
`,
	}, {
		name: "WrapStrings_doNotWrapNonStrings",
		config: Config{
			WrapStringsAtColumn: 15,
		},
		in: `e: VERY_LONG_ENUM_VALUE
i: 12345678901234567890
r: [1, 2, 3, 4, 5, 6, 7, 8, 9, 0]
`,
		out: `e: VERY_LONG_ENUM_VALUE
i: 12345678901234567890
r: [1, 2, 3, 4, 5, 6, 7, 8, 9, 0]
`,
	}, {
		name: "WrapStrings_alreadyWrappedStringsAreNotRewrapped",
		config: Config{
			WrapStringsAtColumn: 15,
		},
		// Total length >15, but each existing line <15, so don't re-wrap 1st line to "I am ".
		in: `s:
  "I "
  "am already "
  "wrapped"
`,
		out: `s:
  "I "
  "am already "
  "wrapped"
`,
	}, {
		name: "WrapStrings_tripleQuotedStringsAreNotWrapped",
		config: Config{
			WrapStringsAtColumn:      15,
			AllowTripleQuotedStrings: true,
		},
		in: `s1: """one two three four five"""
s2: '''six seven eight nine'''
`,
		out: `s1: """one two three four five"""
s2: '''six seven eight nine'''
`,
	}, {
		name: "PreserveAngleBrackets",
		config: Config{
			PreserveAngleBrackets: true,
		},
		in: `foo <
  a: 1
>
foo {
  b: 2
}
`,
		out: `foo <
  a: 1
>
foo {
  b: 2
}
`,
	}, {
		name: "legacy quote behavior",
		config: Config{
			SmartQuotes: false,
		},
		in: `foo: "\"bar\""`,
		out: `foo: "\"bar\""
`,
	}, {
		name: "smart quotes",
		config: Config{
			SmartQuotes: true,
		},
		in: `foo: "\"bar\""`,
		out: `foo: '"bar"'
`,
	}, {
		name: "smart quotes via meta comment",
		config: Config{
			SmartQuotes: false,
		},
		in: `# txtpbfmt: smartquotes
foo: "\"bar\""`,
		out: `# txtpbfmt: smartquotes
foo: '"bar"'
`,
	},
	}
	// Test FormatWithConfig with inputs.
	for _, input := range inputs {
		got, err := FormatWithConfig([]byte(input.in), input.config)
		if input.wantErr != "" {
			if err == nil {
				t.Errorf("FormatWithConfig[%s] got err=nil, want err=%v", input.name, input.wantErr)
				continue
			}
			if !strings.Contains(err.Error(), input.wantErr) {
				t.Errorf("FormatWithConfig[%s] got err=%v, want err=%v", input.name, err, input.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("FormatWithConfig[%s] %v with config %v returned err %v", input.name, input.in, input.config, err)
			continue
		}
		if diff := diff.Diff(input.out, string(got)); diff != "" {
			t.Errorf("FormatWithConfig[%s](\n%s\n)\nreturned different output from expected (-want, +got):\n%s", input.name, input.in, diff)
		}
	}
	// Test ParseWithConfig with inputs.
	for _, input := range inputs {
		nodes, err := ParseWithConfig([]byte(input.in), input.config)
		if input.wantErr != "" {
			if err == nil {
				t.Errorf("ParseWithConfig[%s] got err=nil, want err=%v", input.name, input.wantErr)
				continue
			}
			if !strings.Contains(err.Error(), input.wantErr) {
				t.Errorf("ParseWithConfig[%s] got err=%v, want err=%v", input.name, err, input.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseWithConfig[%s] %v with config %v returned err %v", input.name, input.in, input.config, err)
			continue
		}
		got := Pretty(nodes, 0)
		if diff := diff.Diff(input.out, got); diff != "" {
			t.Errorf("ParseWithConfig[%s](\n%s\n)\nreturned different Pretty output from expected (-want, +got):\n%s", input.name, input.in, diff)
		}
	}
}

func TestDebugFormat(t *testing.T) {
	inputs := []struct {
		in   string
		want string
	}{{
		in: `name: { name: "value" }`,
		want: `
 name: "name"
 PreComments: "" (len 0)
 children:
. name: "name"
. PreComments: "" (len 0)
. values: [{Value: "\"value\"", PreComments: "", InlineComment: ""}]
`,
	}}
	for _, input := range inputs {
		nodes, err := Parse([]byte(input.in))
		if err != nil {
			t.Errorf("Parse %v returned err %v", input.in, err)
			continue
		}
		if len(nodes) == 0 {
			t.Errorf("Parse %v returned no nodes", input.in)
			continue
		}
		got := DebugFormat(nodes, 0 /* depth */)
		if diff := diff.Diff(input.want, got); diff != "" {
			t.Errorf("DebugFormat %v returned diff (-want, +got):\n%s", input.in, diff)
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

func TestSmartQuotes(t *testing.T) {
	inputs := []struct {
		in         string
		wantLegacy string
		wantSmart  string
	}{
		{`1`, `1`, `1`},
		{`""`, `""`, `""`},
		{`''`, `""`, `""`},
		{`"a"`, `"a"`, `"a"`},
		{`'a'`, `"a"`, `"a"`},
		{`'a"b'`, `"a\"b"`, `'a"b'`},
		{`"a\"b"`, `"a\"b"`, `'a"b'`},
		{`'a\'b'`, `"a\'b"`, `"a'b"`},
		{`'a"b\'c'`, `"a\"b\'c"`, `"a\"b'c"`},
		{`"a\"b'c"`, `"a\"b'c"`, `"a\"b'c"`},
		{`'a\"b\'c'`, `"a\"b\'c"`, `"a\"b'c"`},
		{`"a\"b\'c"`, `"a\"b\'c"`, `"a\"b'c"`},
		{`"'\\\'"`, `"'\\\'"`, `"'\\'"`},
	}
	for _, tc := range inputs {
		in := `foo: ` + tc.in
		name := fmt.Sprintf("Format [ %s ] with legacy quote behavior", tc.in)
		want := `foo: ` + tc.wantLegacy
		gotRaw, err := FormatWithConfig([]byte(in), Config{SmartQuotes: false})
		got := strings.TrimSpace(string(gotRaw))
		if err != nil {
			t.Errorf("%s: got error: %s, want no error and [ %s ]", name, err, want)
		} else if got != want {
			t.Errorf("%s: got [ %s ], want [ %s ]", name, got, want)
		}

		name = fmt.Sprintf("Format [ %s ] with smart quote behavior", tc.in)
		want = `foo: ` + tc.wantSmart
		gotRaw, err = FormatWithConfig([]byte(`foo: `+tc.in), Config{SmartQuotes: true})
		got = strings.TrimSpace(string(gotRaw))
		if err != nil {
			t.Errorf("%s: got error: %s, want no error and [ %s ]", name, err, want)
		} else if got != want {
			t.Errorf("%s: got [ %s ], want [ %s ]", name, got, want)
		}
	}
}
