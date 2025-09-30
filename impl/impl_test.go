package impl

import (
	"strings"
	"testing"

	// Google internal testing/gobase/runfilestest package, commented out by copybara
	"github.com/kylelemons/godebug/pretty"
	"github.com/protocolbuffers/txtpbfmt/config"
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

func TestParseWithMetaCommentConfig_ErrorHandling(t *testing.T) {
	descriptorFile := "../testdata/test.desc"

	tests := []struct {
		name           string
		config         config.Config
		input          string
		expectError    bool
		errorSubstring string
	}{
		{
			name: "SortFieldsByFieldNumber without ProtoDescriptor",
			config: config.Config{
				SortFieldsByFieldNumber: true,
				ProtoDescriptor:         "",
				MessageFullName:         "testproto.UserProfile",
			},
			input:          `name: "test"`,
			expectError:    true,
			errorSubstring: "proto_descriptor is required",
		},
		{
			name: "SortFieldsByFieldNumber without MessageFullName",
			config: config.Config{
				SortFieldsByFieldNumber: true,
				ProtoDescriptor:         descriptorFile,
				MessageFullName:         "",
			},
			input:          `name: "test"`,
			expectError:    true,
			errorSubstring: "message_full_name is required",
		},
		{
			name: "SortFieldsByFieldNumber with invalid descriptor file",
			config: config.Config{
				SortFieldsByFieldNumber: true,
				ProtoDescriptor:         "nonexistent.desc",
				MessageFullName:         "testproto.UserProfile",
			},
			input:          `name: "test"`,
			expectError:    true,
			errorSubstring: "failed to read descriptor file",
		},
		{
			name: "SortFieldsByFieldNumber with invalid message name",
			config: config.Config{
				SortFieldsByFieldNumber: true,
				ProtoDescriptor:         descriptorFile,
				MessageFullName:         "testproto.NonExistentMessage",
			},
			input:          `name: "test"`,
			expectError:    true,
			errorSubstring: "failed to get root message descriptor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseWithMetaCommentConfig([]byte(tt.input), tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test %s, but got none", tt.name)
				} else if !strings.Contains(err.Error(), tt.errorSubstring) {
					t.Errorf("Expected error to contain %q, but got: %v", tt.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for test %s: %v", tt.name, err)
				}
			}
		})
	}
}

func TestParseWithMetaCommentConfig_SortFieldsByFieldNumber(t *testing.T) {
	descriptorFile := "../testdata/test.desc"

	tests := []struct {
		name                  string
		config                config.Config
		input                 string
		expectSuccess         bool
		expectedFieldsInOrder []string // Expected field names in order
	}{
		{
			name: "Sort fields by field number - UserProfile",
			config: config.Config{
				SortFieldsByFieldNumber: true,
				ProtoDescriptor:         descriptorFile,
				MessageFullName:         "testproto.UserProfile",
			},
			input: `name: "test"
priority: 10
id: 1
email: "test@example.com"
active: true`,
			expectSuccess: true,
			// Expected order based on field numbers: id=1, priority=2, email=3, name=5, active=7
			expectedFieldsInOrder: []string{"id", "priority", "email", "name", "active"},
		},
		{
			name: "Sort fields by field number - ProductCatalog",
			config: config.Config{
				SortFieldsByFieldNumber: true,
				ProtoDescriptor:         descriptorFile,
				MessageFullName:         "testproto.ProductCatalog",
			},
			input: `name: "product_test"
priority: 100
id: 1
enabled: true`,
			expectSuccess: true,
			// Expected order based on field numbers: id=1, enabled=2, name=3, priority=4
			expectedFieldsInOrder: []string{"id", "enabled", "name", "priority"},
		},
		{
			name: "Regular parsing without field number sorting",
			config: config.Config{
				SortFieldsByFieldNumber: false,
			},
			input: `name: "test"
priority: 10
id: 1`,
			expectSuccess: true,
			// Should maintain original order
			expectedFieldsInOrder: []string{"name", "priority", "id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseWithMetaCommentConfig([]byte(tt.input), tt.config)

			if tt.expectSuccess {
				if err != nil {
					t.Fatalf("Unexpected error for test %s: %v", tt.name, err)
				}

				if len(nodes) == 0 {
					t.Fatalf("Expected nodes but got empty result for test %s", tt.name)
				}

				// Extract field names in order
				var actualFieldOrder []string
				for _, node := range nodes {
					if node.Name != "" {
						actualFieldOrder = append(actualFieldOrder, node.Name)
					}
				}

				// Compare with expected order
				if len(actualFieldOrder) != len(tt.expectedFieldsInOrder) {
					t.Errorf("Expected %d fields, but got %d for test %s",
						len(tt.expectedFieldsInOrder), len(actualFieldOrder), tt.name)
				}

				for i, expected := range tt.expectedFieldsInOrder {
					if i >= len(actualFieldOrder) {
						t.Errorf("Missing field at index %d, expected %s for test %s",
							i, expected, tt.name)
						break
					}
					if actualFieldOrder[i] != expected {
						t.Errorf("Field at index %d: expected %s, got %s for test %s",
							i, expected, actualFieldOrder[i], tt.name)
					}
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for test %s, but parsing succeeded", tt.name)
				}
			}
		})
	}
}
