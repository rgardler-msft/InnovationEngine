package main

import (
	"fmt"
	"regexp"
)

func main() {
	// Simulate kubectl output WITH ANSI color codes - different scenarios
	scenarios := []struct {
		name string
		text string
	}{
		{
			"ANSI in Kubernetes word",
			"\x1b[32mKubernetes\x1b[0m control plane is running at https://...",
		},
		{
			"ANSI between words",
			"Kubernetes\x1b[0m control\x1b[32m plane is running at https://...",
		},
		{
			"ANSI in 'control plane'",
			"Kubernetes control\x1b[0m plane is running at https://...",
		},
		{
			"ANSI in 'is running'",
			"Kubernetes control plane is\x1b[32m running\x1b[0m at https://...",
		},
		{
			"No ANSI",
			"Kubernetes control plane is running at https://...",
		},
	}

	// The regex pattern
	pattern := `(?s).*control plane is running.*`
	re := regexp.MustCompile(pattern)

	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)

	for _, scenario := range scenarios {
		fmt.Printf("\n%s:\n", scenario.name)
		fmt.Printf("  Text: %q\n", scenario.text)

		matches := re.MatchString(scenario.text)
		fmt.Printf("  Matches: %v\n", matches)

		if !matches {
			stripped := ansiPattern.ReplaceAllString(scenario.text, "")
			fmt.Printf("  Stripped: %q\n", stripped)
			matchesAfter := re.MatchString(stripped)
			fmt.Printf("  Matches after strip: %v\n", matchesAfter)
		}
	}
}
