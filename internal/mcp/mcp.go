package mcp

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pouriya/mcpedia/internal/db"
)

const (
	protocolVersion  = "2025-11-25"
	serverName       = "mcpedia"
	serverVersion    = "0.1.0"
	resourcesPerPage = 50
)

// Server implements the MCP protocol over HTTP.
type Server struct {
	DB       *db.DB
	Token    string // empty = no auth required
	sessions sync.Map
}

// --- response writer wrapper ---

type responseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
	rpcMethod   string
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// --- JSON-RPC types ---

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

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

func rpcResult(id any, result any) *jsonrpcResponse {
	return &jsonrpcResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func rpcErr(id any, code int, msg string) *jsonrpcResponse {
	return &jsonrpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
}

// --- HTTP handler ---

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

	s.serveRequest(rw, r)

	duration := time.Since(start)

	if slog.Default().Enabled(r.Context(), slog.LevelDebug) {
		slog.Debug("http request detail",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", duration.Milliseconds(),
			"rpc_method", rw.rpcMethod,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"response_bytes", rw.bytes,
		)
	} else {
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", duration.Milliseconds(),
			"rpc_method", rw.rpcMethod,
		)
	}
}

func (s *Server) serveRequest(w *responseWriter, r *http.Request) {
	// Only accept POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Auth check
	if s.Token != "" {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != s.Token {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var req jsonrpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusOK, rpcErr(nil, -32700, "Parse error"))
		return
	}

	w.rpcMethod = req.Method

	if req.JSONRPC != "2.0" {
		writeJSON(w, http.StatusOK, rpcErr(req.ID, -32600, "Invalid request: jsonrpc must be 2.0"))
		return
	}

	// Notifications (no ID) get 202 Accepted
	if req.ID == nil {
		s.handleNotification(req)
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// Session validation for non-initialize requests
	if req.Method != "initialize" {
		sessionID := r.Header.Get("Mcp-Session-Id")
		if sessionID != "" {
			if _, ok := s.sessions.Load(sessionID); !ok {
				writeJSON(w, http.StatusOK, rpcErr(req.ID, -32600, "Invalid session"))
				return
			}
		}
	}

	resp := s.dispatch(req)

	// For initialize, set session header
	if req.Method == "initialize" && resp.Error == nil {
		if result, ok := resp.Result.(map[string]any); ok {
			if sid, ok := result["_sessionId"].(string); ok {
				w.Header().Set("Mcp-Session-Id", sid)
				delete(result, "_sessionId")
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleNotification(req jsonrpcRequest) {
	// notifications/initialized -- nothing to do
	// notifications/cancelled -- nothing to do
}

func (s *Server) dispatch(req jsonrpcRequest) *jsonrpcResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "ping":
		return rpcResult(req.ID, map[string]any{})
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(req)
	case "resources/templates/list":
		return s.handleResourcesTemplatesList(req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(req)
	default:
		return rpcErr(req.ID, -32601, "Method not found: "+req.Method)
	}
}

// --- Initialize ---

func (s *Server) handleInitialize(req jsonrpcRequest) *jsonrpcResponse {
	sessionID := generateSessionID()
	s.sessions.Store(sessionID, true)

	result := map[string]any{
		"protocolVersion": protocolVersion,
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"resources": map[string]any{},
			"prompts":   map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    serverName,
			"version": serverVersion,
		},
		"_sessionId": sessionID, // stripped by ServeHTTP and set as header
	}
	return rpcResult(req.ID, result)
}

// --- Tools ---

func (s *Server) handleToolsList(req jsonrpcRequest) *jsonrpcResponse {
	tools := toolDefinitions()
	slog.Info("tool call", "tool", "list", "items", len(tools))
	return rpcResult(req.ID, map[string]any{"tools": tools})
}

func (s *Server) handleToolsCall(req jsonrpcRequest) *jsonrpcResponse {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return rpcErr(req.ID, -32602, "Invalid params: "+err.Error())
	}

	switch params.Name {
	case "search_entries":
		return s.toolSearchEntries(req.ID, params.Arguments)
	case "get_entry":
		return s.toolGetEntry(req.ID, params.Arguments)
	case "get_entries_by_context":
		return s.toolGetEntriesByContext(req.ID, params.Arguments)
	case "list_entries":
		return s.toolListEntries(req.ID, params.Arguments)
	case "list_tags":
		return s.toolListTags(req.ID)
	case "create_entry":
		return s.toolCreateEntry(req.ID, params.Arguments)
	case "update_entry":
		return s.toolUpdateEntry(req.ID, params.Arguments)
	case "delete_entry":
		return s.toolDeleteEntry(req.ID, params.Arguments)
	default:
		return rpcErr(req.ID, -32602, "Unknown tool: "+params.Name)
	}
}

func (s *Server) checkLock() error {
	locked, err := s.DB.IsLocked()
	if err != nil {
		return err
	}
	if locked {
		return fmt.Errorf("database is locked. Write operations are disabled")
	}
	return nil
}

func (s *Server) toolSearchEntries(id any, args map[string]any) *jsonrpcResponse {
	query := str(args, "query")
	if query == "" {
		return toolError(id, "query is required")
	}
	limit := intVal(args, "limit", 10)
	f := db.Filter{
		Kind:     str(args, "kind"),
		Language: str(args, "language"),
		Domain:   str(args, "domain"),
		Project:  str(args, "project"),
		Tag:      str(args, "tag"),
	}
	entries, err := s.DB.SearchEntries(query, f, limit)
	if err != nil {
		return toolError(id, err.Error())
	}
	slog.Info("tool call", "tool", "search_entries", "query", query, "items", len(entries))
	return toolResult(id, entries)
}

func (s *Server) toolGetEntry(id any, args map[string]any) *jsonrpcResponse {
	slug := str(args, "slug")
	if slug == "" {
		return toolError(id, "slug is required")
	}
	entry, err := s.DB.GetEntry(slug)
	if err != nil {
		return toolError(id, err.Error())
	}
	slog.Info("tool call", "tool", "get_entry", "slug", slug)
	return toolResult(id, entry)
}

func (s *Server) toolGetEntriesByContext(id any, args map[string]any) *jsonrpcResponse {
	limit := intVal(args, "limit", 20)
	f := db.Filter{
		Kind:     str(args, "kind"),
		Language: str(args, "language"),
		Domain:   str(args, "domain"),
		Project:  str(args, "project"),
		Tags:     strSlice(args, "tags"),
	}
	entries, err := s.DB.GetEntriesByContext(f, limit)
	if err != nil {
		return toolError(id, err.Error())
	}
	slog.Info("tool call", "tool", "get_entries_by_context", "items", len(entries))
	return toolResult(id, entries)
}

func (s *Server) toolListEntries(id any, args map[string]any) *jsonrpcResponse {
	f := db.Filter{
		Kind:     str(args, "kind"),
		Language: str(args, "language"),
		Domain:   str(args, "domain"),
		Project:  str(args, "project"),
	}
	entries, err := s.DB.ListEntries(f)
	if err != nil {
		return toolError(id, err.Error())
	}
	slog.Info("tool call", "tool", "list_entries", "items", len(entries))
	return toolResult(id, entries)
}

func (s *Server) toolListTags(id any) *jsonrpcResponse {
	tags, err := s.DB.ListTags()
	if err != nil {
		return toolError(id, err.Error())
	}
	slog.Info("tool call", "tool", "list_tags", "items", len(tags))
	return toolResult(id, tags)
}

func (s *Server) toolCreateEntry(id any, args map[string]any) *jsonrpcResponse {
	if err := s.checkLock(); err != nil {
		return toolError(id, err.Error())
	}
	slug := str(args, "slug")
	title := str(args, "title")
	content := str(args, "content")
	if slug == "" || title == "" || content == "" {
		return toolError(id, "slug, title, and content are required")
	}
	e := &db.Entry{
		Slug:        slug,
		Title:       title,
		Description: str(args, "description"),
		Content:     content,
		Kind:        str(args, "kind"),
		Language:    str(args, "language"),
		Domain:      str(args, "domain"),
		Project:     str(args, "project"),
		Tags:        strSlice(args, "tags"),
	}
	if err := s.DB.CreateEntry(e); err != nil {
		return toolError(id, err.Error())
	}
	slog.Info("tool call", "tool", "create_entry", "slug", slug)
	return toolResult(id, e)
}

func (s *Server) toolUpdateEntry(id any, args map[string]any) *jsonrpcResponse {
	if err := s.checkLock(); err != nil {
		return toolError(id, err.Error())
	}
	slug := str(args, "slug")
	if slug == "" {
		return toolError(id, "slug is required")
	}
	fields := map[string]any{}
	for _, key := range []string{"title", "description", "content", "kind", "language", "domain", "project"} {
		if v, ok := args[key]; ok {
			fields[key] = v
		}
	}
	if v, ok := args["tags"]; ok {
		fields["tags"] = v
	}
	if err := s.DB.UpdateEntry(slug, fields); err != nil {
		return toolError(id, err.Error())
	}
	// Return the updated entry
	entry, err := s.DB.GetEntry(slug)
	if err != nil {
		return toolError(id, err.Error())
	}
	slog.Info("tool call", "tool", "update_entry", "slug", slug)
	return toolResult(id, entry)
}

func (s *Server) toolDeleteEntry(id any, args map[string]any) *jsonrpcResponse {
	if err := s.checkLock(); err != nil {
		return toolError(id, err.Error())
	}
	slug := str(args, "slug")
	if slug == "" {
		return toolError(id, "slug is required")
	}
	if err := s.DB.DeleteEntry(slug); err != nil {
		return toolError(id, err.Error())
	}
	slog.Info("tool call", "tool", "delete_entry", "slug", slug)
	return toolResult(id, map[string]string{"deleted": slug})
}

// --- Resources ---

func (s *Server) handleResourcesList(req jsonrpcRequest) *jsonrpcResponse {
	var params struct {
		Cursor string `json:"cursor"`
	}
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	offset := 0
	if params.Cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(params.Cursor)
		if err == nil {
			offset, _ = strconv.Atoi(string(decoded))
		}
	}

	entries, err := s.DB.ListEntries(db.Filter{})
	if err != nil {
		return rpcErr(req.ID, -32603, err.Error())
	}

	// Apply pagination
	end := offset + resourcesPerPage
	if end > len(entries) {
		end = len(entries)
	}
	page := entries
	if offset < len(entries) {
		page = entries[offset:end]
	} else {
		page = nil
	}

	resources := make([]map[string]any, 0, len(page))
	for _, e := range page {
		resources = append(resources, map[string]any{
			"uri":         "mcpedia://entries/" + e.Slug,
			"name":        e.Slug,
			"title":       e.Title,
			"description": e.Description,
			"mimeType":    "text/markdown",
		})
	}

	result := map[string]any{"resources": resources}
	if end < len(entries) {
		result["nextCursor"] = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(end)))
	}
	slog.Info("resource call", "resource", "list", "items", len(resources), "total", len(entries))
	return rpcResult(req.ID, result)
}

