package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pouriya/mcpedia/internal/db"
	"github.com/pouriya/mcpedia/internal/mcp"
)

// jsonrpcResponse mirrors the unexported type for test decoding.
type jsonrpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// --- test helpers ---

func setup(t *testing.T) (*mcp.Server, *httptest.Server) {
	t.Helper()
	d, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	s := &mcp.Server{DB: d}
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)
	return s, ts
}

func setupWithToken(t *testing.T, token string) (*mcp.Server, *httptest.Server) {
	t.Helper()
	d, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	s := &mcp.Server{DB: d, Token: token}
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)
	return s, ts
}

// call sends a JSON-RPC request and returns the parsed response.
func call(t *testing.T, url string, method string, id any, params any, headers map[string]string) (int, jsonrpcResponse) {
	t.Helper()
	body := map[string]any{"jsonrpc": "2.0", "method": method}
	if id != nil {
		body["id"] = id
	}
	if params != nil {
		body["params"] = params
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusAccepted {
		return resp.StatusCode, jsonrpcResponse{}
	}
	var result jsonrpcResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return resp.StatusCode, result
}

// toolCall is a shortcut for tools/call.
func toolCall(t *testing.T, url string, name string, args map[string]any) (jsonrpcResponse, string, bool) {
	t.Helper()
	_, resp := call(t, url, "tools/call", 1, map[string]any{"name": name, "arguments": args}, nil)
	if resp.Error != nil {
		return resp, "", false
	}
	result := resp.Result.(map[string]any)
	isErr, _ := result["isError"].(bool)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	return resp, text, isErr
}

// createEntry is a shortcut to create an entry via the tool.
func createEntry(t *testing.T, url string, slug, title, content, kind, language, domain, project string, tags []string) {
	t.Helper()
	args := map[string]any{"slug": slug, "title": title, "content": content}
	if kind != "" {
		args["kind"] = kind
	}
	if language != "" {
		args["language"] = language
	}
	if domain != "" {
		args["domain"] = domain
	}
	if project != "" {
		args["project"] = project
	}
	if tags != nil {
		args["tags"] = tags
	}
	_, text, isErr := toolCall(t, url, "create_entry", args)
	if isErr {
		t.Fatalf("create_entry failed: %s", text)
	}
}

// --- tests ---

func TestInitializeHandshake(t *testing.T) {
	_, ts := setup(t)

	status, resp := call(t, ts.URL, "initialize", 1, map[string]any{
		"protocolVersion": "2025-11-25",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
	}, nil)

	if status != 200 {
		t.Fatalf("status: %d", status)
	}
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	r := resp.Result.(map[string]any)
	if r["protocolVersion"] != "2025-11-25" {
		t.Errorf("protocolVersion: %v", r["protocolVersion"])
	}
	si := r["serverInfo"].(map[string]any)
	if si["name"] != "mcpedia" || si["version"] != "0.1.0" {
		t.Errorf("serverInfo: %v", si)
	}
	caps := r["capabilities"].(map[string]any)
	for _, key := range []string{"tools", "resources", "prompts"} {
		if caps[key] == nil {
			t.Errorf("missing capability: %s", key)
		}
	}
}

func TestToolsList(t *testing.T) {
	_, ts := setup(t)
	_, resp := call(t, ts.URL, "tools/list", 1, nil, nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	tools := resp.Result.(map[string]any)["tools"].([]any)
	if len(tools) != 8 {
		t.Fatalf("expected 8 tools, got %d", len(tools))
	}
	names := map[string]bool{}
	for _, tool := range tools {
		tm := tool.(map[string]any)
		names[tm["name"].(string)] = true
		if tm["inputSchema"] == nil {
			t.Errorf("tool %s missing inputSchema", tm["name"])
		}
	}
	for _, want := range []string{"search_entries", "get_entry", "get_entries_by_context", "list_entries", "list_tags", "create_entry", "update_entry", "delete_entry"} {
		if !names[want] {
			t.Errorf("missing tool: %s", want)
		}
	}
}

func TestCreateAndGetEntry(t *testing.T) {
	_, ts := setup(t)

	// Create
	_, text, isErr := toolCall(t, ts.URL, "create_entry", map[string]any{
		"slug": "rust-errors", "title": "Rust Error Handling",
		"content": "Use Result<T, E>.", "description": "Error handling in Rust",
		"kind": "skill", "language": "rust", "domain": "systems", "project": "myproj",
		"tags": []string{"rust", "errors", "result"},
	})
	if isErr {
		t.Fatalf("create error: %s", text)
	}
	var created db.Entry
	json.Unmarshal([]byte(text), &created)
	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if created.Version != 1 {
		t.Errorf("version: %d", created.Version)
	}
	if created.CreatedAt == "" {
		t.Error("missing created_at")
	}

	// Get
	_, text, isErr = toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "rust-errors"})
	if isErr {
		t.Fatalf("get error: %s", text)
	}
	var got db.Entry
	json.Unmarshal([]byte(text), &got)
	if got.Slug != "rust-errors" {
		t.Errorf("slug: %q", got.Slug)
	}
	if got.Title != "Rust Error Handling" {
		t.Errorf("title: %q", got.Title)
	}
	if got.Content != "Use Result<T, E>." {
		t.Errorf("content: %q", got.Content)
	}
	if got.Description != "Error handling in Rust" {
		t.Errorf("description: %q", got.Description)
	}
	if got.Kind != "skill" {
		t.Errorf("kind: %q", got.Kind)
	}
	if got.Language != "rust" {
		t.Errorf("language: %q", got.Language)
	}
	if got.Domain != "systems" {
		t.Errorf("domain: %q", got.Domain)
	}
	if got.Project != "myproj" {
		t.Errorf("project: %q", got.Project)
	}
	if len(got.Tags) != 3 || got.Tags[0] != "errors" || got.Tags[1] != "result" || got.Tags[2] != "rust" {
		t.Errorf("tags: %v", got.Tags)
	}
}

