// Open source parser tests.
// N.b.: take care when editing this file, as it contains significant trailing whitespace.
package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/kylelemons/godebug/pretty"
	"github.com/protocolbuffers/txtpbfmt/config"
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
	}, {
		in:  `# txtpbfmt: off`,
		err: "unterminated txtpbfmt off",
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
		name: "preserve angle brackets",
		in: `# txtpbfmt: preserve_angle_brackets
foo <
  a: 1
>
foo {
  b: 2
}
`,
		out: `# txtpbfmt: preserve_angle_brackets
foo <
  a: 1
>
foo {
  b: 2
}
`,
	}, {
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
    action: [
      # Should go after REVIEW
      SUBMIT,
      REVIEW
    ]
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
    action: [
      REVIEW,
      # Should go after REVIEW
      SUBMIT
    ]
    check_delta_only: true
    operation: ADD
    # Should go after ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
  }
}
`}, {
		name: "sort fields and values in reverse order",
		in: `# txtpbfmt: sort_fields_by_field_name
# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: reverse_sort
presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    action: [
      # Should go after SUBMIT
      REVIEW,
      SUBMIT
    ]
    # Should go after EDIT
    operation: ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should go after reviewerB
  auto_reviewers: "reviewerA"
}
`,
		out: `# txtpbfmt: sort_fields_by_field_name
# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: reverse_sort
presubmit: {
  check_contents: {
    prohibited_regexp: "UnsafeFunction"
    operation: EDIT
    # Should go after EDIT
    operation: ADD
    check_delta_only: true
    action: [
      SUBMIT,
      # Should go after SUBMIT
      REVIEW
    ]
  }
  auto_reviewers: "reviewerB"
  # Should go after reviewerB
  auto_reviewers: "reviewerA"
}
`}, {
		name: "reverse sort does nothing without another sort_* config option",
		in: `# txtpbfmt: reverse_sort
presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    action: [
      # Should go after SUBMIT
      REVIEW,
      SUBMIT
    ]
    # Should go after EDIT
    operation: ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should go after reviewerB
  auto_reviewers: "reviewerA"
}
`,
		out: `# txtpbfmt: reverse_sort
presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    action: [
      # Should go after SUBMIT
      REVIEW,
      SUBMIT
    ]
    # Should go after EDIT
    operation: ADD
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should go after reviewerB
  auto_reviewers: "reviewerA"
}
`}, {
		name: "blank lines are collapsed to one",
		in: `# txtpbfmt: sort_repeated_fields_by_content
