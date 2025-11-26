package common

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	autoPrereqCommentPattern  = regexp.MustCompile(`^#\s*ie:auto-prereq-([a-z-]+)\s+(.*)$`)
	autoPrereqMetadataPattern = regexp.MustCompile(`([a-zA-Z0-9_-]+)="([^"]*)"`)
)

// ParseAutoPrereqMetadata extracts prerequisite metadata encoded in the
// leading comment that the prerequisite injector adds to each block.
func ParseAutoPrereqMetadata(content string) (string, map[string]string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "", nil, false
	}

	firstLine := strings.TrimSpace(lines[0])
	matches := autoPrereqCommentPattern.FindStringSubmatch(firstLine)
	if len(matches) != 3 {
		return "", nil, false
	}

	metadata := make(map[string]string)
	for _, match := range autoPrereqMetadataPattern.FindAllStringSubmatch(matches[2], -1) {
		if len(match) != 3 {
			continue
		}
		metadata[match[1]] = match[2]
	}

	return matches[1], metadata, true
}

// StripAutoPrereqComment removes the metadata comment header before firing the
// underlying command so bash does not see our internal annotations.
func StripAutoPrereqComment(content string) string {
	if _, _, hasMetadata := ParseAutoPrereqMetadata(content); !hasMetadata {
		return content
	}

	parts := strings.SplitN(content, "\n", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

// WritePrereqMarker persists a marker file that signals the prerequisite body
// should be skipped because verification passed.
func WritePrereqMarker(markerPath, display string) error {
	if strings.TrimSpace(markerPath) == "" {
		return nil
	}

	dir := filepath.Dir(markerPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(markerPath, []byte(display), 0o600)
}

// RemovePrereqMarker ensures prior verification state does not leak into the
// next execution attempt.
func RemovePrereqMarker(markerPath string) error {
	if strings.TrimSpace(markerPath) == "" {
		return nil
	}

	err := os.Remove(markerPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
