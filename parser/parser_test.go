// Open source parser tests.
package parser

import (
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
		in: `child_list_notation: [
		  {`,
		err: "[{}]",
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
		name: "relowed multiline string",
		in: `# txtpbfmt: enable_line_limit
name: "Foo"
description:
  "Foo is an open-source, scalable ⚖️, and efficient storage (📀) solution for the web. It is based on MySQL—so it supports major MySQL features "
  "like transactions, indexes, and joins—but it also provides the scalability of NoSQL. As such, Foo offers the best of both the RDBMS and NoSQL "
  "worlds."
`,
		out: `# txtpbfmt: enable_line_limit
name: "Foo"
description:
  "Foo is an open-source, scalable ⚖️, and efficient storage (📀) solution for "
  "the web. It is based on MySQL—so it supports major MySQL features "
  "like transactions, indexes, and joins—but it also provides the scalability "
  "of NoSQL. As such, Foo offers the best of both the RDBMS and NoSQL "
  "worlds."
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
`}, {
		name: "LineLimit in all kinds of comments",
		in: `# txtpbfmt: enable_line_limit
# presubmit pre comment  with a very long line that should exced the 80 character limit 1
# presubmit pre comment  with a very long line that should exced the 80 character limit 😮 (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
presubmit: {
# short review pre comment 1
  # review pre comment with a very long line that should exced the 80 character limit 1
# review pre comment  with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
  # short review pre comment 2
	# presubmit pre comment  with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
  review_notify: "review_notify_value" # review inline comment
	# short comment for project
	# comment for project with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
  project: [
    # short project1 pre comment 1
		  # project1 pre comment with a very long line that should exced the 80 character limit 1
    # short project1 pre comment 2
			# project1 pre comment with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
    "project1", # project1 inline comment
    # short project2 pre comment 1
		  # project2 pre comment with a very long line that should exced the 80 character limit 1
    # short project2 pre comment 2
			# project2 pre comment with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
    "project2" # project2 inline comment
		  # after comment with a very long line that should exced the 80 character limit 1
    # short after comment 1
			# after comment with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
    # short after comment 2
  ]
  # short description pre comment 1
		  # description pre comment with a very long line that should exced the 80 character limit 1
  # short description pre comment 2
			# description pre comment with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
  description:
	  "line1" # line1 inline comment
    # short line2 pre comment 1
		  # line2 pre comment with a very long line that should exced the 80 character limit 1
    # short line2 pre comment 2
			# line2 pre comment with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
    "line2" # line2 inline comment
  # short after comment 1
		  # after comment with a very long line that should exced the 80 character limit 1
  # short after comment 2
			# after comment with a very long line that should exced the 80 character limit (this should be exceded by more than 160 character to force the creation of more than 2 lines) 2
	name { name: value } # inline comment
} # inline comment
`,
		out: `# txtpbfmt: enable_line_limit
# presubmit pre comment  with a very long line that should exced the 80
# character limit 1
# presubmit pre comment  with a very long line that should exced the 80
# character limit 😮 (this should be exceded by more than 160 character to force
# the creation of more than 2 lines) 2
presubmit: {
  # short review pre comment 1
  # review pre comment with a very long line that should exced the 80 character
  # limit 1
  # review pre comment  with a very long line that should exced the 80 character
  # limit (this should be exceded by more than 160 character to force the
  # creation of more than 2 lines) 2
  # short review pre comment 2
  # presubmit pre comment  with a very long line that should exced the 80
  # character limit (this should be exceded by more than 160 character to force
  # the creation of more than 2 lines) 2
  review_notify: "review_notify_value"  # review inline comment
  # short comment for project
  # comment for project with a very long line that should exced the 80 character
  # limit (this should be exceded by more than 160 character to force the
  # creation of more than 2 lines) 2
  project: [
    # short project1 pre comment 1
    # project1 pre comment with a very long line that should exced the 80
    # character limit 1
    # short project1 pre comment 2
    # project1 pre comment with a very long line that should exced the 80
    # character limit (this should be exceded by more than 160 character to
    # force the creation of more than 2 lines) 2
    "project1",  # project1 inline comment
    # short project2 pre comment 1
    # project2 pre comment with a very long line that should exced the 80
    # character limit 1
    # short project2 pre comment 2
    # project2 pre comment with a very long line that should exced the 80
    # character limit (this should be exceded by more than 160 character to
    # force the creation of more than 2 lines) 2
    "project2"  # project2 inline comment
    # after comment with a very long line that should exced the 80 character
    # limit 1
    # short after comment 1
    # after comment with a very long line that should exced the 80 character
    # limit (this should be exceded by more than 160 character to force the
    # creation of more than 2 lines) 2
    # short after comment 2
  ]
  # short description pre comment 1
  # description pre comment with a very long line that should exced the 80
  # character limit 1
  # short description pre comment 2
  # description pre comment with a very long line that should exced the 80
  # character limit (this should be exceded by more than 160 character to force
  # the creation of more than 2 lines) 2
  description:
    "line1"  # line1 inline comment
    # short line2 pre comment 1
    # line2 pre comment with a very long line that should exced the 80 character
    # limit 1
    # short line2 pre comment 2
    # line2 pre comment with a very long line that should exced the 80 character
    # limit (this should be exceded by more than 160 character to force the
    # creation of more than 2 lines) 2
    "line2"  # line2 inline comment
  # short after comment 1
  # after comment with a very long line that should exced the 80 character limit
  # 1
  # short after comment 2
  # after comment with a very long line that should exced the 80 character limit
  # (this should be exceded by more than 160 character to force the creation of
  # more than 2 lines) 2
  name { name: value }  # inline comment
}  # inline comment
`},
	}
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
		name   string
		in     string
		config Config
		out    string
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
	}}
	// Test FormatWithConfig with inputs.
	for _, input := range inputs {
		got, err := FormatWithConfig([]byte(input.in), input.config)
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
		config := input.config
		nodes, err := ParseWithConfig([]byte(input.in), &config)
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
