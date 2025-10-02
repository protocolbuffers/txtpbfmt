package descriptor

import (
	"testing"

	// Google internal testing/gobase/runfilestest package, commented out by copybara
)

func TestNewLoader(t *testing.T) {
	t.Run("valid descriptor file", func(t *testing.T) {
		descriptorFile := "../testdata/test.desc"
		loader, err := NewLoader(descriptorFile)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}
		if loader == nil {
			t.Fatal("Expected non-nil loader")
		}
	})

	t.Run("empty descriptor file path", func(t *testing.T) {
		_, err := NewLoader("")
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := NewLoader("nonexistent.desc")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestGetRootMessageDescriptor(t *testing.T) {
	descriptorFile := "../testdata/test.desc"
	loader, err := NewLoader(descriptorFile)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	tests := []struct {
		name            string
		messageFullName string
		wantError       bool
	}{
		{"UserProfile", "testproto.UserProfile", false},
		{"ProductCatalog", "testproto.ProductCatalog", false},
		{"Level1Config", "testproto.Level1Config", false},
		{"nested message", "testproto.Level1Config.Level2Config", false},
		{"empty name", "", true},
		{"non-existent", "testproto.NonExistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, err := loader.GetRootMessageDescriptor(tt.messageFullName)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if desc == nil {
					t.Error("Expected descriptor but got nil")
				}
			}
		})
	}
}