func TestCreateEntryValidation(t *testing.T) {
	_, ts := setup(t)

	// Missing required fields
	_, text, isErr := toolCall(t, ts.URL, "create_entry", map[string]any{"slug": "x"})
	if !isErr {
		t.Fatal("expected error for missing title+content")
	}
	if !strings.Contains(text, "required") {
		t.Errorf("expected 'required' in error, got: %s", text)
	}

	// Duplicate slug
	createEntry(t, ts.URL, "dup", "Dup", "content", "", "", "", "", nil)
	_, text, isErr = toolCall(t, ts.URL, "create_entry", map[string]any{
		"slug": "dup", "title": "Dup2", "content": "content2",
	})
	if !isErr {
		t.Fatal("expected error for duplicate slug")
	}

	// Content too large (32KB + 1 byte)
	_, text, isErr = toolCall(t, ts.URL, "create_entry", map[string]any{
		"slug": "big", "title": "Big", "content": strings.Repeat("x", 32769),
	})
	if !isErr {
		t.Fatal("expected error for content too large")
	}
}

func TestUpdateEntry(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "upd", "Original", "original content", "skill", "go", "", "", nil)

	_, text, isErr := toolCall(t, ts.URL, "update_entry", map[string]any{
		"slug": "upd", "title": "Updated Title", "content": "updated content",
	})
	if isErr {
		t.Fatalf("update error: %s", text)
	}
	var got db.Entry
	json.Unmarshal([]byte(text), &got)
	if got.Title != "Updated Title" {
		t.Errorf("title: %q", got.Title)
	}
	if got.Content != "updated content" {
		t.Errorf("content: %q", got.Content)
	}
	if got.Language != "go" {
		t.Errorf("language should be unchanged: %q", got.Language)
	}
	if got.Version != 2 {
		t.Errorf("version: %d", got.Version)
	}

	// Update non-existent
	_, text, isErr = toolCall(t, ts.URL, "update_entry", map[string]any{
		"slug": "nonexistent", "title": "nope",
	})
	if !isErr {
		t.Fatal("expected error for non-existent entry")
	}
}

func TestUpdateEntryTags(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "tg", "Tags", "content", "", "", "", "", []string{"alpha", "beta"})

	_, text, _ := toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "tg"})
	var e db.Entry
	json.Unmarshal([]byte(text), &e)
	if len(e.Tags) != 2 || e.Tags[0] != "alpha" || e.Tags[1] != "beta" {
		t.Fatalf("initial tags: %v", e.Tags)
	}

	_, text, _ = toolCall(t, ts.URL, "update_entry", map[string]any{
		"slug": "tg", "tags": []string{"gamma", "delta"},
	})
	json.Unmarshal([]byte(text), &e)
	if len(e.Tags) != 2 || e.Tags[0] != "delta" || e.Tags[1] != "gamma" {
		t.Errorf("updated tags: %v", e.Tags)
	}
}

func TestDeleteEntry(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "del", "Delete Me", "content", "", "", "", "", []string{"tag1"})

	_, _, isErr := toolCall(t, ts.URL, "delete_entry", map[string]any{"slug": "del"})
	if isErr {
		t.Fatal("delete failed")
	}

	_, _, isErr = toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "del"})
	if !isErr {
		t.Fatal("expected error for deleted entry")
	}

	_, _, isErr = toolCall(t, ts.URL, "delete_entry", map[string]any{"slug": "nonexistent"})
	if !isErr {
		t.Fatal("expected error for non-existent delete")
	}
}

