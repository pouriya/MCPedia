# MCPedia

A knowledge base server for AI agents via the [Model Context Protocol](https://modelcontextprotocol.io/) (MCP). MCPedia stores, searches, and manages reusable knowledge entries -- skills, rules, patterns, guides, and references -- accessible to AI coding agents like Cursor, Codex, Claude, and others through the MCP standard.

## Core Concepts

### Entries

Entries are the primary knowledge units in MCPedia. Each entry has:

- A unique **slug** (URL-safe identifier, e.g. `rust-error-handling`)
- A **title** and optional **description**
- **Content** in Markdown format (up to 32 KB)
- **Kind** classification: `skill`, `rule`, `context`, `pattern`, `reference`, or `guide`
- Optional metadata: **language**, **domain**, **project**
- One or more **tags** for categorization
- Automatic **version** tracking and **timestamps**

Example:

```json
{
  "slug": "rust-error-handling",
  "title": "Rust Error Handling",
  "description": "Idiomatic error handling patterns in Rust",
  "kind": "skill",
  "language": "rust",
  "domain": "",
  "project": "",
  "tags": ["rust", "errors", "result"],
  "version": 1,
  "content": "# Rust Error Handling\n\nUse `Result<T, E>` for recoverable errors..."
}
```

### Tags

Tags provide flexible categorization across entries. Each tag tracks how many entries reference it, enabling discovery of related knowledge. Tags are managed automatically -- they are created when first used and cleaned up when no longer referenced.

### Usage Statistics

MCPedia tracks usage statistics for each entry:

- **Reads**: incremented when an entry is fetched by slug or by context filters
- **Searches**: incremented when an entry appears in full-text search results
- **Updates**: incremented when an entry is modified

This enables agents to understand which knowledge is most frequently accessed.

### Write Lock

MCPedia supports a database-level write lock to prevent AI agents from modifying the knowledge base when controlled access is desired. When locked, all write operations (`create_entry`, `update_entry`, `delete_entry`) are rejected. The lock is protected by a SHA-256 hashed token -- only the holder of the original token can unlock it.

## API

MCPedia implements the MCP protocol version `2025-11-25` over HTTP using JSON-RPC 2.0. The server exposes a single endpoint at `POST /mcp`.

### Tools

- **`search_entries`**
  - Full-text search across entries using SQLite FTS5 with snippet highlighting
  - Inputs:
    - `query` (string, required): Search query text
    - `language` (string, optional): Filter results by programming language (e.g. `"rust"`, `"python"`)
    - `domain` (string, optional): Filter results by domain (e.g. `"backend"`, `"security"`)
    - `kind` (string, optional): Filter results by kind (`"skill"`, `"rule"`, `"context"`, `"pattern"`, `"reference"`, `"guide"`)
    - `tag` (string, optional): Filter results by a specific tag
    - `project` (string, optional): Filter results by project slug
    - `limit` (integer, optional): Maximum number of results to return (default: 10, max: 50)
  - Returns matching entries with search snippets (content is not included in full)

- **`get_entry`**
  - Retrieve a single entry by its unique slug, including full content
  - Inputs:
    - `slug` (string, required): The unique slug identifier of the entry
  - Returns the complete entry with all metadata, tags, and full Markdown content
  - Increments the entry's read count in usage statistics

- **`get_entries_by_context`**
  - Retrieve entries matching contextual filters, with full content included
  - Inputs:
    - `language` (string, optional): Filter by programming language
    - `domain` (string, optional): Filter by domain
    - `kind` (string, optional): Filter by entry kind
    - `tags` (array of strings, optional): Filter by tags -- all specified tags must be present on the entry
    - `project` (string, optional): Filter by project slug
    - `limit` (integer, optional): Maximum number of results (default: 20, max: 50)
  - Returns full entries with content, suitable for injecting knowledge into agent context
  - Increments read counts for all returned entries

- **`list_entries`**
  - List all entries without content, with optional metadata filters
  - Inputs:
    - `kind` (string, optional): Filter by entry kind
    - `language` (string, optional): Filter by programming language
    - `domain` (string, optional): Filter by domain
    - `project` (string, optional): Filter by project slug
  - Returns entry metadata (slug, title, description, kind, language, domain, project) without content

- **`list_tags`**
  - List all tags in the knowledge base with their usage counts
  - No inputs required
  - Returns an array of tags, each with its `name` and `count` of associated entries

- **`create_entry`**
  - Create a new knowledge entry in the database
  - Inputs:
    - `slug` (string, required): Unique slug identifier (URL-safe)
    - `title` (string, required): Human-readable title
    - `content` (string, required): Markdown content (max 32 KB)
    - `description` (string, optional): Short summary of the entry
    - `kind` (string, optional): Entry kind -- one of `"skill"`, `"rule"`, `"context"`, `"pattern"`, `"reference"`, `"guide"` (default: `"skill"`)
    - `language` (string, optional): Programming language the entry relates to
    - `domain` (string, optional): Domain or area (e.g. `"backend"`, `"testing"`)
    - `project` (string, optional): Project slug this entry belongs to
    - `tags` (array of strings, optional): Tags for categorization
  - Returns the created entry with all fields populated
  - Blocked when the database write lock is active

- **`update_entry`**
  - Update an existing entry -- only provided fields are modified
  - Inputs:
    - `slug` (string, required): Slug of the entry to update
    - `title` (string, optional): New title
    - `content` (string, optional): New Markdown content (max 32 KB)
    - `description` (string, optional): New description
    - `kind` (string, optional): New kind classification
    - `language` (string, optional): New programming language
    - `domain` (string, optional): New domain
    - `project` (string, optional): New project slug
    - `tags` (array of strings, optional): New tags -- replaces all existing tags
  - Returns the updated entry; automatically increments version and updates timestamp
  - Blocked when the database write lock is active

- **`delete_entry`**
  - Permanently delete an entry by slug
  - Inputs:
    - `slug` (string, required): Slug of the entry to delete
  - Returns a confirmation message
  - Cascading deletion removes associated tags and usage statistics
  - Blocked when the database write lock is active

### Resources

MCPedia exposes entries as MCP resources, allowing clients to browse and read knowledge entries using standard resource URIs.

- **`resources/list`**
  - Returns a paginated list of all entries as resources
  - Each resource includes:
    - `uri`: `mcpedia://entries/<slug>`
    - `name`: The entry slug
    - `description`: The entry description
    - `mimeType`: `text/markdown`
  - Pagination: 50 entries per page, cursor-based (base64-encoded offset)

- **`resources/read`**
  - Read a single entry's content by its resource URI
  - Input: `uri` (string) in the format `mcpedia://entries/<slug>`
  - Returns the entry's full Markdown content

- **`resources/templates/list`**
  - Returns a URI template for accessing entries:
    - `uriTemplate`: `mcpedia://entries/{slug}`
    - `name`: `MCPedia Entry`
    - `description`: `Access an entry by slug`
    - `mimeType`: `text/markdown`

### Prompts

MCPedia provides three built-in prompts that help AI agents apply, review, and capture knowledge.

- **`apply-entry`**
  - Apply a knowledge entry's guidelines to the current task
  - Arguments:
    - `slug` (string, required): Slug of the entry to apply
  - Fetches the entry from the database and returns a prompt message with the entry's full content embedded, instructing the agent to follow the entry's guidelines

- **`review-with-entry`**
  - Review code against a knowledge entry's guidelines
  - Arguments:
    - `slug` (string, required): Slug of the entry to review against
  - Fetches the entry and returns a prompt message instructing the agent to evaluate code according to the entry's rules and best practices

- **`save-learnings`**
  - Extract and save reusable knowledge from the current task
  - No arguments required
  - Returns a prompt instructing the agent to identify reusable patterns, techniques, or rules from the current session and save them as new entries using the `create_entry` tool

## Configuration

MCPedia is configured through environment variables and/or CLI flags. Flags take precedence over environment variables.

| Environment Variable | CLI Flag  | Default       | Description                                           |
|----------------------|-----------|---------------|-------------------------------------------------------|
| `MCPEDIA_DB`         | `--db`    | `mcpedia.db`  | Path to the SQLite database file                      |
| `MCPEDIA_ADDR`       | `--addr`  | `:8080`       | HTTP server listen address                            |
| `MCPEDIA_TOKEN`      | `--token` | *(empty)*     | Bearer token for authentication (empty = no auth)     |

When a token is set, all HTTP requests must include an `Authorization: Bearer <token>` header. This protects the MCP endpoint from unauthorized access.

## CLI Commands

MCPedia ships as a single binary with subcommands for database management, server operation, and entry management.

```
mcpedia <command> [flags]

Commands:
  init      Create and initialize the database
  serve     Start the MCP HTTP server
  add       Add a new knowledge entry
  edit      Edit an existing entry
  list      List entries with optional filters
  lock      Lock the database (prevent AI writes)
  unlock    Unlock the database
  export    Export all entries as Markdown files
```

### `mcpedia init`

Creates and initializes the SQLite database with the required schema.

```bash
mcpedia init --db ./mcpedia.db
```

### `mcpedia serve`

Starts the MCP HTTP server, ready to accept JSON-RPC 2.0 requests from MCP clients.

```bash
mcpedia serve --db ./mcpedia.db --addr :8080 --token my-secret-token
```

### `mcpedia add`

Adds a new knowledge entry. Content is read from a file.

```bash
mcpedia add \
  --slug rust-error-handling \
  --title "Rust Error Handling" \
  --description "Idiomatic error handling patterns in Rust" \
  --file content.md \
  --kind skill \
  --language rust \
  --tags rust,errors,result
```

### `mcpedia edit`

Edits an existing entry. Only provided fields are updated.

```bash
mcpedia edit \
  --slug rust-error-handling \
  --title "Rust Error Handling (Updated)" \
  --file updated-content.md \
  --tags rust,errors,result,anyhow
```

### `mcpedia list`

Lists entries with optional filters.

```bash
mcpedia list --language rust --kind skill
```

### `mcpedia lock` / `mcpedia unlock`

Lock the database to prevent AI agents from creating, updating, or deleting entries. Useful when you want read-only access for agents.

```bash
mcpedia lock --db ./mcpedia.db --token my-lock-secret
mcpedia unlock --db ./mcpedia.db --token my-lock-secret
```

### `mcpedia export`

Export all entries as Markdown files with YAML frontmatter.

```bash
mcpedia export --db ./mcpedia.db --out ./backup
```

Each entry is written to `<slug>.md`:

```markdown
---
title: "Rust Error Handling"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, errors, result]
description: "Idiomatic error handling patterns in Rust"
---

# Rust Error Handling

Use `Result<T, E>` for recoverable errors...
```

## Usage with Cursor

Add MCPedia as an MCP server in your Cursor configuration (`.cursor/mcp.json`):

### Binary

```json
{
  "mcpServers": {
    "mcpedia": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCPEDIA_TOKEN": "your-secret-token"
      }
    }
  }
}
```

Then start the server:

```bash
mcpedia serve --addr :8080 --token your-secret-token
```

### Docker

```json
{
  "mcpServers": {
    "mcpedia": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

Then run via Docker:

```bash
docker run -p 8080:8080 \
  -v $(pwd)/mcpedia.db:/mcpedia.db \
  -e MCPEDIA_DB=/mcpedia.db \
  -e MCPEDIA_TOKEN=your-secret-token \
  ghcr.io/pouriya/mcpedia:latest serve
```

## Usage with Claude Desktop

Add MCPedia to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mcpedia": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

## Usage with VS Code

Add the configuration to `.vscode/mcp.json` in your workspace or to your user-level MCP configuration (Command Palette > `MCP: Open User Configuration`):

```json
{
  "servers": {
    "mcpedia": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [GitHub Releases](https://github.com/pouriya/mcpedia/releases) page:

| Platform          | Binary                          |
|-------------------|---------------------------------|
| macOS (arm64)     | `mcpedia-*-darwin-arm64`        |
| Linux (amd64)     | `mcpedia-*-linux-amd64`        |
| Windows (amd64)   | `mcpedia-*-windows-amd64.exe`  |

### Docker

Pull from GitHub Container Registry:

```bash
docker pull ghcr.io/pouriya/mcpedia:latest
```

Run with a persistent database:

```bash
docker run -p 8080:8080 \
  -v $(pwd)/mcpedia.db:/mcpedia.db \
  -e MCPEDIA_DB=/mcpedia.db \
  -e MCPEDIA_TOKEN=your-secret-token \
  ghcr.io/pouriya/mcpedia:latest serve
```

### Build from Source

Requires Go 1.24+.

```bash
git clone https://github.com/pouriya/mcpedia.git
cd mcpedia

# Development build
make dev

# Or release build (stripped, cross-platform)
make release GOOS=linux GOARCH=amd64
```

Initialize and start:

```bash
./mcpedia init
./mcpedia serve --addr :8080 --token my-secret-token
```

### Build Docker Image Locally

```bash
make docker
```

This creates `mcpedia:latest` and `mcpedia:<version>` images.

## Architecture

```
┌─────────────────────────────────────────────┐
│              MCP Clients                    │
│  (Cursor, Codex, Claude, VS Code, etc.)    │
└──────────────────┬──────────────────────────┘
                   │ HTTP POST /mcp
                   │ JSON-RPC 2.0
                   ▼
┌─────────────────────────────────────────────┐
│           mcpedia binary                    │
│                                             │
│  ┌──────────────┐    ┌──────────────────┐   │
│  │ CLI Commands  │    │  MCP HTTP Server │   │
│  │ (init, add,   │    │  (tools,         │   │
│  │  edit, list,  │    │   resources,     │   │
│  │  lock, export)│    │   prompts)       │   │
│  └──────┬───────┘    └────────┬─────────┘   │
│         │                     │             │
│         ▼                     ▼             │
│  ┌──────────────────────────────────────┐   │
│  │         SQLite DB Layer              │   │
│  │  (FTS5 search, tags, stats, lock)    │   │
│  └──────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

### Project Structure

```
mcpedia/
├── cmd/mcpedia/
│   └── main.go              # CLI entry point, all subcommands
├── internal/
│   ├── db/
│   │   ├── db.go            # Database operations (CRUD, search, stats, lock)
│   │   └── schema.sql       # SQLite schema (embedded via go:embed)
│   └── mcp/
│       └── mcp.go           # MCP HTTP server (JSON-RPC 2.0, tools, resources, prompts)
├── test/
│   └── integration_test.go  # Comprehensive integration tests
├── Makefile                 # Build automation
├── Dockerfile               # Multi-stage Docker build
├── go.mod                   # Go module (pure Go SQLite, no CGO dependency at runtime)
└── go.sum
```

### Key Design Decisions

- **Pure Go SQLite** via `modernc.org/sqlite` -- no CGO runtime dependency, single static binary
- **Single endpoint** (`POST /mcp`) -- standard MCP streamable HTTP transport
- **MCP protocol `2025-11-25`** -- full compliance with tools, resources, and prompts
- **FTS5 full-text search** -- fast, ranked search with snippet highlighting
- **Minimal codebase** -- 3 Go source files + 1 SQL schema, no unnecessary abstractions
- **Session management** -- UUID-based sessions with `Mcp-Session-Id` header
- **Vendored dependencies** -- reproducible builds without network access

## Database Schema

MCPedia uses SQLite with the following tables:

| Table          | Purpose                                          |
|----------------|--------------------------------------------------|
| `entries`      | Knowledge entries with slug, title, content, metadata |
| `tags`         | Unique tag names                                 |
| `entry_tags`   | Many-to-many relationship between entries and tags |
| `entry_stats`  | Usage statistics (reads, searches, updates)      |
| `lock`         | Write lock state (single row)                    |
| `entries_fts`  | FTS5 virtual table for full-text search          |

Constraints and features:
- `CHECK(length(content) <= 32768)` -- 32 KB content size limit
- Unique slug constraint on entries
- Foreign keys with `CASCADE` deletes
- FTS5 sync via `AFTER INSERT/UPDATE/DELETE` triggers
- Indexes on `language`, `domain`, `kind`, `project` columns

## Security

- **Bearer token authentication** -- optional but recommended; protects the MCP endpoint
- **Write lock mechanism** -- SHA-256 hashed token prevents unauthorized modifications
- **Parameterized SQL queries** -- protection against SQL injection
- **Content size limits** -- 32 KB maximum prevents abuse
- **Session validation** -- requests after initialization must include a valid `Mcp-Session-Id`

## Development

```bash
# Run tests with race detection
make test

# Run tests with coverage report
make test-cover

# Format check
make fmt

# Go vet
make vet

# Clean build artifacts
make clean
```

## License

This MCP server is licensed under the BSD 3-Clause License. Copyright (c) 2022, Pouriya Jim Jahanbakhsh. For more details, please see the [LICENSE](LICENSE) file in the project repository.