func (s *Server) handleResourcesRead(req jsonrpcRequest) *jsonrpcResponse {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return rpcErr(req.ID, -32602, "Invalid params")
	}

	slug := strings.TrimPrefix(params.URI, "mcpedia://entries/")
	if slug == "" || slug == params.URI {
		return rpcErr(req.ID, -32002, "Invalid resource URI: "+params.URI)
	}

	entry, err := s.DB.GetEntry(slug)
	if err != nil {
		return rpcErr(req.ID, -32002, err.Error())
	}

	slog.Info("resource call", "resource", "read", "slug", slug)
	return rpcResult(req.ID, map[string]any{
		"contents": []map[string]any{
			{
				"uri":      params.URI,
				"mimeType": "text/markdown",
				"text":     entry.Content,
			},
		},
	})
}

func (s *Server) handleResourcesTemplatesList(req jsonrpcRequest) *jsonrpcResponse {
	slog.Info("resource call", "resource", "templates_list", "items", 1)
	return rpcResult(req.ID, map[string]any{
		"resourceTemplates": []map[string]any{
			{
				"uriTemplate": "mcpedia://entries/{slug}",
				"name":        "MCPedia Entry",
				"description": "Access a knowledge entry by its slug",
				"mimeType":    "text/markdown",
			},
		},
	})
}