func TestListEntries(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "go-test", "Go Testing", "test content", "skill", "go", "backend", "", nil)
	createEntry(t, ts.URL, "go-http", "Go HTTP", "http content", "skill", "go", "backend", "", nil)
	createEntry(t, ts.URL, "rust-cli", "Rust CLI", "cli content", "guide", "rust", "cli", "", nil)
	createEntry(t, ts.URL, "py-ml", "Python ML", "ml content", "skill", "python", "ml", "", []string{"ml"})
	createEntry(t, ts.URL, "general", "Git Basics", "git content", "reference", "", "", "", nil)

	_, text, _ := toolCall(t, ts.URL, "list_entries", map[string]any{})
	var all []db.Entry
	json.Unmarshal([]byte(text), &all)
	if len(all) != 5 {
		t.Errorf("all: got %d, want 5", len(all))
	}
	for _, e := range all {
		if e.Content != "" {
			t.Errorf("list should not include content for %s", e.Slug)
		}
	}

	_, text, _ = toolCall(t, ts.URL, "list_entries", map[string]any{"language": "go"})
	var goEntries []db.Entry
	json.Unmarshal([]byte(text), &goEntries)
	if len(goEntries) != 2 {
		t.Errorf("go entries: %d", len(goEntries))
	}

	_, text, _ = toolCall(t, ts.URL, "list_entries", map[string]any{"kind": "guide"})
	var guides []db.Entry
	json.Unmarshal([]byte(text), &guides)
	if len(guides) != 1 {
		t.Errorf("guides: %d", len(guides))
	}

	_, text, _ = toolCall(t, ts.URL, "list_entries", map[string]any{"domain": "ml"})
	var ml []db.Entry
	json.Unmarshal([]byte(text), &ml)
	if len(ml) != 1 {
		t.Errorf("ml: %d", len(ml))
	}
}

func TestSearchEntries(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "s1", "Go Error Handling", "Always check error values returned by functions in Go.", "skill", "go", "", "", []string{"go", "errors"})
	createEntry(t, ts.URL, "s2", "Rust Ownership", "Rust uses ownership and borrowing for memory management.", "skill", "rust", "", "", []string{"rust"})
	createEntry(t, ts.URL, "s3", "Python Error Handling", "Python uses try/except for error handling and recovery.", "skill", "python", "", "", []string{"python", "errors"})

	_, text, _ := toolCall(t, ts.URL, "search_entries", map[string]any{"query": "error"})
	var results []db.Entry
	json.Unmarshal([]byte(text), &results)
	if len(results) < 2 {
		t.Errorf("expected >=2 results for 'error', got %d", len(results))
	}
	for _, r := range results {
		if r.Content != "" {
			t.Errorf("search should not return content for %s", r.Slug)
		}
	}

	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "error", "language": "go"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "s1" {
		t.Errorf("expected 1 Go result, got %d", len(results))
	}

	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "ownership borrowing"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "s2" {
		t.Errorf("expected s2 for 'ownership', got %v", results)
	}
}

func TestGetEntriesByContext(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "ctx1", "Go HTTP", "Build HTTP servers", "", "go", "backend", "", []string{"http", "server"})
	createEntry(t, ts.URL, "ctx2", "Go Testing", "Testing basics", "", "go", "backend", "", []string{"testing"})
	createEntry(t, ts.URL, "ctx3", "Rust CLI", "CLI in Rust", "", "rust", "cli", "", []string{"cli"})

	_, text, _ := toolCall(t, ts.URL, "get_entries_by_context", map[string]any{"language": "go", "domain": "backend"})
	var results []db.Entry
	json.Unmarshal([]byte(text), &results)
	if len(results) != 2 {
		t.Errorf("expected 2 go/backend, got %d", len(results))
	}
	for _, r := range results {
		if r.Content == "" {
			t.Errorf("context should include content for %s", r.Slug)
		}
	}

	_, text, _ = toolCall(t, ts.URL, "get_entries_by_context", map[string]any{"tags": []string{"http"}})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "ctx1" {
		t.Errorf("expected ctx1 for tag 'http', got %v", results)
	}
}

func TestListTags(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "lt1", "T1", "c1", "", "", "", "", []string{"alpha", "beta"})
	createEntry(t, ts.URL, "lt2", "T2", "c2", "", "", "", "", []string{"beta", "gamma"})
	createEntry(t, ts.URL, "lt3", "T3", "c3", "", "", "", "", []string{"beta"})

	_, text, _ := toolCall(t, ts.URL, "list_tags", map[string]any{})
	var tags []db.Tag
	json.Unmarshal([]byte(text), &tags)
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
	if tags[0].Name != "beta" || tags[0].Count != 3 {
		t.Errorf("first tag: %+v", tags[0])
	}
}

