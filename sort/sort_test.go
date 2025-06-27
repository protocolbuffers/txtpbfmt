package sort

import (
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/protocolbuffers/txtpbfmt/ast"
	"github.com/protocolbuffers/txtpbfmt/config"
)

func TestParseSubfieldSpec(t *testing.T) {
	tests := []struct {
		name             string
		subfieldSpec     string
		wantField        string
		wantSubfieldPath []string
	}{{
		name:             "empty",
		subfieldSpec:     "",
		wantField:        "",
		wantSubfieldPath: []string{""},
	}, {
		name:             "no dot",
		subfieldSpec:     "subfield",
		wantField:        "",
		wantSubfieldPath: []string{"subfield"},
	}, {
		name:             "one dot",
		subfieldSpec:     "field.subfield",
		wantField:        "field",
		wantSubfieldPath: []string{"subfield"},
	}, {
		name:             "multiple dots",
		subfieldSpec:     "field.subfield1.subfield2",
		wantField:        "field",
		wantSubfieldPath: []string{"subfield1", "subfield2"},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotField, gotSubfieldPath := parseSubfieldSpec(tc.subfieldSpec)
			if gotField != tc.wantField {
				t.Errorf("parseSubfieldSpec(%q) got field %q, want %q", tc.subfieldSpec, gotField, tc.wantField)
			}
			if diff := cmp.Diff(tc.wantSubfieldPath, gotSubfieldPath); diff != "" {
				t.Errorf("parseSubfieldSpec(%q) returned diff (-want +got):\n%s", tc.subfieldSpec, diff)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name               string
		nodes              []*ast.Node
		c                  config.Config
		want               []*ast.Node
		wantErr            string
		skipSortFunction   bool
		skipFilterFunction bool
		skipValuesFunction bool
	}{{
		name:  "empty nodes",
		nodes: []*ast.Node{},
		c:     config.Config{},
		want:  []*ast.Node{},
	}, {
		name: "filter function",
		nodes: []*ast.Node{
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
			},
			{
				Name: "b",
				Values: []*ast.Value{
					{Value: "2"},
				},
			},
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
			},
		},
		c: config.Config{
			RemoveDuplicateValuesForRepeatedFields: true,
		},
		want: []*ast.Node{
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
			},
			{
				Name: "b",
				Values: []*ast.Value{
					{Value: "2"},
				},
			},
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
				Deleted: true,
			},
		},
	}, {
		name: "sort function",
		nodes: []*ast.Node{
			{Name: "b"},
			{Name: "a"},
		},
		c: config.Config{
			SortFieldsByFieldName: true,
		},
		want: []*ast.Node{
			{Name: "a"},
			{Name: "b"},
		},
	}, {
		name: "values function",
		nodes: []*ast.Node{
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "2"},
					{Value: "1"},
				},
				ValuesAsList: true,
			},
		},
		c: config.Config{
			SortRepeatedFieldsByContent: true,
		},
		want: []*ast.Node{
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
					{Value: "2"},
				},
				ValuesAsList: true,
			},
		},
	}, {
		name: "error in sort function",
		nodes: []*ast.Node{
			{Name: "a"},
			{Name: "b"},
		},
		c: config.Config{
			SortFieldsByFieldName: true,
			FieldSortOrder: map[string][]string{
				config.RootName: []string{},
			},
			RequireFieldSortOrderToMatchAllFieldsInNode: true,
		},
		want: []*ast.Node{
			{Name: "a"},
			{Name: "b"},
		},
		wantErr: "fields parsed that were not specified",
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sortFunction := nodeSortFunction(tc.c)
			if tc.skipSortFunction {
				sortFunction = nil
			}
			filterFunction := nodeFilterFunction(tc.c)
			if tc.skipFilterFunction {
				filterFunction = nil
			}
			valuesFunction := valuesSortFunction(tc.c)
			if tc.skipValuesFunction {
				valuesFunction = nil
			}
			err := process(nil, tc.nodes, sortFunction, filterFunction, valuesFunction)
			if tc.wantErr != "" {
				if err == nil {
					t.Errorf("process(%v, %v, %v, %v) got nil error, want %q", tc.nodes, sortFunction, filterFunction, valuesFunction, tc.wantErr)
				} else if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("process(%v, %v, %v, %v) got error %q, want %q", tc.nodes, sortFunction, filterFunction, valuesFunction, err.Error(), tc.wantErr)
				}
			} else if err != nil {
				t.Errorf("process(%v, %v, %v, %v) got error %v, want nil", tc.nodes, sortFunction, filterFunction, valuesFunction, err)
			}

			if diff := cmp.Diff(tc.want, tc.nodes); diff != "" {
				t.Errorf("process(%v, %v, %v, %v) returned diff (-want +got):\n%s", tc.nodes, sortFunction, filterFunction, valuesFunction, diff)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name  string
		nodes []*ast.Node
		want  []*ast.Node
	}{{
		name:  "empty nodes",
		nodes: []*ast.Node{},
		want:  []*ast.Node{},
	}, {
		name: "no duplicates",
		nodes: []*ast.Node{
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
			},
			{
				Name: "b",
				Values: []*ast.Value{
					{Value: "2"},
				},
			},
		},
		want: []*ast.Node{
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
			},
			{
				Name: "b",
				Values: []*ast.Value{
					{Value: "2"},
				},
			},
		},
	}, {
		name: "duplicates",
		nodes: []*ast.Node{
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
			},
			{
				Name: "b",
				Values: []*ast.Value{
					{Value: "2"},
				},
			},
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
			},
		},
		want: []*ast.Node{
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
			},
			{
				Name: "b",
				Values: []*ast.Value{
					{Value: "2"},
				},
			},
			{
				Name: "a",
				Values: []*ast.Value{
					{Value: "1"},
				},
				Deleted: true,
			},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			RemoveDuplicates(tc.nodes)
			if diff := cmp.Diff(tc.want, tc.nodes); diff != "" {
				t.Errorf("RemoveDuplicates(%v) returned diff (-want +got):\n%s", tc.nodes, diff)
			}
		})
	}
}

