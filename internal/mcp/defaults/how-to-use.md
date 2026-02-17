# How to Use MCPedia

This guide is for AI agents. Read it to learn how to use MCPedia correctly.

## What Is MCPedia

MCPedia is a knowledge base. You can read, search, and store reusable knowledge (skills, rules, patterns, guides). Use it to remember patterns across sessions and apply them to tasks.

## Tools — When to Use Each

| Tool | Use when |
|------|----------|
| `search_entries` | You have a keyword or phrase. Returns snippets, no full content. Good for discovery. |
| `get_entry` | You know the exact slug. Returns full content. Use after search or when slug is known. |
| `get_entries_by_context` | You want entries by language, domain, kind, tags, or project. Returns full content. Use for contextual injection. |
| `list_entries` | You need slugs and metadata only (no content). Use to browse or verify existence. |
| `list_tags` | You need all tags and their counts. Use to discover tags before filtering. |
| `create_entry` | Save new knowledge. Blocked when database is locked. |
| `update_entry` | Modify an existing entry by slug. Blocked when locked. |
| `delete_entry` | Remove an entry by slug. Blocked when locked. |

## Workflow

1. **Find knowledge** — Use `search_entries` with query and optional filters (`language`, `domain`, `kind`, `tag`, `project`). Or use `list_tags` then `get_entries_by_context` with `tags`.
2. **Get full content** — Use `get_entry` with the slug from search results.
3. **Apply it** — Use the `apply-entry` prompt with the slug to inject guidelines into your task.
4. **Save new knowledge** — Use the `save-learnings` prompt to extract and create entries, or call `create_entry` directly.

## Resources

Entries are exposed as MCP resources. URI format: `mcpedia://entries/{slug}`. Use `resources/read` with that URI to fetch entry content. This guide is available at `mcpedia://how-to-use`—read it when you need usage instructions. Use `resources/templates/list` to discover available URIs.

## Prompts

- **apply-entry** — Fetches an entry and instructs you to follow its guidelines. Argument: `slug`.
- **review-with-entry** — Fetches an entry and instructs you to review code against it. Argument: `slug`.
- **save-learnings** — Instructs you to extract reusable knowledge and create entries. No arguments.

## Entry Metadata

- **Kind**: `skill`, `rule`, `context`, `pattern`, `reference`, `guide`
- **language**: e.g. `rust`, `python`
- **domain**: e.g. `backend`, `testing`
- **project**: project slug
- **tags**: array of strings for flexible filtering

Use filters to narrow search and context queries.

## Write Lock

When the database is locked, `create_entry`, `update_entry`, and `delete_entry` fail. You can still read and search. Do not retry writes when locked.

## Rules

1. Prefer `search_entries` when you do not know the slug.
2. Prefer `get_entry` when you already have the slug.
3. Use `get_entries_by_context` when loading knowledge for a specific language, domain, or project.
4. Call `list_tags` to discover available tags before filtering by tag.
5. Use the prompts when the user asks to apply, review, or save knowledge.
6. Do not create duplicate entries; check with `search_entries` or `list_entries` first.
