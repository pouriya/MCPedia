package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/pouriya/mcpedia/internal/db"
	"github.com/pouriya/mcpedia/internal/mcp"
)

const defaultDB = "mcpedia.db"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "init":
		cmdInit(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	case "add":
		cmdAdd(os.Args[2:])
	case "edit":
		cmdEdit(os.Args[2:])
	case "list":
		cmdList(os.Args[2:])
	case "lock":
		cmdLock(os.Args[2:])
	case "unlock":
		cmdUnlock(os.Args[2:])
	case "export":
		cmdExport(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `mcpedia - Knowledge base for AI agents via MCP

Usage:
  mcpedia <command> [flags]

Commands:
  init     Create and initialize the database
  serve    Start the MCP HTTP server
  add      Add a new entry
  edit     Edit an existing entry
  list     List entries
  lock     Lock the database (prevent AI writes)
  unlock   Unlock the database
  export   Export entries as markdown files

Environment variables:
  MCPEDIA_DB      Database path (default: %s)
  MCPEDIA_ADDR    Server address (default: :8080)
  MCPEDIA_TOKEN   Bearer token for auth
  MCPEDIA_DEBUG   Enable debug logging (any non-empty value)

Run 'mcpedia <command> --help' for more information.
`, defaultDB)
}

// --- init ---

func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path")
	fs.Parse(args)

	path := resolve(*dbPath, "MCPEDIA_DB", defaultDB)

	d, err := db.Open(path)
	if err != nil {
		fatal("init: %v", err)
	}
	d.Close()
	fmt.Printf("Database initialized at %s\n", path)
}

// --- serve ---

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path")
	addr := fs.String("addr", "", "Listen address")
	token := fs.String("token", "", "Bearer token for auth (empty = no auth)")
	debug := fs.Bool("debug", false, "Enable debug logging")
	fs.Parse(args)

	path := resolve(*dbPath, "MCPEDIA_DB", defaultDB)
	listenAddr := resolve(*addr, "MCPEDIA_ADDR", ":8080")
	authToken := resolve(*token, "MCPEDIA_TOKEN", "")

	if !*debug && os.Getenv("MCPEDIA_DEBUG") != "" {
		*debug = true
	}

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	d, err := db.Open(path)
	if err != nil {
		fatal("serve: %v", err)
	}
	defer d.Close()

	server := &mcp.Server{DB: d, Token: authToken}

	mux := http.NewServeMux()
	mux.Handle("/mcp", server)
	// Also handle root for convenience
	mux.Handle("/", server)

	slog.Info("server starting",
		"addr", listenAddr,
		"db", path,
		"auth", authToken != "",
		"debug", *debug,
	)

	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		fatal("serve: %v", err)
	}
}

// --- add ---

func cmdAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path")
	slug := fs.String("slug", "", "Unique slug (required)")
	title := fs.String("title", "", "Title (required)")
	kind := fs.String("kind", "skill", "Entry kind")
	language := fs.String("language", "", "Programming language")
	domain := fs.String("domain", "", "Domain")
	project := fs.String("project", "", "Project")
	tags := fs.String("tags", "", "Comma-separated tags")
	description := fs.String("description", "", "Short description")
	file := fs.String("file", "", "Path to content file (required)")
	fs.Parse(args)

	path := resolve(*dbPath, "MCPEDIA_DB", defaultDB)

	if *slug == "" || *title == "" || *file == "" {
		fmt.Fprintln(os.Stderr, "Error: --slug, --title, and --file are required")
		fs.Usage()
		os.Exit(1)
	}

	content, err := os.ReadFile(*file)
	if err != nil {
		fatal("read file: %v", err)
	}

	d, err := db.Open(path)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer d.Close()

	e := &db.Entry{
		Slug:        *slug,
		Title:       *title,
		Description: *description,
		Content:     string(content),
		Kind:        *kind,
		Language:    *language,
		Domain:      *domain,
		Project:     *project,
		Tags:        parseTags(*tags),
	}
	if err := d.CreateEntry(e); err != nil {
		fatal("create: %v", err)
	}

	fmt.Printf("Entry created: %s (%s)\n", e.Slug, e.Title)
	fmt.Printf("  Kind: %s  Language: %s  Domain: %s  Project: %s\n", e.Kind, e.Language, e.Domain, e.Project)
	if len(e.Tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(e.Tags, ", "))
	}
	fmt.Printf("  Version: %d  Content: %d bytes\n", e.Version, len(e.Content))
}

// --- edit ---

