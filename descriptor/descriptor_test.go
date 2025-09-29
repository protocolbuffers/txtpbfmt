package descriptor

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDescriptor(t *testing.T) {
	descriptorFile := filepath.Join("..", "testdata", "sort_by_field_number_test.desc")

	loader := NewLoader(descriptorFile)
	err := loader.LoadDescriptor()
	if err != nil {
		t.Fatalf("Failed to load descriptor: %v", err)
	}

	if loader.files == nil {
		t.Error("files should not be nil after loading")
	}
}

func TestInvalidDescriptorFile(t *testing.T) {
	loader := NewLoader("nonexistent.desc")
	err := loader.LoadDescriptor()
	if err == nil {
		t.Error("Expected error when loading nonexistent descriptor file")
	}
}

func TestGetRootMessageDescriptor(t *testing.T) {
	descriptorFile := filepath.Join("..", "testdata", "test.desc")

	loader := NewLoader(descriptorFile)
	err := loader.LoadDescriptor()
	if err != nil {
		t.Fatalf("Failed to load descriptor: %v", err)
	}

	tests := []struct {
		name            string
		messageFullName string
		expectError     bool
		expectedName    string
	}{
		{
			name:            "valid message full name - UserProfile",
			messageFullName: "testproto.UserProfile",
			expectError:     false,
			expectedName:    "testproto.UserProfile",
		},
		{
			name:            "valid message full name - ProductCatalog",
			messageFullName: "testproto.ProductCatalog",
			expectError:     false,
			expectedName:    "testproto.ProductCatalog",
		},
		{
			name:            "empty message full name - should error",
			messageFullName: "",
			expectError:     true,
		},
		{
			name:            "non-existent message full name",
			messageFullName: "testproto.NonExistentMessage",
			expectError:     true,
		},
		{
			name:            "invalid message full name format",
			messageFullName: "InvalidName",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, err := loader.GetRootMessageDescriptor(tt.messageFullName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for messageFullName: %s, but got none", tt.messageFullName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for messageFullName: %s, got: %v", tt.messageFullName, err)
				}
				if desc == nil {
					t.Errorf("Expected descriptor for messageFullName: %s, but got nil", tt.messageFullName)
				} else {
					actualName := string(desc.FullName())
					if actualName != tt.expectedName {
						t.Errorf("Expected descriptor full name: %s, but got: %s", tt.expectedName, actualName)
					}
				}
			}
		})
	}
}

func TestGetRootMessageDescriptorWithoutLoadDescriptor(t *testing.T) {
	loader := NewLoader("any.desc")
	// Don't call LoadDescriptor

	_, err := loader.GetRootMessageDescriptor("testproto.UserProfile")
	if err == nil {
		t.Error("Expected error when calling GetRootMessageDescriptor without LoadDescriptor")
	}

	expectedError := "descriptor not loaded"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, but got: %v", expectedError, err)
	}
}
