package environments

import "testing"

func TestIsValidEnvironment_KnownValues(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		isValid bool
		isAzure bool
	}{
		{"local", string(EnvironmentsLocal), true, false},
		{"github-action", string(EnvironmentsGithubAction), true, false},
		{"ocd", string(EnvironmentsOCD), true, true},
		{"azure", string(EnvironmentsAzure), true, true},
		{"unknown", "something-else", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEnvironment(tt.input); got != tt.isValid {
				t.Fatalf("IsValidEnvironment(%q) = %v, want %v", tt.input, got, tt.isValid)
			}
			if gotAzure := IsAzureEnvironment(tt.input); gotAzure != tt.isAzure {
				t.Fatalf("IsAzureEnvironment(%q) = %v, want %v", tt.input, gotAzure, tt.isAzure)
			}
		})
	}
}