func TestNodeSortFunction(t *testing.T) {
	tests := []struct {
		name string
		c    config.Config
		want bool
	}{{
		name: "empty config",
		c:    config.Config{},
		want: false,
	}, {
		name: "SortFieldsByFieldName",
		c:    config.Config{SortFieldsByFieldName: true},
		want: true,
	}, {
		name: "SortRepeatedFieldsByContent",
		c:    config.Config{SortRepeatedFieldsByContent: true},
		want: true,
	}, {
		name: "SortRepeatedFieldsBySubfield",
		c:    config.Config{SortRepeatedFieldsBySubfield: []string{"field.subfield"}},
		want: true,
	}, {
		name: "FieldSortOrder",
		c: config.Config{
			FieldSortOrder: map[string][]string{
				"parent": []string{"child"},
			},
		},
		want: true,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := nodeSortFunction(tc.c)
			if (got != nil) != tc.want {
				t.Errorf("nodeSortFunction(%v) got %v, want %v", tc.c, got, tc.want)
			}
		})
	}
}

func TestNodeFilterFunction(t *testing.T) {
	tests := []struct {
		name string
		c    config.Config
		want bool
	}{{
		name: "empty config",
		c:    config.Config{},
		want: false,
	}, {
		name: "RemoveDuplicateValuesForRepeatedFields",
		c:    config.Config{RemoveDuplicateValuesForRepeatedFields: true},
		want: true,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := nodeFilterFunction(tc.c)
			if (got != nil) != tc.want {
				t.Errorf("nodeFilterFunction(%v) got %v, want %v", tc.c, got, tc.want)
			}
		})
	}
}

func TestValuesSortFunction(t *testing.T) {
	tests := []struct {
		name string
		c    config.Config
		want bool
	}{{
		name: "empty config",
		c:    config.Config{},
		want: false,
	}, {
		name: "SortRepeatedFieldsByContent",
		c:    config.Config{SortRepeatedFieldsByContent: true},
		want: true,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := valuesSortFunction(tc.c)
			if (got != nil) != tc.want {
				t.Errorf("valuesSortFunction(%v) got %v, want %v", tc.c, got, tc.want)
			}
		})
	}
}

func TestGetNodePriorityForByFieldOrder(t *testing.T) {
	tests := []struct {
		name              string
		parent            *ast.Node
		node              *ast.Node
		nodeName          string
		priorities        map[string]int
		unsortedCollector UnsortedFieldCollectorFunc
		want              int
	}{{
		name: "parent name mismatch",
		parent: &ast.Node{
			Name: "parent",
		},
		node: &ast.Node{
			Name: "child",
		},
		nodeName:          "other",
		priorities:        map[string]int{"child": 1},
		unsortedCollector: func(string, int32, string) {},
		want:              math.MaxInt,
	}, {
		name:   "root name mismatch",
		parent: nil,
		node: &ast.Node{
			Name: "child",
		},
		nodeName:          "other",
		priorities:        map[string]int{"child": 1},
		unsortedCollector: func(string, int32, string) {},
		want:              math.MaxInt,
	}, {
		name: "comment only",
		parent: &ast.Node{
			Name: "parent",
		},
		node: &ast.Node{
			Name: "",
		},
		nodeName:          "parent",
		priorities:        map[string]int{"child": 1},
		unsortedCollector: func(string, int32, string) {},
		want:              math.MaxInt,
	}, {
		name: "not in priorities",
		parent: &ast.Node{
			Name: "parent",
		},
		node: &ast.Node{
			Name: "child",
		},
		nodeName:          "parent",
		priorities:        map[string]int{"other": 1},
		unsortedCollector: func(string, int32, string) {},
		want:              0,
	}, {
		name: "in priorities",
		parent: &ast.Node{
			Name: "parent",
		},
		node: &ast.Node{
			Name: "child",
		},
		nodeName:          "parent",
		priorities:        map[string]int{"child": 1},
		unsortedCollector: func(string, int32, string) {},
		want:              1,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getNodePriorityForByFieldOrder(tc.parent, tc.node, tc.nodeName, tc.priorities, tc.unsortedCollector)
			if got == nil {
				if tc.want != math.MaxInt {
					t.Errorf("getNodePriorityForByFieldOrder(%v, %v, %q, %v, %v) got nil, want %v", tc.parent, tc.node, tc.nodeName, tc.priorities, tc.unsortedCollector, tc.want)
				}
				return
			}
			if diff := cmp.Diff(tc.want, *got); diff != "" {
				t.Errorf("getNodePriorityForByFieldOrder(%v, %v, %q, %v, %v) returned diff (-want +got):\n%s", tc.parent, tc.node, tc.nodeName, tc.priorities, tc.unsortedCollector, diff)
			}
		})
	}
}
