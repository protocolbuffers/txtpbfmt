// Package sort provides functions for sorting nodes and values.
package sort

import (
	"fmt"
	"math"
	"strings"

	"github.com/protocolbuffers/txtpbfmt/ast"
	"github.com/protocolbuffers/txtpbfmt/config"
)

// UnsortedFieldsError will be returned by ParseWithConfig if
// Config.RequireFieldSortOrderToMatchAllFieldsInNode is set, and an unrecognized field is found
// while parsing.
type UnsortedFieldsError struct {
	UnsortedFields []UnsortedField
}

// UnsortedField records details about a single unsorted field.
type UnsortedField struct {
	FieldName       string
	Line            int32
	ParentFieldName string
}

func (e *UnsortedFieldsError) Error() string {
	var errs []string
	for _, us := range e.UnsortedFields {
		errs = append(errs, fmt.Sprintf("  line: %d, parent field: %q, unsorted field: %q", us.Line, us.ParentFieldName, us.FieldName))
	}
	return fmt.Sprintf("fields parsed that were not specified in the parser.AddFieldSortOrder() call:\n%s", strings.Join(errs, "\n"))
}

// NodeSortFunction sorts the given nodes, using the parent node as context. parent can be nil.
type NodeSortFunction func(parent *ast.Node, nodes []*ast.Node) error

// NodeFilterFunction filters the given nodes.
type NodeFilterFunction func(nodes []*ast.Node)

// ValuesSortFunction sorts the given values.
type ValuesSortFunction func(values []*ast.Value)

// Process sorts and filters the given nodes.
func Process(parent *ast.Node, nodes []*ast.Node, c config.Config) error {
	return process(parent, nodes, nodeSortFunction(c), nodeFilterFunction(c), valuesSortFunction(c))
}

// process sorts and filters the given nodes.
func process(parent *ast.Node, nodes []*ast.Node, sortFunction NodeSortFunction, filterFunction NodeFilterFunction, valuesSortFunction ValuesSortFunction) error {
	if len(nodes) == 0 {
		return nil
	}
	if filterFunction != nil {
		filterFunction(nodes)
	}
	for _, nd := range nodes {
		err := process(nd, nd.Children, sortFunction, filterFunction, valuesSortFunction)
		if err != nil {
			return err
		}
		if valuesSortFunction != nil && nd.ValuesAsList {
			valuesSortFunction(nd.Values)
		}
	}
	if sortFunction != nil {
		return sortFunction(parent, nodes)
	}
	return nil
}

// RemoveDuplicates marks duplicate key:value pairs from nodes as Deleted.
func RemoveDuplicates(nodes []*ast.Node) {
	type nameAndValue struct {
		name, value string
	}
	seen := make(map[nameAndValue]bool)
	for _, nd := range nodes {
		if len(nd.Values) == 1 {
			key := nameAndValue{nd.Name, nd.Values[0].Value}
			if _, value := seen[key]; value {
				// Name-Value pair found in the same nesting level, deleting.
				nd.Deleted = true
			} else {
				seen[key] = true
			}
		}
	}
}

// UnsortedFieldCollector collects UnsortedFields during parsing.
type UnsortedFieldCollector struct {
	fields map[string]UnsortedField
}

// newUnsortedFieldCollector returns a new UnsortedFieldCollector.
func newUnsortedFieldCollector() *UnsortedFieldCollector {
	return &UnsortedFieldCollector{
		fields: make(map[string]UnsortedField),
	}
}

// UnsortedFieldCollectorFunc collects UnsortedFields during parsing.
type UnsortedFieldCollectorFunc func(name string, line int32, parent string)

// collect collects the unsorted field.
func (ufc *UnsortedFieldCollector) collect(name string, line int32, parent string) {
	ufc.fields[name] = UnsortedField{name, line, parent}
}

// asError returns an error if any unsorted fields were collected.
func (ufc *UnsortedFieldCollector) asError() error {
	if len(ufc.fields) == 0 {
		return nil
	}
	var fields []UnsortedField
	for _, f := range ufc.fields {
		fields = append(fields, f)
	}
	return &UnsortedFieldsError{fields}
}

