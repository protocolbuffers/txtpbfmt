package config

import (
	"reflect"
	"testing"
)

type mockLogger struct {
	infofCalls []string
}

func (m *mockLogger) Infof(format string, args ...any) {
	m.infofCalls = append(m.infofCalls, format)
}

func TestConfigInfof(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		format string
		args   []any
		want   []string
	}{{
		name:   "LoggerIsNil",
		config: Config{},
		format: "test message",
		want:   nil,
	}, {
		name:   "LoggerNotNil",
		config: Config{Logger: &mockLogger{}},
		format: "test message %d",
		args:   []any{1},
		want:   []string{"test message %d"},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.Infof(tc.format, tc.args...)
			if tc.config.Logger != nil {
				got := tc.config.Logger.(*mockLogger).infofCalls
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("Infof() calls = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestConfigInfoLevel(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{{
		name:   "LoggerIsNil",
		config: Config{},
		want:   false,
	}, {
		name:   "LoggerNotNil",
		config: Config{Logger: &mockLogger{}},
		want:   true,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.InfoLevel()
			if got != tc.want {
				t.Errorf("InfoLevel() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAddFieldSortOrder(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		nodeName   string
		fieldOrder []string
		want       map[string][]string
	}{{
		name:       "NilFieldSortOrder",
		config:     Config{},
		nodeName:   "node1",
		fieldOrder: []string{"field1", "field2"},
		want: map[string][]string{
			"node1": {"field1", "field2"},
		},
	}, {
		name: "ExistingFieldSortOrder",
		config: Config{
			FieldSortOrder: map[string][]string{
				"node1": {"fieldA", "fieldB"},
			},
		},
		nodeName:   "node2",
		fieldOrder: []string{"fieldC", "fieldD"},
		want: map[string][]string{
			"node1": {"fieldA", "fieldB"},
			"node2": {"fieldC", "fieldD"},
		},
	}, {
		name: "OverwriteExistingNode",
		config: Config{
			FieldSortOrder: map[string][]string{
				"node1": {"fieldA", "fieldB"},
			},
		},
		nodeName:   "node1",
		fieldOrder: []string{"fieldC", "fieldD"},
		want: map[string][]string{
			"node1": {"fieldC", "fieldD"},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.AddFieldSortOrder(tc.nodeName, tc.fieldOrder...)
			got := tc.config.FieldSortOrder
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("AddFieldSortOrder() FieldSortOrder = %v, want %v", got, tc.want)
			}
		})
	}
}
