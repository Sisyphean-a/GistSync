package profileutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"GistSync/internal/pathmap"
)

var windowsDrivePrefix = regexp.MustCompile(`^[A-Za-z]:/+`)

func NormalizeRelativePath(sourcePath string) string {
	path := strings.ReplaceAll(sourcePath, "\\", "/")
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	path = windowsDrivePrefix.ReplaceAllString(path, "")
	path = strings.TrimPrefix(path, "{{HOME}}/")
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, ":", "")
	return path
}

func StableItemID(sourcePath string, relativePath string) string {
	return StableItemIDForOccurrence(sourcePath, relativePath, 0)
}

func StableItemIDForOccurrence(sourcePath string, relativePath string, occurrence int) string {
	source := normalizeItemPath(pathmap.CompactHomePath(sourcePath))
	relative := normalizeItemRelativePath(source, relativePath)
	sum := sha256.Sum256([]byte(source + "|" + relative))
	base := "item-" + hex.EncodeToString(sum[:12])
	if occurrence <= 0 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, occurrence+1)
}

func GenerateID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, randomHex(8))
}

func GenerateScopedID(prefix string, scope string) string {
	trimmedScope := strings.TrimSpace(scope)
	if trimmedScope == "" {
		return GenerateID(prefix)
	}
	return fmt.Sprintf("%s-%s-%s", prefix, trimmedScope, randomHex(6))
}

func normalizeItemPath(value string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	for strings.Contains(normalized, "//") {
		normalized = strings.ReplaceAll(normalized, "//", "/")
	}
	return normalized
}

func normalizeItemRelativePath(sourcePath string, relativePath string) string {
	normalizedSource := NormalizeRelativePath(sourcePath)
	if strings.TrimSpace(normalizedSource) != "" {
		return normalizedSource
	}
	if strings.TrimSpace(relativePath) != "" {
		return NormalizeRelativePath(relativePath)
	}
	return ""
}

func randomHex(byteLen int) string {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
