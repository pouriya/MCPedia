package importfm

import (
	"strings"
	"testing"
)

func TestParseImportFile_Valid(t *testing.T) {
	content := `---
title: "Rust Error Handling"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, errors, result]
description: "Idiomatic error handling patterns in Rust"
---

# Rust Error Handling

Use Result for recoverable errors.
`
	e, err := ParseImportFile([]byte(content), "rust-error-handling.md")
	if err != nil {
		t.Fatalf("ParseImportFile: %v", err)
	}
	if e.Slug != "rust-error-handling" {
		t.Errorf("slug: got %q", e.Slug)
	}
	if e.Title != "Rust Error Handling" {
		t.Errorf("title: got %q", e.Title)
	}
	if e.Kind != "skill" {
		t.Errorf("kind: got %q", e.Kind)
	}
	if e.Language != "rust" {
		t.Errorf("language: got %q", e.Language)
	}
	if e.Description != "Idiomatic error handling patterns in Rust" {
		t.Errorf("description: got %q", e.Description)
	}
	wantTags := []string{"rust", "errors", "result"}
	if len(e.Tags) != len(wantTags) || e.Tags[0] != wantTags[0] || e.Tags[1] != wantTags[1] || e.Tags[2] != wantTags[2] {
		t.Errorf("tags: got %v", e.Tags)
	}
	body := strings.TrimSpace(e.Content)
	if !strings.Contains(body, "# Rust Error Handling") || !strings.Contains(body, "Use Result") {
		t.Errorf("content: got %q", e.Content)
	}
}

func TestParseImportFile_SlugFromFilename(t *testing.T) {
	content := `---
title: "X"
kind: skill
language: go
domain: ""
project: ""
tags: []
---

body
`
	e, err := ParseImportFile([]byte(content), "path/to/golang-context.md")
	if err != nil {
		t.Fatalf("ParseImportFile: %v", err)
	}
	if e.Slug != "golang-context" {
		t.Errorf("slug: got %q", e.Slug)
	}
}

func TestParseImportFile_NoLeadingDelimiter(t *testing.T) {
	content := `title: "X"
kind: skill
---
body`
	_, err := ParseImportFile([]byte(content), "x.md")
	if err == nil {
		t.Fatal("expected error for missing leading ---")
	}
	if !strings.Contains(err.Error(), "start with") {
		t.Errorf("error: %v", err)
	}
}

func TestParseImportFile_NoClosingDelimiter(t *testing.T) {
	content := `---
title: "X"
kind: skill
language: ""
domain: ""
project: ""
tags: []
body`
	_, err := ParseImportFile([]byte(content), "x.md")
	if err == nil {
		t.Fatal("expected error for missing closing ---")
	}
	if !strings.Contains(err.Error(), "closing") {
		t.Errorf("error: %v", err)
	}
}

func TestParseImportFile_UnknownKey(t *testing.T) {
	content := `---
title: "X"
kind: skill
language: ""
domain: ""
project: ""
tags: []
extra: bad
---

body
`
	_, err := ParseImportFile([]byte(content), "x.md")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error: %v", err)
	}
}

func TestParseImportFile_InvalidKind(t *testing.T) {
	content := `---
title: "X"
kind: invalid
language: ""
domain: ""
project: ""
tags: []
---

body
`
	_, err := ParseImportFile([]byte(content), "x.md")
	if err == nil {
		t.Fatal("expected error for invalid kind")
	}
	if !strings.Contains(err.Error(), "kind") {
		t.Errorf("error: %v", err)
	}
}

func TestParseImportFile_MissingRequiredKey(t *testing.T) {
	content := `---
title: "X"
language: ""
domain: ""
project: ""
tags: []
---

body
`
	_, err := ParseImportFile([]byte(content), "x.md")
	if err == nil {
		t.Fatal("expected error for missing kind")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error: %v", err)
	}
}

func TestParseImportFile_ContentTooLong(t *testing.T) {
	content := `---
title: "X"
kind: skill
language: ""
domain: ""
project: ""
tags: []
---

` + strings.Repeat("x", maxContentLen+1)
	_, err := ParseImportFile([]byte(content), "x.md")
	if err == nil {
		t.Fatal("expected error for content too long")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("error: %v", err)
	}
}

func TestParseImportFile_InvalidFilename(t *testing.T) {
	content := `---
title: "X"
kind: skill
language: ""
domain: ""
project: ""
tags: []
---

body
`
	_, err := ParseImportFile([]byte(content), "not-md.txt")
	if err == nil {
		t.Fatal("expected error for non-.md filename")
	}
	_, err = ParseImportFile([]byte(content), ".md")
	if err == nil {
		t.Fatal("expected error for .md only")
	}
}

func TestParseImportFile_InvalidSlug(t *testing.T) {
	content := `---
title: "X"
kind: skill
language: ""
domain: ""
project: ""
tags: []
---

body
`
	_, err := ParseImportFile([]byte(content), "UPPERCASE.md")
	if err == nil {
		t.Fatal("expected error for invalid slug (uppercase)")
	}
	_, err = ParseImportFile([]byte(content), "has space.md")
	if err == nil {
		t.Fatal("expected error for invalid slug (space)")
	}
}

func TestParseImportFile_DescriptionOptional(t *testing.T) {
	content := `---
title: "No Desc"
kind: rule
language: go
domain: ""
project: ""
tags: [go]
---

body
`
	e, err := ParseImportFile([]byte(content), "no-desc.md")
	if err != nil {
		t.Fatalf("ParseImportFile: %v", err)
	}
	if e.Description != "" {
		t.Errorf("description: got %q", e.Description)
	}
}