presubmit: {
  check_contents: {


    # Should go after ADD.
    # And the empty lines above this are collapsed into one blank line.
    operation: EDIT
    operation: ADD



    operation: REMOVE
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_content
presubmit: {
  check_contents: {
    operation: ADD

    # Should go after ADD.
    # And the empty lines above this are collapsed into one blank line.
    operation: EDIT

    operation: REMOVE
  }
}
`}, {
		name: "blank line after comment separates it from the node",
		in: `# txtpbfmt: sort_repeated_fields_by_content
presubmit: {
  check_contents: {
    # This comment is a separate node and does not move when the fields are sorted.

    operation: EDIT
    operation: ADD
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_content
presubmit: {
  check_contents: {
    # This comment is a separate node and does not move when the fields are sorted.
    operation: ADD

    operation: EDIT
  }
}
`}, {
		name: "sort by subfield values",
		in: `# txtpbfmt: sort_repeated_fields_by_subfield=operation.name
# txtpbfmt: sort_repeated_fields_by_subfield=test.id
presubmit: {
  operation {
    name: EDIT
  }
  operation {
    name: ADD
  }
  test {
    id: 4
  }
  test {
    id: 2
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_subfield=operation.name
# txtpbfmt: sort_repeated_fields_by_subfield=test.id
presubmit: {
  operation {
    name: ADD
  }
  operation {
    name: EDIT
  }
  test {
    id: 2
  }
  test {
    id: 4
  }
}
`}, {
		name: "reverse sort by subfield values",
		in: `# txtpbfmt: sort_repeated_fields_by_subfield=operation.name
# txtpbfmt: sort_repeated_fields_by_subfield=test.id
# txtpbfmt: reverse_sort
presubmit: {
  operation {
    name: ADD
  }
  operation {
    name: EDIT
  }
  test {
    id: 2
  }
  test {
    id: 4
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_subfield=operation.name
# txtpbfmt: sort_repeated_fields_by_subfield=test.id
# txtpbfmt: reverse_sort
presubmit: {
  operation {
    name: EDIT
  }
  operation {
    name: ADD
  }
  test {
    id: 4
  }
  test {
    id: 2
  }
}
`}, {
		name: "sort by deeper subfield path",
		in: `# txtpbfmt: sort_repeated_fields_by_subfield=test.metadata.identifiers.id
presubmit: {
  test {
    metadata {
      identifiers {
        id: 4
      }
    }
  }
  test {
    metadata {
      identifiers {
        id: 2
      }
    }
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_subfield=test.metadata.identifiers.id
presubmit: {
  test {
    metadata {
      identifiers {
        id: 2
      }
    }
  }
  test {
    metadata {
      identifiers {
        id: 4
      }
    }
  }
}
`}, {
		name: "reverse sort by deeper subfield path",
		in: `# txtpbfmt: sort_repeated_fields_by_subfield=test.metadata.identifiers.id
# txtpbfmt: reverse_sort
presubmit: {
  test {
    metadata {
      identifiers {
        id: 2
      }
    }
  }
  test {
    metadata {
      identifiers {
        id: 4
      }
    }
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_subfield=test.metadata.identifiers.id
# txtpbfmt: reverse_sort
presubmit: {
  test {
    metadata {
      identifiers {
        id: 4
      }
    }
  }
  test {
    metadata {
      identifiers {
        id: 2
      }
    }
  }
}
`}, {
		name: "No sort for repeated final subfield",
		in: `# txtpbfmt: sort_repeated_fields_by_subfield=test.metadata.identifiers.id
presubmit: {
  test {
    metadata {
      identifiers {
        id: 3
        id: 4
      }
    }
  }
  test {
    metadata {
      identifiers {
        id: 1
        id: 2
      }
    }
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_subfield=test.metadata.identifiers.id
presubmit: {
  test {
    metadata {
      identifiers {
        id: 3
        id: 4
      }
    }
  }
  test {
    metadata {
      identifiers {
        id: 1
        id: 2
      }
    }
  }
}
`}, {
		name: "No sort for message final subfield",
		in: `# txtpbfmt: sort_repeated_fields_by_subfield=test.metadata.identifiers
presubmit: {
  test {
    metadata {
      identifiers {
        id: 4
      }
    }
  }
  test {
    metadata {
      identifiers {
        id: 2
      }
    }
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_subfield=test.metadata.identifiers
presubmit: {
  test {
    metadata {
      identifiers {
        id: 4
      }
    }
  }
  test {
    metadata {
      identifiers {
        id: 2
      }
    }
  }
}
`}, {
		// In this test multiple subfields of `test` are given. The expected behavior is: first sort by
		// test.id; in case of a tie, sort by test.type; in case of a tie again, sort by test.name.
		name: "sort by multiple subfield values",
		in: `# txtpbfmt: sort_repeated_fields_by_subfield=operation.name
# txtpbfmt: sort_repeated_fields_by_subfield=test.id
# txtpbfmt: sort_repeated_fields_by_subfield=test.type
# txtpbfmt: sort_repeated_fields_by_subfield=test.name
presubmit: {
  operation {
    name: EDIT
  }
  operation {
    name: ADD
  }
  test {
    id: 4
    name: bar
    unrelated_field: 1
    type: type_1
  }
  test {
    id: 2
    name: foo
    unrelated_field: 3
    type: type_2
  }
  test {
    id: 2
    name: baz
    unrelated_field: 2
    type: type_1
  }
  test {
    id: 2
    name: bar
    unrelated_field: 1
    type: type_2
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_subfield=operation.name
# txtpbfmt: sort_repeated_fields_by_subfield=test.id
# txtpbfmt: sort_repeated_fields_by_subfield=test.type
# txtpbfmt: sort_repeated_fields_by_subfield=test.name
presubmit: {
  operation {
    name: ADD
  }
  operation {
    name: EDIT
  }
  test {
    id: 2
    name: baz
    unrelated_field: 2
    type: type_1
  }
  test {
    id: 2
    name: bar
    unrelated_field: 1
    type: type_2
  }
  test {
    id: 2
    name: foo
    unrelated_field: 3
    type: type_2
  }
  test {
    id: 4
    name: bar
    unrelated_field: 1
    type: type_1
  }
}
`}, {
		// In this test multiple subfields of `test` are given. The expected behavior is: first reverse
		// sort by test.id; in case of a tie, reverse sort by test.type; in case of a tie again, reverse
		// sort by test.name.
		name: "reverse sort by multiple subfield values",
		in: `# txtpbfmt: sort_repeated_fields_by_subfield=operation.name
# txtpbfmt: sort_repeated_fields_by_subfield=test.id
# txtpbfmt: sort_repeated_fields_by_subfield=test.type
# txtpbfmt: reverse_sort
# txtpbfmt: sort_repeated_fields_by_subfield=test.name
presubmit: {
  operation {
    name: ADD
  }
  operation {
    name: EDIT
  }
  test {
    id: 2
    name: foo
    unrelated_field: 3
    type: type_2
  }
  test {
    id: 4
    name: bar
    unrelated_field: 1
    type: type_1
  }
  test {
    id: 2
    name: baz
    unrelated_field: 2
    type: type_1
  }
  test {
    id: 2
    name: bar
    unrelated_field: 1
    type: type_2
  }
}
`,
		out: `# txtpbfmt: sort_repeated_fields_by_subfield=operation.name
# txtpbfmt: sort_repeated_fields_by_subfield=test.id
# txtpbfmt: sort_repeated_fields_by_subfield=test.type
# txtpbfmt: reverse_sort
# txtpbfmt: sort_repeated_fields_by_subfield=test.name
presubmit: {
  operation {
    name: EDIT
  }
  operation {
    name: ADD
  }
  test {
    id: 4
    name: bar
    unrelated_field: 1
    type: type_1
  }
  test {
    id: 2
    name: foo
    unrelated_field: 3
    type: type_2
  }
  test {
    id: 2
    name: bar
    unrelated_field: 1
    type: type_2
  }
  test {
    id: 2
    name: baz
    unrelated_field: 2
    type: type_1
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
		name: "reverse sort and remove duplicates",
		in: `# txtpbfmt: sort_fields_by_field_name
# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: reverse_sort
# txtpbfmt: remove_duplicate_values_for_repeated_fields
presubmit: {
  auto_reviewers: "reviewerA"
  # Should go before reviewerA
  auto_reviewers: "reviewerB"
  check_contents: {
    operation: ADD
    operation: EDIT
    # Should be removed
    operation: EDIT
    prohibited_regexp: "UnsafeFunction"
    # Should go after operation: ADD
    check_delta_only: true
  }
  # Should be removed
  auto_reviewers: "reviewerA"
}
`,
		out: `# txtpbfmt: sort_fields_by_field_name
# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: reverse_sort
# txtpbfmt: remove_duplicate_values_for_repeated_fields
presubmit: {
  check_contents: {
    prohibited_regexp: "UnsafeFunction"
    operation: EDIT
    operation: ADD
    # Should go after operation: ADD
    check_delta_only: true
  }
  # Should go before reviewerA
  auto_reviewers: "reviewerB"
  auto_reviewers: "reviewerA"
}
`}, {
		name: "multiple groups of repeated fields",
		in: `# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: sort_repeated_fields_by_subfield=id

# field b
field: "b"

# field a
field: "a"
message: { id: "b" }
message: { id: "a" }

# new group

# field b
field: "b"

# field a
field: "a"
message: { id: "b" }
message: { id: "a" }
`,
		out: `# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: sort_repeated_fields_by_subfield=id

# field a
field: "a"

# field b
field: "b"
message: { id: "a" }
message: { id: "b" }

# new group

# field a
field: "a"

# field b
field: "b"
message: { id: "a" }
message: { id: "b" }
`}, {
		name: "reverse sort multiple groups of repeated fields",
		in: `# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: sort_repeated_fields_by_subfield=id
# txtpbfmt: reverse_sort

# field a
field: "a"

# field b
field: "b"
message: { id: "a" }
message: { id: "b" }

# new group

# field a
field: "a"

# field b
field: "b"
message: { id: "a" }
message: { id: "b" }
`,
		out: `# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: sort_repeated_fields_by_subfield=id
# txtpbfmt: reverse_sort

# field b
field: "b"

# field a
field: "a"
message: { id: "b" }
message: { id: "a" }

# new group

# field b
field: "b"

# field a
field: "a"
message: { id: "b" }
message: { id: "a" }
`}, {
		name: "detached comment creates a new group for sorting",
		in: `# txtpbfmt: sort_repeated_fields_by_content

# field c
field: "c"

# field a
field: "a"

# new group - the fields below don't get sorted with the fields above

# field b
field: "b"
`,
		out: `# txtpbfmt: sort_repeated_fields_by_content

# field a
field: "a"

# field c
field: "c"

# new group - the fields below don't get sorted with the fields above

# field b
field: "b"
`}, {
		name: "detached comment creates a new group for reverse sorting",
		in: `# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: reverse_sort

# field a
field: "a"

# field c
field: "c"

# new group - the fields below don't get sorted with the fields above

# field b
field: "b"
`,
		out: `# txtpbfmt: sort_repeated_fields_by_content
# txtpbfmt: reverse_sort

# field c
field: "c"

# field a
field: "a"

# new group - the fields below don't get sorted with the fields above

# field b
field: "b"
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
		name: "plx dashboard mixed quotes",
		in: `# txtpbfmt: wrap_strings_after_newlines
# txtpbfmt: smartquotes
types_text_content: {
  text: "Some text\nwith a <a href=\"https://www.google.com\">hyperlink</a>\nincluded"
}
chart_spec: "{\"columnDefinitions\":[]}"
inline_script: "SELECT \'Hello\' AS hello"
`,
		out: `# txtpbfmt: wrap_strings_after_newlines
# txtpbfmt: smartquotes
types_text_content: {
  text:
    'Some text\n'
    'with a <a href="https://www.google.com">hyperlink</a>\n'
    'included'
}
chart_spec: '{"columnDefinitions":[]}'
inline_script: "SELECT 'Hello' AS hello"
`}, {
		name: "txtpbfmt off/on",
		in: `# txtpbfmt: off
  fmt:    "off" # txtpbfmt: on
foo: "bar"
bar {
      baz: "qux"
   # txtpbfmt: off
             # comment
# comment
 no_format {
  foo:   "bar"
}
      enabled: {TEMPLATE_plx}
# txtpbfmt: on


}
  should_format {
foo:  "bar"
  }


  
# txtpbfmt: off
      no_format {    foo:   "bar"  } # txtpbfmt: on
  should_format {
foo:  "bar"
  }
`,
		out: `# txtpbfmt: off
  fmt:    "off" # txtpbfmt: on
foo: "bar"
bar {
  baz: "qux"
   # txtpbfmt: off
             # comment
# comment
 no_format {
  foo:   "bar"
}
      enabled: {TEMPLATE_plx}
# txtpbfmt: on

}
should_format {
  foo: "bar"
}

# txtpbfmt: off
      no_format {    foo:   "bar"  } # txtpbfmt: on
should_format {
  foo: "bar"
}
`}, {
		name: "txtpbfmt off/on doesn't work within list",
		in: `foo: [
# txtpbfmt: off
a, b,
              c
# txtpbfmt: on
]
`,
		out: `foo: [
  # txtpbfmt: off
  a,
  b,
  c
  # txtpbfmt: on
]
`}, {
		name: "carriage return \\r is formatted away",
		in:   `foo: "bar"` + "\r" + `baz: "bat"` + "\r",
		out:  `foo: "bar"` + "\n" + `baz: "bat"` + "\n"}, {
		name: "Windows-style newline \\r\\n is formatted away",
		in:   `foo: "bar"` + "\r\n" + `baz: "bat"` + "\r\n",
		out:  `foo: "bar"` + "\n" + `baz: "bat"` + "\n"}}
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
		config  config.Config
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
		config: config.Config{ExpandAllChildren: false},
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
		config: config.Config{ExpandAllChildren: true},
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
		config: config.Config{ExpandAllChildren: false},
		out: `presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
	}, {
		name: "ExpandedOnFix",
		in: `presubmit: { check_presubmit_service: { address: "address" failure_status: WARNING options: "options" } }
`,
		config: config.Config{ExpandAllChildren: true},
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
		config: config.Config{SortFieldsByFieldName: true},
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
		config: config.Config{SortRepeatedFieldsByContent: true},
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
		name: "SortNamedFieldBySubfieldContents",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    # Should go after ADD
    operation: {
      name: EDIT
    }
    operation: {
      name: ADD
    }
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should remain below
  auto_reviewers: "reviewerA"
}
`,
		config: config.Config{SortRepeatedFieldsBySubfield: []string{"operation.name"}},
		out: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    operation: {
      name: ADD
    }
    # Should go after ADD
    operation: {
      name: EDIT
    }
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should remain below
  auto_reviewers: "reviewerA"
}
`,
	}, {
		name: "SortNamedFieldByMultipleSubfieldContents",
		in: `presubmit: {
  operation {
    name: EDIT
  }
  operation {
    name: ADD
  }
  test {
    id: 4
  }
  test {
    id: 2
  }
}
`,
		config: config.Config{SortRepeatedFieldsBySubfield: []string{"operation.name", "test.id"}},
		out: `presubmit: {
  operation {
    name: ADD
  }
  operation {
    name: EDIT
  }
  test {
    id: 2
  }
  test {
    id: 4
  }
}
`,
	}, {
		name: "SortAnyFieldBySubfieldContents",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    # Should go after ADD
    operation: {
      name: EDIT
    }
    operation: {
      name: ADD
    }
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should remain below
  auto_reviewers: "reviewerA"
}
`,
		config: config.Config{SortRepeatedFieldsBySubfield: []string{"name"}},
		out: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    operation: {
      name: ADD
    }
    # Should go after ADD
    operation: {
      name: EDIT
    }
    prohibited_regexp: "UnsafeFunction"
    check_delta_only: true
  }
  # Should remain below
  auto_reviewers: "reviewerA"
}
`,
	}, {
		name: "SortBySubfieldsDontSortFieldsWithDifferentNames",
		in: `presubmit: {
  check_contents: {
    operation1: {
      name: EDIT
    }
    operation2: {
      name: ADD
    }
  }
}
`,
		config: config.Config{SortRepeatedFieldsBySubfield: []string{"name"}},
		out: `presubmit: {
  check_contents: {
    operation1: {
      name: EDIT
    }
    operation2: {
      name: ADD
    }
  }
}
`,
	}, {
		name: "SortSeparatedFieldBySubfieldContents",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    # Should go after ADD
    operation: {
      name: EDIT
    }
    split: 1
    operation: {
      name: ADD
    }
  }
  # Should move above
  auto_reviewers: "reviewerA"
}
`,
		config: config.Config{SortFieldsByFieldName: true, SortRepeatedFieldsBySubfield: []string{"name"}},
		out: `presubmit: {
  auto_reviewers: "reviewerB"
  # Should move above
  auto_reviewers: "reviewerA"
  check_contents: {
    operation: {
      name: ADD
    }
    # Should go after ADD
    operation: {
      name: EDIT
    }
    split: 1
  }
}
`,
	}, {
		name: "SortSubfieldsIgnoreEmptySubfieldName",
		in: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    operation: {
      name: EDIT
    }
    operation: {
      name: ADD
    }
  }
  auto_reviewers: "reviewerA"
}
`,
		config: config.Config{SortRepeatedFieldsBySubfield: []string{"operation."}},
		out: `presubmit: {
  auto_reviewers: "reviewerB"
  check_contents: {
    operation: {
      name: EDIT
    }
    operation: {
      name: ADD
    }
  }
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
		config: config.Config{SortFieldsByFieldName: true, SortRepeatedFieldsByContent: true},
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
		in: `
	below_wrapper: true
	wrapper: {
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
		config: config.Config{
			FieldSortOrder: map[string][]string{
				config.RootName:  {"wrapper", "below_wrapper"},
				"check_contents": {"z_first", "x_second", "x_third"},
			},
		},
		// Nodes are sorted by the specified order, else left untouched.
		out: `wrapper: {
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
below_wrapper: true
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
		config: config.Config{
			FieldSortOrder: map[string][]string{
				"check_contents": {"z_first", "x_second", "x_third", "not_required"},
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
		config: config.Config{
			FieldSortOrder: map[string][]string{
				"check_contents": {"z_first", "x_second", "x_third"},
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
		config: config.Config{RemoveDuplicateValuesForRepeatedFields: true},
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
		config: config.Config{
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
		config: config.Config{
			AllowTripleQuotedStrings: true,
		},
		in: `foo: """bar"""
`,
		out: `foo: """bar"""
`,
	}, {
		name: "TripleQuotedStrings_multiLine",
		config: config.Config{
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
		config: config.Config{
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
		config: config.Config{
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
		name: "WrapStringsAtColumn",
		config: config.Config{
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
		name: "WrapStringsAtColumn_inlineChildren",
		config: config.Config{
			WrapStringsAtColumn: 14,
		},
		in: `root {
  # inline children don't wrap
  inner { inline: "89 1234" }
  # Verify that skipping an inline string doesn't skip the rest of the file.
  wrappable {
    s: "will wrap"
  }
}
`,
		out: `root {
  # inline children don't wrap
  inner { inline: "89 1234" }
  # Verify that skipping an inline string doesn't skip the rest of the file.
  wrappable {
    s:
      "will "
      "wrap"
  }
}
`,
	}, {
		name: "WrapStringsAtColumn_exactlyNumColumnsDoesNotWrap",
		config: config.Config{
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
		name: "WrapStringsAtColumn_numColumnsPlus1Wraps",
		config: config.Config{
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
		name: "WrapStringsAtColumn_commentKeptWhenLinesReduced",
		config: config.Config{
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
		name: "WrapStringsAtColumn_doNotBreakLongWords",
		config: config.Config{
			WrapStringsAtColumn: 15,
		},
		in: `s: "one@two_three-four&five"
`,
		out: `s: "one@two_three-four&five"
`,
	}, {
		name: "WrapStringsAtColumn_wrapHtml",
		config: config.Config{
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
		name: "WrapStringsAtColumn_empty",
		config: config.Config{
			WrapStringsAtColumn: 15,
		},
		in: `s: 
`,
		out: `s: 
`,
	}, {
		name: "WrapStringsAtColumn_doNoWrapHtmlByDefault",
		config: config.Config{
			WrapStringsAtColumn: 15,
		},
		in: `s: "one two three <four>"
`,
		out: `s: "one two three <four>"
`,
	}, {
		name: "WrapStringsAtColumn_doNoWrapHtmlRealistic",
		config: config.Config{
			WrapStringsAtColumn: 40,
		},
		in: `text:
  "The two lines below should not wrap since they contains HTML tags "
  "<a href=\"https://support.example.com/project/a/pageone/pagetwo/pagethree?hl=1234\">"
  "some' text</a>."`,
		out: `text:
  "The two lines below should not wrap since they contains HTML tags "
  "<a href=\"https://support.example.com/project/a/pageone/pagetwo/pagethree?hl=1234\">"
  "some' text</a>."
`,
	}, {
		name: "WrapStringsAtColumn_metaComment",
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
		name: "WrapStringsAtColumn_doNotWrapNonStrings",
		config: config.Config{
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
		name: "WrapStringsAtColumn_alreadyWrappedStringsAreNotRewrapped",
		config: config.Config{
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
		name: "WrapStringsAtColumn_alreadyWrappedStringsAreNotRewrappedUnlessSomeAreLonger",
		config: config.Config{
			WrapStringsAtColumn: 15,
		},
		in: `s:
  "I "
  "am already "
  "wrapped"
  " but I am not!"
`,
		out: `s:
  "I am "
  "already "
  "wrapped "
  "but I am "
  "not!"
`,
	}, {
		name: "WrapStringsAtColumn_tripleQuotedStringsAreNotWrapped",
		config: config.Config{
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
		name: "WrapStringsAfterNewlines",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `# Comments are not \n wrapped
s: "one two \nthree four\nfive"
`,
		out: `# Comments are not \n wrapped
s:
  "one two \n"
  "three four\n"
  "five"
`,
	}, {
		name: "WrapStringsAfterNewlines_motivatingExampleWithMarkup",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `root {
  doc: "<body>\n  <p>\n    Hello\n  </p>\n</body>\n"
}
`,
		out: `root {
  doc:
    "<body>\n"
    "  <p>\n"
    "    Hello\n"
    "  </p>\n"
    "</body>\n"
}
`,
	}, {
		name: "WrapStringsAfterNewlines_inlineChildren",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `root {
  # inline children don't wrap
  inner { inline: "89 1234" }
  # Verify that skipping an inline string doesn't skip the rest of the file.
  wrappable {
    s: "will \nwrap"
  }
}
`,
		out: `root {
  # inline children don't wrap
  inner { inline: "89 1234" }
  # Verify that skipping an inline string doesn't skip the rest of the file.
  wrappable {
    s:
      "will \n"
      "wrap"
  }
}
`,
	}, {
		name: "WrapStringsAfterNewlines_noNewlineDoesNotWrap",
		config: config.Config{
			WrapStringsAfterNewlines: true,
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
		name: "WrapStringsAfterNewlines_trailingNewlineDoesNotWrap",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `root {
  s: "89 123\n"
}
`,
		out: `root {
  s: "89 123\n"
}
`,
	}, {
		name: "WrapStringsAfterNewlines_trailingNewlineDoesNotLeaveSuperfluousEmptyString",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `root {
  s: "89\n123\n"
}
`,
		out: `root {
  s:
    "89\n"
    "123\n"
}
`,
	}, {
		name: "WrapStringsAfterNewlines_empty",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `s: 
`,
		out: `s: 
`,
	}, {
		name: "WrapStringsAfterNewlines_metaComment",
		in: `# txtpbfmt: wrap_strings_after_newlines
s: "one two \nthree\n four"
`,
		out: `# txtpbfmt: wrap_strings_after_newlines
s:
  "one two \n"
  "three\n"
  " four"
`,
	}, {
		name: "WrapStringsAfterNewlines_alreadyWrappedStringsAreRewrapped",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `s:
  "I "
  "am already\n"
  "wrapped. \nBut this was not."
`,
		out: `s:
  "I am already\n"
  "wrapped. \n"
  "But this was not."
`,
	}, {
		name: "WrapStringsAfterNewlines_tripleQuotedStringsAreNotWrapped",
		config: config.Config{
			WrapStringsAfterNewlines: true,
			AllowTripleQuotedStrings: true,
		},
		in: `s1: """one two three four five"""
s2: '''six seven \neight nine'''
`,
		out: `s1: """one two three four five"""
s2: '''six seven \neight nine'''
`,
	}, {
		name: "WrapStringsAfterNewlines_tooManyEscapesDoesNotWrap",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `s: "7\nsev\xADen\x00"
`,
		out: `s: "7\nsev\xADen\x00"
`,
	}, {
		name: "WrapStringsAfterNewlines_wayTooManyEscapesDoesNotWrap",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `s: "ﾭ\xde\x00\x00\x00\x08\n(\x02\n\x0b\x00\x07\x01h\x0c\x14\x01"
`,
		out: `s: "ﾭ\xde\x00\x00\x00\x08\n(\x02\n\x0b\x00\x07\x01h\x0c\x14\x01"
`,
	}, {
		name: "WrapStringsAfterNewlines_aFewEscapesStillWrap",
		config: config.Config{
			WrapStringsAfterNewlines: true,
		},
		in: `s: "aaaaaaaaaa \n bbbbbbbbbb \n cccccccccc \n dddddddddd \n eeeeeeeeee\x00 \n"
`,
		out: `s:
  "aaaaaaaaaa \n"
  " bbbbbbbbbb \n"
  " cccccccccc \n"
  " dddddddddd \n"
  " eeeeeeeeee\x00 \n"
`,
	}, {
		name: "WrapStringsAtColumn_noWordwrap",
		config: config.Config{
			WrapStringsAtColumn:        12,
			WrapStringsWithoutWordwrap: true,
		},
		in: `# 3456789012
s: "Curabitur\040elit\x20nec mi egestas,\u000Dtincidunt \U00010309nterdum elit porta.\n"
`,
		out: `# 3456789012
s:
  "Curabitu"
  "r\040eli"
  "t\x20nec"
  " mi eges"
  "tas,"
  "\u000Dti"
  "ncidunt "
  "\U00010309"
  "nterdum "
  "elit por"
  "ta.\n"
`,
	}, {
		name: "WrapStringsAtColumn_noWordwrapDeep",
		config: config.Config{
			WrapStringsAtColumn:        12,
			WrapStringsWithoutWordwrap: true,
		},
		in: `
this_field_name_displays_wider_than_the_twelve_requested:  "this_goes_to_a_new_line"
`,
		out: `this_field_name_displays_wider_than_the_twelve_requested:
  "this_goe"
  "s_to_a_n"
  "ew_line"
`,
	}, {
		name: "WrapStringsAtColumn_noWordwrapDeepInlinePromotion",
		config: config.Config{
			WrapStringsAtColumn:        12,
			WrapStringsWithoutWordwrap: true,
		},
		in: `
this_field_name_displays_wider_than_the_twelve_requested: "0C" # XII
`,
		out: `this_field_name_displays_wider_than_the_twelve_requested:
  # XII
  "0C"
`,
	}, {
		name: "WrapStringsAtColumn_noWordwrapMetacomment",
		in: `# txtpbfmt: wrap_strings_at_column=12
# txtpbfmt: wrap_strings_without_wordwrap
# 3456789012
s: "1\tone\r\n2\ttwo\r\n3\tthree\r\n4\tfour\r\n"
`,
		out: `# txtpbfmt: wrap_strings_at_column=12
# txtpbfmt: wrap_strings_without_wordwrap
# 3456789012
s:
  "1\tone\r"
  "\n2\ttwo"
  "\r\n3\tt"
  "hree\r\n"
  "4\tfour"
  "\r\n"
`,
	}, {
		name: "PreserveAngleBrackets",
		config: config.Config{
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
		config: config.Config{
			SmartQuotes: false,
		},
		in: `foo: "\"bar\""`,
		out: `foo: "\"bar\""
`,
	}, {
		name: "smart quotes",
		config: config.Config{
			SmartQuotes: true,
		},
		in: `foo: "\"bar\""`,
		out: `foo: '"bar"'
`,
	}, {
		name: "smart quotes via meta comment",
		config: config.Config{
			SmartQuotes: false,
		},
		in: `# txtpbfmt: smartquotes
foo: "\"bar\""`,
		out: `# txtpbfmt: smartquotes
foo: '"bar"'
`,
	}, {
		name: "carriage returns",
		in:   "a{\r\n}\r\n",
		out:  "a {\n}\n",
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
		gotRaw, err := FormatWithConfig([]byte(in), config.Config{SmartQuotes: false})
		got := strings.TrimSpace(string(gotRaw))
		if err != nil {
			t.Errorf("%s: got error: %s, want no error and [ %s ]", name, err, want)
		} else if got != want {
			t.Errorf("%s: got [ %s ], want [ %s ]", name, got, want)
		}

		name = fmt.Sprintf("Format [ %s ] with smart quote behavior", tc.in)
		want = `foo: ` + tc.wantSmart
		gotRaw, err = FormatWithConfig([]byte(`foo: `+tc.in), config.Config{SmartQuotes: true})
		got = strings.TrimSpace(string(gotRaw))
		if err != nil {
			t.Errorf("%s: got error: %s, want no error and [ %s ]", name, err, want)
		} else if got != want {
			t.Errorf("%s: got [ %s ], want [ %s ]", name, got, want)
		}
	}
}

func FuzzParse(f *testing.F) {
	testcases := []string{"", "a: 123", "input { dimension: [2, 4, 6, 8] }", "]", "\":%\"",
		"%''#''0'0''0''0''0''0\""}
	for _, tc := range testcases {
		f.Add([]byte(tc))
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		Parse(in)
	})
}