// --- Prompts ---

func (s *Server) handlePromptsList(req jsonrpcRequest) *jsonrpcResponse {
	slog.Info("prompt call", "prompt", "list", "items", 3)
	return rpcResult(req.ID, map[string]any{
		"prompts": []map[string]any{
			{
				"name":        "apply-entry",
				"title":       "Apply Entry",
				"description": "Apply a knowledge entry's guidelines to the current task",
				"arguments": []map[string]any{
					{"name": "slug", "description": "The slug of the entry to apply", "required": true},
				},
			},
			{
				"name":        "review-with-entry",
				"title":       "Review With Entry",
				"description": "Review code against a knowledge entry's guidelines",
				"arguments": []map[string]any{
					{"name": "slug", "description": "The slug of the entry to review against", "required": true},
				},
			},
			{
				"name":        "save-learnings",
				"title":       "Save Learnings",
				"description": "Extract and save reusable knowledge from the current task",
				"arguments":   []map[string]any{},
			},
		},
	})
}

func (s *Server) handlePromptsGet(req jsonrpcRequest) *jsonrpcResponse {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return rpcErr(req.ID, -32602, "Invalid params")
	}

	switch params.Name {
	case "apply-entry":
		slug := params.Arguments["slug"]
		if slug == "" {
			return rpcErr(req.ID, -32602, "slug argument is required")
		}
		entry, err := s.DB.GetEntry(slug)
		if err != nil {
			return rpcErr(req.ID, -32602, err.Error())
		}
		slog.Info("prompt call", "prompt", "apply-entry", "slug", slug)
		return rpcResult(req.ID, map[string]any{
			"description": "Apply entry: " + entry.Title,
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": fmt.Sprintf("You have been given the following guide to follow:\n\n# %s\n\n%s\n\nApply these guidelines to the current task. Follow them precisely.", entry.Title, entry.Content),
					},
				},
			},
		})

	case "review-with-entry":
		slug := params.Arguments["slug"]
		if slug == "" {
			return rpcErr(req.ID, -32602, "slug argument is required")
		}
		entry, err := s.DB.GetEntry(slug)
		if err != nil {
			return rpcErr(req.ID, -32602, err.Error())
		}
		slog.Info("prompt call", "prompt", "review-with-entry", "slug", slug)
		return rpcResult(req.ID, map[string]any{
			"description": "Review with entry: " + entry.Title,
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": fmt.Sprintf("Review the code in this conversation against the following guidelines:\n\n# %s\n\n%s\n\nPoint out any violations and suggest fixes. Be specific with line references.", entry.Title, entry.Content),
					},
				},
			},
		})

	case "save-learnings":
		slog.Info("prompt call", "prompt", "save-learnings")
		return rpcResult(req.ID, map[string]any{
			"description": "Save learnings from the current task",
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": `Analyze what was accomplished in this conversation and identify reusable knowledge that should be saved. For each piece of knowledge:

1. Determine if it's a skill, rule, context, pattern, reference, or guide
2. Choose a descriptive slug (e.g. "rust-error-handling", "project-foo-auth-flow")
3. Write concise, actionable content (under 32KB)
4. Assign appropriate language, domain, project, and tags

Use the create_entry tool to save each piece. Keep entries granular -- one concept per entry.`,
					},
				},
			},
		})

	default:
		return rpcErr(req.ID, -32602, "Unknown prompt: "+params.Name)
	}
}

// --- Tool definitions ---

func toolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"name":        "search_entries",
			"description": "Search knowledge entries using full-text search. Returns matching entries with snippets (no full content).",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query":    map[string]any{"type": "string", "description": "Search query"},
					"language": map[string]any{"type": "string", "description": "Filter by programming language"},
					"domain":   map[string]any{"type": "string", "description": "Filter by domain (e.g. fintech, ml, cli)"},
					"kind":     map[string]any{"type": "string", "description": "Filter by kind (skill, rule, context, pattern, reference, guide)"},
					"tag":      map[string]any{"type": "string", "description": "Filter by tag"},
					"project":  map[string]any{"type": "string", "description": "Filter by project"},
					"limit":    map[string]any{"type": "integer", "description": "Max results (default 10, max 50)"},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "get_entry",
			"description": "Get a single knowledge entry by its slug, including full content.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"slug": map[string]any{"type": "string", "description": "The unique slug of the entry"},
				},
				"required": []string{"slug"},
			},
		},
		{
			"name":        "get_entries_by_context",
			"description": "Get all entries matching the given context (language, domain, kind, tags, project). Returns full content. Use this at the start of a task to load relevant knowledge.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"language": map[string]any{"type": "string", "description": "Programming language"},
					"domain":   map[string]any{"type": "string", "description": "Domain"},
					"kind":     map[string]any{"type": "string", "description": "Entry kind"},
					"tags":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tags to match (all must be present)"},
					"project":  map[string]any{"type": "string", "description": "Project slug"},
					"limit":    map[string]any{"type": "integer", "description": "Max results (default 20, max 50)"},
				},
			},
		},
		{
			"name":        "list_entries",
			"description": "List all knowledge entries (slug, title, kind, language, domain -- no content). Supports optional filters.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"kind":     map[string]any{"type": "string", "description": "Filter by kind"},
					"language": map[string]any{"type": "string", "description": "Filter by language"},
					"domain":   map[string]any{"type": "string", "description": "Filter by domain"},
					"project":  map[string]any{"type": "string", "description": "Filter by project"},
				},
			},
		},
		{
			"name":        "list_tags",
			"description": "List all tags with their entry counts.",
			"inputSchema": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
			},
		},
		{
			"name":        "create_entry",
			"description": "Create a new knowledge entry. Requires slug, title, and content. Blocked if the database is locked.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"slug":        map[string]any{"type": "string", "description": "Unique slug (e.g. rust-error-handling)"},
					"title":       map[string]any{"type": "string", "description": "Entry title"},
					"content":     map[string]any{"type": "string", "description": "Main content (markdown, max 32KB)"},
					"description": map[string]any{"type": "string", "description": "Short summary for discovery"},
					"kind":        map[string]any{"type": "string", "description": "Entry kind: skill, rule, context, pattern, reference, guide"},
					"language":    map[string]any{"type": "string", "description": "Programming language"},
					"domain":      map[string]any{"type": "string", "description": "Domain"},
					"project":     map[string]any{"type": "string", "description": "Project slug"},
					"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tags"},
				},
				"required": []string{"slug", "title", "content"},
			},
		},
		{
			"name":        "update_entry",
			"description": "Update an existing knowledge entry by slug. Only provided fields are updated. Blocked if the database is locked.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"slug":        map[string]any{"type": "string", "description": "Slug of the entry to update"},
					"title":       map[string]any{"type": "string", "description": "New title"},
					"content":     map[string]any{"type": "string", "description": "New content"},
					"description": map[string]any{"type": "string", "description": "New description"},
					"kind":        map[string]any{"type": "string", "description": "New kind"},
					"language":    map[string]any{"type": "string", "description": "New language"},
					"domain":      map[string]any{"type": "string", "description": "New domain"},
					"project":     map[string]any{"type": "string", "description": "New project"},
					"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "New tags (replaces all existing tags)"},
				},
				"required": []string{"slug"},
			},
		},
		{
			"name":        "delete_entry",
			"description": "Delete a knowledge entry by slug. Blocked if the database is locked.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"slug": map[string]any{"type": "string", "description": "Slug of the entry to delete"},
				},
				"required": []string{"slug"},
			},
		},
	}
}

// --- helpers ---

func toolResult(id any, data any) *jsonrpcResponse {
	j, _ := json.Marshal(data)
	return rpcResult(id, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(j)},
		},
		"isError": false,
	})
}

func toolError(id any, msg string) *jsonrpcResponse {
	return rpcResult(id, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": msg},
		},
		"isError": true,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func str(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intVal(m map[string]any, key string, def int) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case json.Number:
			i, _ := n.Int64()
			return int(i)
		}
	}
	return def
}

func strSlice(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	switch s := v.(type) {
	case []any:
		result := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	case []string:
		return s
	}
	return nil
}