func TestFTS5Sync(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "fts", "Animal Guide", "The monotreme is a unique mammal.", "", "", "", "", nil)

	_, text, _ := toolCall(t, ts.URL, "search_entries", map[string]any{"query": "monotreme"})
	var results []db.Entry
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 {
		t.Errorf("after create: expected 1, got %d", len(results))
	}

	toolCall(t, ts.URL, "update_entry", map[string]any{"slug": "fts", "content": "The echidna is a spiny anteater."})

	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "monotreme"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 0 {
		t.Errorf("after update: expected 0 for old word, got %d", len(results))
	}

	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "echidna"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 {
		t.Errorf("after update: expected 1 for new word, got %d", len(results))
	}

	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "animal guide"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 {
		t.Errorf("title search: expected 1, got %d", len(results))
	}

	toolCall(t, ts.URL, "delete_entry", map[string]any{"slug": "fts"})

	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "echidna"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 0 {
		t.Errorf("after delete: expected 0, got %d", len(results))
	}
}

func TestStats(t *testing.T) {
	s, ts := setup(t)
	createEntry(t, ts.URL, "st", "Stats Entry", "unique stats content here", "", "", "", "", nil)

	for i := 0; i < 3; i++ {
		toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "st"})
	}

	stats, err := s.DB.GetStats(context.Background(), "st")
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Reads < 3 {
		t.Errorf("reads: got %d, want >= 3", stats.Reads)
	}
	if stats.LastReadAt == nil {
		t.Error("last_read_at should be set")
	}

	toolCall(t, ts.URL, "search_entries", map[string]any{"query": "unique stats content"})
	stats, _ = s.DB.GetStats(context.Background(), "st")
	if stats.Searches < 1 {
		t.Errorf("searches: got %d, want >= 1", stats.Searches)
	}

	toolCall(t, ts.URL, "update_entry", map[string]any{"slug": "st", "title": "Updated"})
	stats, _ = s.DB.GetStats(context.Background(), "st")
	if stats.Updates < 1 {
		t.Errorf("updates: got %d, want >= 1", stats.Updates)
	}
}

func TestWriteWhenLocked(t *testing.T) {
	s, ts := setup(t)
	createEntry(t, ts.URL, "exists", "Exists", "content", "", "", "", "", nil)

	if err := s.DB.Lock(context.Background(), "secret"); err != nil {
		t.Fatalf("lock: %v", err)
	}

	_, text, isErr := toolCall(t, ts.URL, "create_entry", map[string]any{
		"slug": "blocked", "title": "Blocked", "content": "nope",
	})
	if !isErr {
		t.Fatal("expected error for create when locked")
	}
	if !strings.Contains(text, "locked") {
		t.Errorf("expected 'locked' in error, got: %s", text)
	}

	_, _, isErr = toolCall(t, ts.URL, "update_entry", map[string]any{"slug": "exists", "title": "Nope"})
	if !isErr {
		t.Fatal("expected error for update when locked")
	}

	_, _, isErr = toolCall(t, ts.URL, "delete_entry", map[string]any{"slug": "exists"})
	if !isErr {
		t.Fatal("expected error for delete when locked")
	}

	_, _, isErr = toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "exists"})
	if isErr {
		t.Fatal("get_entry should work when locked")
	}

	if err := s.DB.Unlock(context.Background(), "wrong"); err == nil {
		t.Fatal("expected error for wrong token")
	}
	if err := s.DB.Unlock(context.Background(), "secret"); err != nil {
		t.Fatalf("unlock: %v", err)
	}

	_, _, isErr = toolCall(t, ts.URL, "update_entry", map[string]any{"slug": "exists", "title": "Now Works"})
	if isErr {
		t.Fatal("update should work after unlock")
	}
}

func TestLockEdgeCases(t *testing.T) {
	s, _ := setup(t)

	if err := s.DB.Lock(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty token")
	}
	if err := s.DB.Lock(context.Background(), "tok"); err != nil {
		t.Fatalf("lock: %v", err)
	}
	if err := s.DB.Lock(context.Background(), "tok2"); err == nil {
		t.Fatal("expected error for double lock")
	}

	s.DB.Unlock(context.Background(), "tok")
	if err := s.DB.Unlock(context.Background(), "tok"); err == nil {
		t.Fatal("expected error for unlocking when not locked")
	}
}

func TestResourcesList(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "r1", "Res One", "content 1", "", "", "", "", nil)
	createEntry(t, ts.URL, "r2", "Res Two", "content 2", "", "", "", "", nil)

	_, resp := call(t, ts.URL, "resources/list", 1, nil, nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	resources := resp.Result.(map[string]any)["resources"].([]any)
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
	r0 := resources[0].(map[string]any)
	if r0["mimeType"] != "text/markdown" {
		t.Errorf("mimeType: %v", r0["mimeType"])
	}
	if !strings.HasPrefix(r0["uri"].(string), "mcpedia://entries/") {
		t.Errorf("bad uri: %s", r0["uri"])
	}
}

