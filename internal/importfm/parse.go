package importfm

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pouriya/mcpedia/internal/db"
)

const maxContentLen = 32768

var validKinds = map[string]bool{
	"skill": true, "rule": true, "context": true,
	"pattern": true, "reference": true, "guide": true,
}

// slugRegex matches export-style slugs: lowercase letters, digits, hyphens.
var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// allowedKeys is the exact set of keys export produces; unknown keys are rejected.
var allowedKeys = map[string]bool{
	"title": true, "kind": true, "language": true, "domain": true,
	"project": true, "tags": true, "description": true,
}

// ParseImportFile parses file content (export-format Markdown with YAML frontmatter)
// and the given filename (used to derive slug). Returns a db.Entry ready for CreateEntry,
// or an error if the format is invalid.
func ParseImportFile(content []byte, filename string) (*db.Entry, error) {
	slug, err := slugFromFilename(filename)
	if err != nil {
		return nil, err
	}

	raw := string(content)
	if !strings.HasPrefix(raw, "---\n") {
		return nil, fmt.Errorf("invalid format: file must start with \"---\"")
	}

	// Find end of frontmatter: first "---" after the opening one, on its own line
	rest := raw[len("---\n"):]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return nil, fmt.Errorf("invalid format: missing closing \"---\" for frontmatter")
	}
	frontmatter := strings.TrimSpace(rest[:idx])
	body := rest[idx+4:] // skip "\n---"
	body = strings.TrimPrefix(body, "\n")
	body = strings.TrimSpace(body)

	if len(body) > maxContentLen {
		return nil, fmt.Errorf("invalid format: content exceeds %d bytes", maxContentLen)
	}

	meta, err := parseFrontmatter(frontmatter)
	if err != nil {
		return nil, err
	}

	kind := meta["kind"]
	if !validKinds[kind] {
		return nil, fmt.Errorf("invalid format: kind %q is not one of skill, rule, context, pattern, reference, guide", kind)
	}

	tags := meta["tags"]
	if tags == "" {
		tags = "[]"
	}
	tagList := parseTagsList(tags)

	e := &db.Entry{
		Slug:        slug,
		Title:       meta["title"],
		Description: meta["description"],
		Content:     body,
		Kind:        kind,
		Language:    meta["language"],
		Domain:      meta["domain"],
		Project:     meta["project"],
		Tags:        tagList,
	}
	return e, nil
}

func slugFromFilename(filename string) (string, error) {
	base := filepath.Base(filename)
	if base == "." || base == "/" {
		return "", fmt.Errorf("invalid format: filename must be a .md file")
	}
	slug := strings.TrimSuffix(base, ".md")
	if slug == base {
		return "", fmt.Errorf("invalid format: filename must end with .md")
	}
	if slug == "" {
		return "", fmt.Errorf("invalid format: slug derived from filename is empty")
	}
	if !slugRegex.MatchString(slug) {
		return "", fmt.Errorf("invalid format: slug %q must match [a-z0-9][a-z0-9-]*", slug)
	}
	return slug, nil
}

// parseFrontmatter parses the frontmatter block. Only allowed keys are accepted.
// Returns map of key -> value (raw string); tags are kept as "[a, b, c]" and parsed separately.
func parseFrontmatter(block string) (map[string]string, error) {
	seen := make(map[string]bool)
	out := map[string]string{
		"title": "", "kind": "", "language": "", "domain": "", "project": "",
		"tags": "", "description": "",
	}
	lines := strings.Split(block, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		colon := strings.Index(line, ":")
		if colon <= 0 {
			return nil, fmt.Errorf("invalid format: frontmatter line %q has no key", line)
		}
		key := strings.TrimSpace(line[:colon])
		if !allowedKeys[key] {
			return nil, fmt.Errorf("invalid format: unknown frontmatter key %q", key)
		}
		if seen[key] {
			return nil, fmt.Errorf("invalid format: duplicate key %q", key)
		}
		seen[key] = true
		val := strings.TrimSpace(line[colon+1:])
		out[key] = unquoteVal(val)
	}
	// Required keys (export always writes these)
	for _, k := range []string{"title", "kind", "language", "domain", "project", "tags"} {
		if !seen[k] {
			return nil, fmt.Errorf("invalid format: missing required key %q", k)
		}
	}
	if out["title"] == "" {
		return nil, fmt.Errorf("invalid format: title must not be empty")
	}
	return out, nil
}

func unquoteVal(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		// Simple unquote: no escape handling for simplicity; export uses double quotes
		return s[1 : len(s)-1]
	}
	return s
}

// parseTagsList parses "[a, b, c]" or "[]" into a slice of strings.
func parseTagsList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "[]" || s == "" {
		return nil
	}
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return nil
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = unquoteVal(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}
