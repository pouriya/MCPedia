# Learnings

Seed knowledge entries for MCPedia, organized by language.

## Layout

- `rust/` — Markdown files for Rust (e.g. `rust-error-handling.md`)
- `golang/` — Markdown files for Go
- `python/` — Markdown files for Python

Each file must be in **export format**: YAML frontmatter (`title`, `kind`, `language`, `domain`, `project`, `tags`, optional `description`) between `---` delimiters, then the body. The filename (without `.md`) is the entry slug.

## Build

From this directory, run:

```bash
make build
```

This imports every `*.md` file in each language directory into mcpedia. Override the binary or database:

```bash
make build mcpedia_exe=../mcpedia mcpedia_db=../mcpedia.db
```

From the project root:

```bash
make -C learnings build
```
