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

func TestParseEnvironment(t *testing.T) {
	t.Run("valid values", func(t *testing.T) {
		for _, expected := range []Environment{
			EnvironmentsLocal,
			EnvironmentsGithubAction,
			EnvironmentsOCD,
			EnvironmentsAzure,
		} {
			parsed, err := Parse(string(expected))
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", expected, err)
			}
			if parsed != expected {
				t.Fatalf("Parse(%q) = %q, want %q", expected, parsed, expected)
			}
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		if _, err := Parse("bogus-env"); err == nil {
			t.Fatalf("Parse should error for invalid value")
		}
	})
}

func TestEnvironmentHelpers(t *testing.T) {
	tests := []struct {
		name        string
		value       Environment
		isGithub    bool
		isAzureLike bool
	}{
		{"local", EnvironmentsLocal, false, false},
		{"github-action", EnvironmentsGithubAction, true, false},
		{"ocd", EnvironmentsOCD, false, true},
		{"azure", EnvironmentsAzure, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsGithubAction(); got != tt.isGithub {
				t.Fatalf("IsGithubAction for %q = %v, want %v", tt.value, got, tt.isGithub)
			}
			if got := tt.value.IsAzureLike(); got != tt.isAzureLike {
				t.Fatalf("IsAzureLike for %q = %v, want %v", tt.value, got, tt.isAzureLike)
			}
		})
	}
}