func TestResourcesListExcludesHowToUse(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "aaa", "AAA", "content", "", "", "", "", nil)
	args := map[string]any{"slug": "how-to-use", "title": "How-To", "content": "Guide", "description": "Guide"}
	toolCall(t, ts.URL, "create_entry", args)

	// how-to-use is excluded from list; accessed via mcpedia://how-to-use only
	_, resp := call(t, ts.URL, "resources/list", 1, nil, nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	resources := resp.Result.(map[string]any)["resources"].([]any)
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource (how-to-use excluded), got %d", len(resources))
	}
	if resources[0].(map[string]any)["uri"] == "mcpedia://entries/how-to-use" {
		t.Error("how-to-use must not appear in resources list")
	}
}

func TestResourcesReadHowToUseURI(t *testing.T) {
	_, ts := setup(t)
	// mcpedia://how-to-use returns default when not in DB
	_, resp := call(t, ts.URL, "resources/read", 1, map[string]any{"uri": "mcpedia://how-to-use"}, nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	contents := resp.Result.(map[string]any)["contents"].([]any)
	c0 := contents[0].(map[string]any)
	text := c0["text"].(string)
	if !strings.Contains(text, "How to Use MCPedia") || !strings.Contains(text, "search_entries") {
		t.Errorf("expected default how-to-use content; got %s", text[:min(200, len(text))])
	}

	// User's how-to-use overrides default
	toolCall(t, ts.URL, "create_entry", map[string]any{"slug": "how-to-use", "title": "Custom", "content": "My custom guide.", "description": "Custom"})
	_, resp = call(t, ts.URL, "resources/read", 2, map[string]any{"uri": "mcpedia://how-to-use"}, nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	contents = resp.Result.(map[string]any)["contents"].([]any)
	text = contents[0].(map[string]any)["text"].(string)
	if text != "My custom guide." {
		t.Errorf("expected user's content; got %s", text)
	}
}

func TestGetEntryDefaultHowToUse(t *testing.T) {
	_, ts := setup(t)
	_, text, isErr := toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "how-to-use"})
	if isErr {
		t.Fatalf("get_entry how-to-use should return default: %s", text)
	}
	var e db.Entry
	json.Unmarshal([]byte(text), &e)
	if e.Slug != "how-to-use" || e.Title != "How to Use MCPedia" {
		t.Errorf("expected default entry; got slug=%q title=%q", e.Slug, e.Title)
	}
	if !strings.Contains(e.Content, "search_entries") {
		t.Errorf("expected default content with tools; got %s", e.Content[:min(200, len(e.Content))])
	}
}

func TestResourcesRead(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "readme", "Read Me", "This is the content.", "", "", "", "", nil)

	_, resp := call(t, ts.URL, "resources/read", 1, map[string]any{"uri": "mcpedia://entries/readme"}, nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	contents := resp.Result.(map[string]any)["contents"].([]any)
	c0 := contents[0].(map[string]any)
	if c0["text"] != "This is the content." {
		t.Errorf("text: %v", c0["text"])
	}
	if c0["mimeType"] != "text/markdown" {
		t.Errorf("mimeType: %v", c0["mimeType"])
	}

	_, resp = call(t, ts.URL, "resources/read", 2, map[string]any{"uri": "bogus://x"}, nil)
	if resp.Error == nil {
		t.Error("expected error for invalid URI")
	}

	_, resp = call(t, ts.URL, "resources/read", 3, map[string]any{"uri": "mcpedia://entries/nope"}, nil)
	if resp.Error == nil {
		t.Error("expected error for non-existent resource")
	}
}