// nodeSortFunction returns a function that sorts nodes based on the config.
func nodeSortFunction(c config.Config) NodeSortFunction {
	var sorter ast.NodeLess = nil
	unsortedFieldCollector := newUnsortedFieldCollector()
	for name, fieldOrder := range c.FieldSortOrder {
		sorter = ast.ChainNodeLess(sorter, ByFieldOrder(name, fieldOrder, unsortedFieldCollector.collect))
	}
	if c.SortFieldsByFieldName {
		sorter = ast.ChainNodeLess(sorter, ast.ByFieldName)
	}
	if c.SortRepeatedFieldsByContent {
		sorter = ast.ChainNodeLess(sorter, ast.ByFieldValue)
	}
	for _, sf := range c.SortRepeatedFieldsBySubfield {
		field, subfieldPath := parseSubfieldSpec(sf)
		if len(subfieldPath) > 0 {
			sorter = ast.ChainNodeLess(sorter, ast.ByFieldSubfieldPath(field, subfieldPath))
		}
	}
	if sorter != nil {
		return func(parent *ast.Node, ns []*ast.Node) error {
			ast.SortNodes(parent, ns, sorter, ast.ReverseOrdering(c.ReverseSort))
			if c.RequireFieldSortOrderToMatchAllFieldsInNode {
				return unsortedFieldCollector.asError()
			}
			return nil
		}
	}
	return nil
}

// Returns the field and subfield path parts of spec "{field}.{subfield1}.{subfield2}...".
// Spec without a dot is considered to be "{subfield}".
func parseSubfieldSpec(subfieldSpec string) (field string, subfieldPath []string) {
	parts := strings.Split(subfieldSpec, ".")
	if len(parts) == 1 {
		return "", parts
	}
	return parts[0], parts[1:]
}

// nodeFilterFunction returns a function that filters nodes based on the config.
func nodeFilterFunction(c config.Config) NodeFilterFunction {
	if c.RemoveDuplicateValuesForRepeatedFields {
		return RemoveDuplicates
	}
	return nil
}

// valuesSortFunction returns a function that sorts values based on the config.
func valuesSortFunction(c config.Config) ValuesSortFunction {
	if c.SortRepeatedFieldsByContent {
		if c.ReverseSort {
			return ast.SortValuesReverse
		}
		return ast.SortValues
	}
	return nil
}

func getNodePriorityForByFieldOrder(parent, node *ast.Node, name string, priorities map[string]int, unsortedCollector UnsortedFieldCollectorFunc) *int {
	if parent != nil && parent.Name != name {
		return nil
	}
	if parent == nil && name != config.RootName {
		return nil
	}
	// CommentOnly nodes don't set priority below, and default to MaxInt, which keeps them at the bottom
	prio := math.MaxInt

	// Unknown fields will get the int nil value of 0 from the order map, and bubble to the top.
	if !node.IsCommentOnly() {
		var ok bool
		prio, ok = priorities[node.Name]
		if !ok {
			parentName := config.RootName
			if parent != nil {
				parentName = parent.Name
			}
			unsortedCollector(node.Name, node.Start.Line, parentName)
		}
	}
	return &prio
}

// ByFieldOrder returns a NodeLess function that orders fields within a node named name
// by the order specified in fieldOrder. Nodes sorted but not specified by the field order
// are bubbled to the top and reported to unsortedCollector.
func ByFieldOrder(name string, fieldOrder []string, unsortedCollector UnsortedFieldCollectorFunc) ast.NodeLess {
	priorities := make(map[string]int)
	for i, fieldName := range fieldOrder {
		priorities[fieldName] = i + 1
	}
	return func(parent, ni, nj *ast.Node, isWholeSlice bool) bool {
		if !isWholeSlice {
			return false
		}
		vi := getNodePriorityForByFieldOrder(parent, ni, name, priorities, unsortedCollector)
		vj := getNodePriorityForByFieldOrder(parent, nj, name, priorities, unsortedCollector)
		if vi == nil {
			return vj != nil
		}
		if vj == nil {
			return false
		}
		return *vi < *vj
	}
}
