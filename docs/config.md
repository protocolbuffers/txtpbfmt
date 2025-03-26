# Text Proto Formatter - Configuration

<!--* freshness: { exempt: true } *-->

`txtpbfmt` provides several configuration options that customize the specifics
of the output format. These are configured by adding a comment line to top of
the proto file (before the first non-empty non-comment line) of the form:

`# txtpbfmt: [config-option]`

This doc describes each of these options.

## AllowTripleQuotedStrings
`# txtpbfmt: allow_triple_quoted_strings`

Permit usage of Python-style `"""` or `'''` delimited strings.

## AllowUnnamedNodesEverywhere
`# txtpbfmt: allow_unnamed_nodes_everywhere`

Allow unnamed nodes everywhere.
Default is to allow only top-level nodes to be unnamed.

## ExpandAllChildren
`# txtpbfmt: expand_all_children`

Expand all children irrespective of the initial state.

### Before formatting

[Example](examples/expand_all_children.IN.textproto)

### After formatting

[Example](examples/expand_all_children.OUT.textproto)

## PreserveAngleBrackets

`# txtpbfmt: preserve_angle_brackets`

Whether angle brackets used instead of curly braces should be preserved
when outputting a formatted textproto.

### Before formatting

[Example](examples/preserve_angle_brackets.IN.textproto)

### After formatting

[Example](examples/preserve_angle_brackets.OUT.textproto)

## RemoveDuplicateValuesForRepeatedFields
`# txtpbfmt: remove_duplicate_values_for_repeated_fields`

Remove lines that have the same field name and scalar value as another.

### Before formatting

[Example](examples/remove_duplicate_values_for_repeated_fields.IN.textproto)

### After formatting

[Example](examples/remove_duplicate_values_for_repeated_fields.OUT.textproto)

## SkipAllColons
`# txtpbfmt: skip_all_colons`

Skip colons whenever possible.

### Before formatting

[Example](examples/skip_all_colons.IN.textproto)

### After formatting

[Example](examples/skip_all_colons.OUT.textproto)

## SmartQuotes

`# txtpbfmt: smartquotes`

Use single quotes around strings that contain double but not single quotes.

### Before formatting

[Example](examples/smartquotes.IN.textproto)

### After formatting

[Example](examples/smartquotes.OUT.textproto)

## SortFieldsByFieldName
`# txtpbfmt: sort_fields_by_field_name`

Sort fields by field name.

### Before formatting

[Example](examples/sort_fields_by_field_name.IN.textproto)

### After formatting

[Example](examples/sort_fields_by_field_name.OUT.textproto)

## SortRepeatedFieldsByContent
`# txtpbfmt: sort_repeated_fields_by_content`

Sort adjacent scalar fields of the same field name by their contents.

### Before formatting

[Example](examples/sort_repeated_fields_by_content.IN.textproto)

### After formatting

[Example](examples/sort_repeated_fields_by_content.OUT.textproto)

## SortRepeatedFieldsBySubfield
`# txtpbfmt: sort_repeated_fields_by_subfield=[subfieldSpec]`

Sort adjacent message fields of the given field name by the contents of the
given subfield path.

`subfieldSpec` is of one of two forms:

*   `fieldName.subfieldName1.subfieldName2...subfieldNameN`, which will sort
    fields named `fieldName` by the value of the final subfield named
    `subfieldNameN`.

*   `subfieldName`, which will sort any field by its subfield named
    `subfieldName`

### Before formatting

[Example](examples/sort_repeated_fields_by_subfield.IN.textproto)

### After formatting

[Example](examples/sort_repeated_fields_by_subfield.OUT.textproto)

## ReverseSort

`# txtpbfmt: reverse_sort`

Sorts all `sort_*` fields in descending order instead of the default ascending
order. Does nothing if not used with at least 1 other `sort_*` field.

### Before formatting

[Example](examples/reverse_sort.IN.textproto)

### After formatting

[Example](examples/reverse_sort.OUT.textproto)

## WrapHTMLStrings
`# txtpbfmt: wrap_html_strings`

Whether strings that appear to contain HTML tags should be wrapped
(requires WrapStringsAtColumn to be set).

### Before formatting

[Example](examples/wrap_html_strings.IN.textproto)

### After formatting

[Example](examples/wrap_html_strings.OUT.textproto)

## WrapStringsAtColumn
`# txtpbfmt: wrap_strings_at_column=[column]`

Max columns for string field values. If zero, no string wrapping will occur.
Strings that may contain HTML tags will not be wrapped unless
`wrap_html_strings` is also specified.

### Before formatting

[Example](examples/wrap_strings_at_column.IN.textproto)

### After formatting

[Example](examples/wrap_strings_at_column.OUT.textproto)