func TestResourcesTemplatesList(t *testing.T) {
	_, ts := setup(t)
	_, resp := call(t, ts.URL, "resources/templates/list", 1, nil, nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	templates := resp.Result.(map[string]any)["resourceTemplates"].([]any)
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
	if templates[0].(map[string]any)["uriTemplate"] != "mcpedia://how-to-use" {
		t.Errorf("first template should be how-to-use; got %v", templates[0].(map[string]any)["uriTemplate"])
	}
	if templates[1].(map[string]any)["uriTemplate"] != "mcpedia://entries/{slug}" {
		t.Errorf("second template should be entries; got %v", templates[1].(map[string]any)["uriTemplate"])
	}
}

func TestPromptsList(t *testing.T) {
	_, ts := setup(t)
	_, resp := call(t, ts.URL, "prompts/list", 1, nil, nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	prompts := resp.Result.(map[string]any)["prompts"].([]any)
	if len(prompts) != 3 {
		t.Errorf("expected 3 prompts, got %d", len(prompts))
	}
	names := map[string]bool{}
	for _, p := range prompts {
		names[p.(map[string]any)["name"].(string)] = true
	}
	for _, want := range []string{"apply-entry", "review-with-entry", "save-learnings"} {
		if !names[want] {
			t.Errorf("missing prompt: %s", want)
		}
	}
}

func TestPromptsGet(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "ptest", "Prompt Test", "Follow these rules carefully.", "", "", "", "", nil)

	_, resp := call(t, ts.URL, "prompts/get", 1, map[string]any{
		"name": "apply-entry", "arguments": map[string]string{"slug": "ptest"},
	}, nil)
	if resp.Error != nil {
		t.Fatalf("apply-entry error: %+v", resp.Error)
	}
	text := resp.Result.(map[string]any)["messages"].([]any)[0].(map[string]any)["content"].(map[string]any)["text"].(string)
	if !strings.Contains(text, "Follow these rules carefully") {
		t.Error("apply-entry should contain entry content")
	}
	if !strings.Contains(text, "Prompt Test") {
		t.Error("apply-entry should contain entry title")
	}

	_, resp = call(t, ts.URL, "prompts/get", 2, map[string]any{
		"name": "review-with-entry", "arguments": map[string]string{"slug": "ptest"},
	}, nil)
	if resp.Error != nil {
		t.Fatalf("review error: %+v", resp.Error)
	}

	_, resp = call(t, ts.URL, "prompts/get", 3, map[string]any{
		"name": "save-learnings", "arguments": map[string]string{},
	}, nil)
	if resp.Error != nil {
		t.Fatalf("save-learnings error: %+v", resp.Error)
	}
	text = resp.Result.(map[string]any)["messages"].([]any)[0].(map[string]any)["content"].(map[string]any)["text"].(string)
	if !strings.Contains(text, "create_entry") {
		t.Error("save-learnings should reference create_entry")
	}

	_, resp = call(t, ts.URL, "prompts/get", 4, map[string]any{"name": "nonexistent", "arguments": map[string]string{}}, nil)
	if resp.Error == nil {
		t.Error("expected error for unknown prompt")
	}

	_, resp = call(t, ts.URL, "prompts/get", 5, map[string]any{"name": "apply-entry", "arguments": map[string]string{}}, nil)
	if resp.Error == nil {
		t.Error("expected error for missing slug")
	}
}

func TestPromptsGetNonExistentEntry(t *testing.T) {
	_, ts := setup(t)

	_, resp := call(t, ts.URL, "prompts/get", 1, map[string]any{"name": "apply-entry", "arguments": map[string]string{"slug": "nope"}}, nil)
	if resp.Error == nil {
		t.Error("expected error for non-existent entry in apply-entry")
	}

	_, resp = call(t, ts.URL, "prompts/get", 2, map[string]any{"name": "review-with-entry", "arguments": map[string]string{"slug": "nope"}}, nil)
	if resp.Error == nil {
		t.Error("expected error for non-existent entry in review-with-entry")
	}

	_, resp = call(t, ts.URL, "prompts/get", 3, map[string]any{"name": "review-with-entry", "arguments": map[string]string{}}, nil)
	if resp.Error == nil {
		t.Error("expected error for missing slug in review-with-entry")
	}
}

func TestAuth(t *testing.T) {
	_, ts := setupWithToken(t, "supersecret")

	req, _ := http.NewRequest("POST", ts.URL, bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Errorf("no auth: expected 401, got %d", resp.StatusCode)
	}

	req, _ = http.NewRequest("POST", ts.URL, bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong")
	resp, _ = http.DefaultClient.Do(req)
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Errorf("wrong token: expected 401, got %d", resp.StatusCode)
	}

	status, r := call(t, ts.URL, "ping", 1, nil, map[string]string{"Authorization": "Bearer supersecret"})
	if status != 200 || r.Error != nil {
		t.Errorf("valid auth failed: status=%d, error=%v", status, r.Error)
	}
}

func TestHTTPEdgeCases(t *testing.T) {
	_, ts := setup(t)

	resp, _ := http.Get(ts.URL)
	resp.Body.Close()
	if resp.StatusCode != 405 {
		t.Errorf("GET: expected 405, got %d", resp.StatusCode)
	}

	req, _ := http.NewRequest("POST", ts.URL, bytes.NewReader([]byte(`not json`)))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = http.DefaultClient.Do(req)
	var errResp jsonrpcResponse
	json.NewDecoder(resp.Body).Decode(&errResp)
	resp.Body.Close()
	if errResp.Error == nil || errResp.Error.Code != -32700 {
		t.Errorf("invalid JSON: expected parse error, got %+v", errResp.Error)
	}

	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	req, _ = http.NewRequest("POST", ts.URL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = http.DefaultClient.Do(req)
	resp.Body.Close()
	if resp.StatusCode != 202 {
		t.Errorf("notification: expected 202, got %d", resp.StatusCode)
	}

	_, r := call(t, ts.URL, "bogus/method", 1, nil, nil)
	if r.Error == nil || r.Error.Code != -32601 {
		t.Errorf("unknown method: expected -32601, got %+v", r.Error)
	}

	_, r = call(t, ts.URL, "tools/call", 1, map[string]any{"name": "bogus_tool", "arguments": map[string]any{}}, nil)
	if r.Error == nil && r.Result != nil {
		result := r.Result.(map[string]any)
		if result["isError"] == nil {
			t.Error("expected error for unknown tool")
		}
	}

	_, _, isErr := toolCall(t, ts.URL, "search_entries", map[string]any{})
	if !isErr {
		t.Error("expected error for search without query")
	}
	_, _, isErr = toolCall(t, ts.URL, "get_entry", map[string]any{})
	if !isErr {
		t.Error("expected error for get without slug")
	}
	_, _, isErr = toolCall(t, ts.URL, "delete_entry", map[string]any{})
	if !isErr {
		t.Error("expected error for delete without slug")
	}
	_, _, isErr = toolCall(t, ts.URL, "update_entry", map[string]any{})
	if !isErr {
		t.Error("expected error for update without slug")
	}
}

func TestDefaultKind(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "dk", "Default Kind", "content", "", "", "", "", nil)
	_, text, _ := toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "dk"})
	var e db.Entry
	json.Unmarshal([]byte(text), &e)
	if e.Kind != "skill" {
		t.Errorf("default kind: got %q, want 'skill'", e.Kind)
	}
}

