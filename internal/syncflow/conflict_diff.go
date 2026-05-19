package syncflow

import (
	"bytes"
	"context"
	"os"
	"strings"

	"GistSync/internal/security"
)

const (
	diffStatusReady      = "ready"
	diffStatusBinary     = "binary_unsupported"
	diffStatusTooLarge   = "too_large"
	diffStatusDecodeFail = "decode_failed"
	diffStatusReadFail   = "read_failed"
)

const (
	maxDiffSourceBytes  = 200 * 1024
	maxDiffPreviewRunes = 12 * 1024
)

func (s *Service) buildConflict(
	ctx context.Context,
	gistID string,
	item manifestSnapshotItem,
	req ApplySnapshotRequest,
	targetPath string,
) ApplyConflict {
	conflict := ApplyConflict{
		ItemID:     item.ItemID,
		TargetPath: targetPath,
	}
	diff, status := s.buildDiffPreview(ctx, gistID, item, req.MasterPassword, targetPath)
	conflict.DiffPreview = diff
	conflict.DiffStatus = status
	return conflict
}

func (s *Service) buildDiffPreview(
	ctx context.Context,
	gistID string,
	item manifestSnapshotItem,
	password string,
	targetPath string,
) (string, string) {
	if empty(password) {
		return "", diffStatusDecodeFail
	}
	localRaw, err := os.ReadFile(targetPath)
	if err != nil {
		return "", diffStatusReadFail
	}
	if len(localRaw) > maxDiffSourceBytes {
		return "", diffStatusTooLarge
	}
	if isLikelyBinary(localRaw) {
		return "", diffStatusBinary
	}
	encrypted, err := s.cloud.GetFileContent(ctx, FileRequest{GistID: gistID, FileName: item.BlobFile})
	if err != nil {
		return "", diffStatusReadFail
	}
	remoteText, err := security.DecryptString(encrypted, password)
	if err != nil {
		return "", diffStatusDecodeFail
	}
	remoteRaw := []byte(remoteText)
	if len(remoteRaw) > maxDiffSourceBytes {
		return "", diffStatusTooLarge
	}
	if isLikelyBinary(remoteRaw) {
		return "", diffStatusBinary
	}
	diff := buildSimpleUnifiedDiff(string(localRaw), remoteText)
	return trimRunes(diff, maxDiffPreviewRunes), diffStatusReady
}

func buildSimpleUnifiedDiff(local string, remote string) string {
	left := splitLines(local)
	right := splitLines(remote)
	var out strings.Builder
	out.WriteString("--- local\n")
	out.WriteString("+++ remote\n")
	out.WriteString("@@ conflict @@\n")

	maxLen := max(len(left), len(right))
	for i := 0; i < maxLen; i++ {
		var l string
		var r string
		hasL := i < len(left)
		hasR := i < len(right)
		if hasL {
			l = left[i]
		}
		if hasR {
			r = right[i]
		}
		if hasL && hasR && l == r {
			continue
		}
		if hasL {
			out.WriteString("-")
			out.WriteString(l)
			out.WriteString("\n")
		}
		if hasR {
			out.WriteString("+")
			out.WriteString(r)
			out.WriteString("\n")
		}
	}

	diff := out.String()
	if diff == "--- local\n+++ remote\n@@ conflict @@\n" {
		return diff + " (no textual difference)\n"
	}
	return diff
}

func splitLines(raw string) []string {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	return strings.Split(normalized, "\n")
}

func trimRunes(input string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(input)
	if len(runes) <= limit {
		return input
	}
	return string(runes[:limit]) + "\n... [diff truncated]"
}

func isLikelyBinary(raw []byte) bool {
	if len(raw) == 0 {
		return false
	}
	if bytes.IndexByte(raw, 0x00) >= 0 {
		return true
	}
	printable := 0
	for _, b := range raw {
		if b == '\n' || b == '\r' || b == '\t' || (b >= 0x20 && b <= 0x7E) {
			printable++
		}
	}
	return float64(printable)/float64(len(raw)) < 0.9
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