func cmdEdit(args []string) {
	fs := flag.NewFlagSet("edit", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path")
	slug := fs.String("slug", "", "Slug of entry to edit (required)")
	title := fs.String("title", "", "New title")
	kind := fs.String("kind", "", "New kind")
	language := fs.String("language", "", "New language")
	domain := fs.String("domain", "", "New domain")
	project := fs.String("project", "", "New project")
	tags := fs.String("tags", "", "New comma-separated tags (replaces all)")
	description := fs.String("description", "", "New description")
	file := fs.String("file", "", "Path to new content file")
	fs.Parse(args)

	path := resolve(*dbPath, "MCPEDIA_DB", defaultDB)

	if *slug == "" {
		fmt.Fprintln(os.Stderr, "Error: --slug is required")
		fs.Usage()
		os.Exit(1)
	}

	d, err := db.Open(path)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer d.Close()

	fields := map[string]any{}

	// Only include flags that were explicitly set
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "title":
			fields["title"] = *title
		case "kind":
			fields["kind"] = *kind
		case "language":
			fields["language"] = *language
		case "domain":
			fields["domain"] = *domain
		case "project":
			fields["project"] = *project
		case "description":
			fields["description"] = *description
		case "tags":
			fields["tags"] = parseTags(*tags)
		case "file":
			content, err := os.ReadFile(*file)
			if err != nil {
				fatal("read file: %v", err)
			}
			fields["content"] = string(content)
		}
	})

	if len(fields) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no fields to update. Provide at least one field flag.")
		os.Exit(1)
	}

	if err := d.UpdateEntry(*slug, fields); err != nil {
		fatal("update: %v", err)
	}

	entry, err := d.GetEntry(*slug)
	if err != nil {
		fatal("get: %v", err)
	}

	fmt.Printf("Entry updated: %s (%s)\n", entry.Slug, entry.Title)
	fmt.Printf("  Kind: %s  Language: %s  Domain: %s  Project: %s\n", entry.Kind, entry.Language, entry.Domain, entry.Project)
	if len(entry.Tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(entry.Tags, ", "))
	}
	fmt.Printf("  Version: %d  Content: %d bytes\n", entry.Version, len(entry.Content))
}

// --- list ---

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path")
	kind := fs.String("kind", "", "Filter by kind")
	language := fs.String("language", "", "Filter by language")
	domain := fs.String("domain", "", "Filter by domain")
	project := fs.String("project", "", "Filter by project")
	tag := fs.String("tag", "", "Filter by tag")
	fs.Parse(args)

	path := resolve(*dbPath, "MCPEDIA_DB", defaultDB)

	d, err := db.Open(path)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer d.Close()

	entries, err := d.ListEntries(db.Filter{
		Kind:     *kind,
		Language: *language,
		Domain:   *domain,
		Project:  *project,
		Tag:      *tag,
	})
	if err != nil {
		fatal("list: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No entries found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SLUG\tTITLE\tKIND\tLANGUAGE\tDOMAIN\tVERSION")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n", e.Slug, e.Title, e.Kind, e.Language, e.Domain, e.Version)
	}
	w.Flush()
	fmt.Printf("\n%d entries\n", len(entries))
}

// --- lock ---

func cmdLock(args []string) {
	fs := flag.NewFlagSet("lock", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path")
	token := fs.String("token", "", "Lock token (required)")
	fs.Parse(args)

	path := resolve(*dbPath, "MCPEDIA_DB", defaultDB)

	if *token == "" {
		fmt.Fprintln(os.Stderr, "Error: --token is required")
		os.Exit(1)
	}

	d, err := db.Open(path)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer d.Close()

	if err := d.Lock(*token); err != nil {
		fatal("lock: %v", err)
	}
	fmt.Println("Database locked. AI write operations are now disabled.")
}

// --- unlock ---

func cmdUnlock(args []string) {
	fs := flag.NewFlagSet("unlock", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path")
	token := fs.String("token", "", "Lock token (required, must match the token used to lock)")
	fs.Parse(args)

	path := resolve(*dbPath, "MCPEDIA_DB", defaultDB)

	if *token == "" {
		fmt.Fprintln(os.Stderr, "Error: --token is required")
		os.Exit(1)
	}

	d, err := db.Open(path)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer d.Close()

	if err := d.Unlock(*token); err != nil {
		fatal("unlock: %v", err)
	}
	fmt.Println("Database unlocked. AI write operations are now enabled.")
}

// --- export ---

func cmdExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path")
	out := fs.String("out", "export", "Output directory")
	fs.Parse(args)

	path := resolve(*dbPath, "MCPEDIA_DB", defaultDB)

	d, err := db.Open(path)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer d.Close()

	entries, err := d.AllEntries()
	if err != nil {
		fatal("export: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No entries to export.")
		return
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fatal("mkdir: %v", err)
	}

	for _, e := range entries {
		filename := filepath.Join(*out, e.Slug+".md")

		var sb strings.Builder
		sb.WriteString("---\n")
		sb.WriteString(fmt.Sprintf("title: %q\n", e.Title))
		sb.WriteString(fmt.Sprintf("kind: %s\n", e.Kind))
		sb.WriteString(fmt.Sprintf("language: %s\n", e.Language))
		sb.WriteString(fmt.Sprintf("domain: %s\n", e.Domain))
		sb.WriteString(fmt.Sprintf("project: %s\n", e.Project))
		if len(e.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(e.Tags, ", ")))
		} else {
			sb.WriteString("tags: []\n")
		}
		if e.Description != "" {
			sb.WriteString(fmt.Sprintf("description: %q\n", e.Description))
		}
		sb.WriteString("---\n\n")
		sb.WriteString(e.Content)
		sb.WriteString("\n")

		if err := os.WriteFile(filename, []byte(sb.String()), 0o644); err != nil {
			fatal("write %s: %v", filename, err)
		}
		fmt.Printf("Exported: %s\n", filename)
	}
	fmt.Printf("\n%d entries exported to %s/\n", len(entries), *out)
}

// --- helpers ---

// resolve returns the flag value if non-empty, otherwise the env var, otherwise the default.
func resolve(flagVal, envKey, def string) string {
	if flagVal != "" {
		return flagVal
	}
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return def
}

func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