func TestSearchWithLimit(t *testing.T) {
	_, ts := setup(t)
	for i := 0; i < 5; i++ {
		slug := fmt.Sprintf("lim-%d", i)
		createEntry(t, ts.URL, slug, "Limit Test "+slug, "shared keyword searchable content", "", "", "", "", nil)
	}

	_, text, _ := toolCall(t, ts.URL, "search_entries", map[string]any{"query": "searchable", "limit": 2})
	var results []db.Entry
	json.Unmarshal([]byte(text), &results)
	if len(results) != 2 {
		t.Errorf("expected 2 with limit, got %d", len(results))
	}
}

func TestGetEntriesByContextWithLimit(t *testing.T) {
	_, ts := setup(t)
	for i := 0; i < 5; i++ {
		slug := fmt.Sprintf("ctx-lim-%d", i)
		createEntry(t, ts.URL, slug, "Ctx "+slug, "content", "", "go", "", "", nil)
	}

	_, text, _ := toolCall(t, ts.URL, "get_entries_by_context", map[string]any{"language": "go", "limit": 3})
	var results []db.Entry
	json.Unmarshal([]byte(text), &results)
	if len(results) != 3 {
		t.Errorf("expected 3 with limit, got %d", len(results))
	}
}

func TestResourcesPagination(t *testing.T) {
	_, ts := setup(t)
	for i := 0; i < 55; i++ {
		slug := fmt.Sprintf("pg-%03d", i)
		createEntry(t, ts.URL, slug, "Page "+slug, "content", "", "", "", "", nil)
	}
	// Total: 55 (how-to-use excluded from list)

	_, resp := call(t, ts.URL, "resources/list", 1, nil, nil)
	result := resp.Result.(map[string]any)
	resources := result["resources"].([]any)
	if len(resources) != 50 {
		t.Errorf("first page: expected 50, got %d", len(resources))
	}
	nextCursor, hasCursor := result["nextCursor"].(string)
	if !hasCursor || nextCursor == "" {
		t.Fatal("expected nextCursor for first page")
	}

	_, resp = call(t, ts.URL, "resources/list", 2, map[string]any{"cursor": nextCursor}, nil)
	result = resp.Result.(map[string]any)
	resources = result["resources"].([]any)
	if len(resources) != 5 {
		t.Errorf("second page: expected 5, got %d", len(resources))
	}
	if _, has := result["nextCursor"]; has {
		t.Error("should not have nextCursor on last page")
	}
}

func TestAllEntries(t *testing.T) {
	s, ts := setup(t)
	createEntry(t, ts.URL, "all1", "All One", "content one", "skill", "go", "", "", []string{"x"})
	createEntry(t, ts.URL, "all2", "All Two", "content two", "rule", "rust", "", "", []string{"y"})

	entries, err := s.DB.AllEntries(context.Background())
	if err != nil {
		t.Fatalf("all entries: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Content == "" {
			t.Errorf("expected content for %s", e.Slug)
		}
		if len(e.Tags) == 0 {
			t.Errorf("expected tags for %s", e.Slug)
		}
	}
}

func TestSearchWithFilters(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "stf1", "Tagged Search A", "searchable content alpha", "", "", "", "", []string{"tagged"})
	createEntry(t, ts.URL, "stf2", "Tagged Search B", "searchable content beta", "", "", "", "", []string{"other"})
	createEntry(t, ts.URL, "df1", "Domain Filter A", "searchable domain content alpha", "", "", "backend", "", nil)
	createEntry(t, ts.URL, "df2", "Domain Filter B", "searchable domain content beta", "", "", "frontend", "", nil)
	createEntry(t, ts.URL, "kf1", "Kind Filter A", "searchable kind content here", "rule", "", "", "", nil)
	createEntry(t, ts.URL, "kf2", "Kind Filter B", "searchable kind content here", "guide", "", "", "", nil)
	createEntry(t, ts.URL, "sp1", "Project Search A", "findable content alpha", "", "", "", "proj-a", nil)
	createEntry(t, ts.URL, "sp2", "Project Search B", "findable content beta", "", "", "", "proj-b", nil)

	// By tag
	_, text, _ := toolCall(t, ts.URL, "search_entries", map[string]any{"query": "searchable content", "tag": "tagged"})
	var results []db.Entry
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "stf1" {
		t.Errorf("tag filter: expected stf1, got %v", results)
	}

	// By domain
	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "searchable domain", "domain": "backend"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "df1" {
		t.Errorf("domain filter: expected df1, got %v", results)
	}

	// By kind
	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "searchable kind", "kind": "rule"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "kf1" {
		t.Errorf("kind filter: expected kf1, got %v", results)
	}

	// By project
	_, text, _ = toolCall(t, ts.URL, "search_entries", map[string]any{"query": "findable", "project": "proj-a"})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "sp1" {
		t.Errorf("project filter: expected sp1, got %v", results)
	}
}

func TestGetEntriesByContextVariations(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "ckp1", "CKP One", "c1", "rule", "go", "", "proj-x", nil)
	createEntry(t, ts.URL, "ckp2", "CKP Two", "c2", "skill", "go", "", "proj-x", nil)
	createEntry(t, ts.URL, "ckp3", "CKP Three", "c3", "rule", "go", "", "proj-y", nil)
	createEntry(t, ts.URL, "ct1", "CT One", "c1", "", "", "", "", []string{"special"})
	createEntry(t, ts.URL, "ct2", "CT Two", "c2", "", "", "", "", []string{"other"})

	_, text, _ := toolCall(t, ts.URL, "get_entries_by_context", map[string]any{"kind": "rule", "project": "proj-x"})
	var results []db.Entry
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "ckp1" {
		t.Errorf("kind+project: expected ckp1, got %v", results)
	}

	_, text, _ = toolCall(t, ts.URL, "get_entries_by_context", map[string]any{"tags": []string{"special"}})
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "ct1" {
		t.Errorf("tags: expected ct1, got %v", results)
	}
}

func TestListEntriesWithProjectFilter(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "lp1", "LP One", "c", "", "", "", "alpha", nil)
	createEntry(t, ts.URL, "lp2", "LP Two", "c", "", "", "", "beta", nil)

	_, text, _ := toolCall(t, ts.URL, "list_entries", map[string]any{"project": "alpha"})
	var results []db.Entry
	json.Unmarshal([]byte(text), &results)
	if len(results) != 1 || results[0].Slug != "lp1" {
		t.Errorf("expected lp1, got %v", results)
	}
}

func TestGetStatsNotFound(t *testing.T) {
	s, _ := setup(t)
	_, err := s.DB.GetStats(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent entry stats")
	}
}

func TestCreateEntryWithEmptyTags(t *testing.T) {
	_, ts := setup(t)
	createEntry(t, ts.URL, "et", "Empty Tags", "content", "", "", "", "", []string{"valid", "", "  "})
	_, text, _ := toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "et"})
	var e db.Entry
	json.Unmarshal([]byte(text), &e)
	if len(e.Tags) != 1 || e.Tags[0] != "valid" {
		t.Errorf("expected [valid], got %v", e.Tags)
	}
}

func TestListTagsEmpty(t *testing.T) {
	_, ts := setup(t)
	_, text, isErr := toolCall(t, ts.URL, "list_tags", map[string]any{})
	if isErr {
		t.Fatal("list_tags failed")
	}
	if text != "null" && text != "[]" {
		var tags []db.Tag
		json.Unmarshal([]byte(text), &tags)
		if len(tags) != 0 {
			t.Errorf("expected empty tags, got %v", tags)
		}
	}
}

func TestGetEntryNotFound(t *testing.T) {
	_, ts := setup(t)
	_, text, isErr := toolCall(t, ts.URL, "get_entry", map[string]any{"slug": "does-not-exist"})
	if !isErr {
		t.Fatal("expected error")
	}
	if !strings.Contains(text, "not found") {
		t.Errorf("expected 'not found', got: %s", text)
	}
}

func TestReopenDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reopen.db")

	d1, err := db.Open(path)
	if err != nil {
		t.Fatalf("open1: %v", err)
	}
	d1.CreateEntry(context.Background(), &db.Entry{Slug: "persist", Title: "Persist", Content: "persisted"})
	d1.Close()

	d2, err := db.Open(path)
	if err != nil {
		t.Fatalf("open2: %v", err)
	}
	defer d2.Close()
	got, err := d2.GetEntry(context.Background(), "persist")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Persist" {
		t.Errorf("title: %q", got.Title)
	}
}
