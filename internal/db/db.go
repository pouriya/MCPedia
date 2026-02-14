package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Sentinel errors for known failure conditions. Use errors.Is(err, db.ErrNotFound) to check.
var (
	ErrNotFound = errors.New("entry not found")
	ErrLocked   = errors.New("database is locked")
)

//go:embed schema.sql
var schemaSQL string

// DB wraps the SQLite connection and provides all data operations.
type DB struct {
	db *sql.DB
}

// Entry represents a knowledge entry in the database.
type Entry struct {
	ID          int64    `json:"id"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content,omitempty"`
	Kind        string   `json:"kind"`
	Language    string   `json:"language"`
	Domain      string   `json:"domain"`
	Project     string   `json:"project"`
	Version     int      `json:"version"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	Tags        []string `json:"tags"`
	// Snippet is populated by search results only.
	Snippet string `json:"snippet,omitempty"`
}

// EntryStats holds usage statistics for an entry.
type EntryStats struct {
	Reads        int     `json:"reads"`
	Searches     int     `json:"searches"`
	Updates      int     `json:"updates"`
	LastReadAt   *string `json:"last_read_at"`
	LastSearchAt *string `json:"last_search_at"`
	LastUpdateAt *string `json:"last_update_at"`
}

// Tag represents a tag with its usage count.
type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Filter is used for list/search/context queries.
type Filter struct {
	Kind     string
	Language string
	Domain   string
	Project  string
	Tag      string
	Tags     []string // for get_entries_by_context
}

// Open opens (or creates) a SQLite database at path, runs PRAGMAs and schema.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// Set PRAGMAs
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err := sqlDB.Exec(pragma); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("pragma %q: %w", pragma, err)
		}
	}
	// Run main schema
	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}
	// Create FTS5 table if it doesn't exist.
	// We use a standalone FTS5 table (not external content) and manage sync manually
	// in CreateEntry/UpdateEntry/DeleteEntry for maximum reliability.
	var ftsExists int
	err = sqlDB.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='entries_fts'`).Scan(&ftsExists)
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("check fts: %w", err)
	}
	if ftsExists == 0 {
		ftsSQL := `CREATE VIRTUAL TABLE entries_fts USING fts5(title, description, content)`
		if _, err := sqlDB.Exec(ftsSQL); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("create fts: %w", err)
		}
	}
	// Connection pool limits (database/sql best practices)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	return &DB{db: sqlDB}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// CreateEntry inserts a new entry with its tags and stats row.
func (d *DB) CreateEntry(ctx context.Context, e *Entry) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO entries (slug, title, description, content, kind, language, domain, project)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Slug, e.Title, e.Description, e.Content,
		defaultStr(e.Kind, "skill"), e.Language, e.Domain, e.Project,
	)
	if err != nil {
		return fmt.Errorf("insert entry: %w", err)
	}
	entryID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	e.ID = entryID

	// Insert into FTS
	if _, err := tx.ExecContext(ctx, `INSERT INTO entries_fts(rowid, title, description, content) VALUES (?, ?, ?, ?)`,
		entryID, e.Title, e.Description, e.Content); err != nil {
		return fmt.Errorf("insert fts: %w", err)
	}

	// Insert stats row
	if _, err := tx.ExecContext(ctx, `INSERT INTO entry_stats (entry_id) VALUES (?)`, entryID); err != nil {
		return fmt.Errorf("insert stats: %w", err)
	}

	// Insert tags
	if err := setTags(ctx, tx, entryID, e.Tags); err != nil {
		return fmt.Errorf("set tags: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	// Read back the created_at/updated_at/version that the DB set
	row := d.db.QueryRowContext(ctx, `SELECT version, created_at, updated_at FROM entries WHERE id = ?`, entryID)
	return row.Scan(&e.Version, &e.CreatedAt, &e.UpdatedAt)
}

// GetEntry retrieves a full entry by slug and bumps the read counter.
func (d *DB) GetEntry(ctx context.Context, slug string) (*Entry, error) {
	e := &Entry{}
	row := d.db.QueryRowContext(ctx,
		`SELECT id, slug, title, description, content, kind, language, domain, project, version, created_at, updated_at
		 FROM entries WHERE slug = ?`, slug,
	)
	if err := row.Scan(&e.ID, &e.Slug, &e.Title, &e.Description, &e.Content,
		&e.Kind, &e.Language, &e.Domain, &e.Project, &e.Version,
		&e.CreatedAt, &e.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entry not found: %s: %w", slug, ErrNotFound)
		}
		return nil, fmt.Errorf("get entry: %w", err)
	}
	tags, err := getTagsForEntry(ctx, d.db, e.ID)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}
	e.Tags = tags

	// Bump read stats (best-effort; do not fail the request)
	now := time.Now().UTC().Format(time.DateTime)
	if _, err := d.db.ExecContext(ctx, `UPDATE entry_stats SET reads = reads + 1, last_read_at = ? WHERE entry_id = ?`, now, e.ID); err != nil {
		slog.Debug("update read stats", "err", err, "entry_id", e.ID)
	}
	return e, nil
}

// UpdateEntry updates only the provided fields for the entry identified by slug.
// Supported keys: title, description, content, kind, language, domain, project, tags.
func (d *DB) UpdateEntry(ctx context.Context, slug string, fields map[string]any) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	// Get entry ID
	var entryID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM entries WHERE slug = ?`, slug).Scan(&entryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("entry not found: %s: %w", slug, ErrNotFound)
		}
		return fmt.Errorf("lookup: %w", err)
	}

	// Build dynamic UPDATE
	setClauses := []string{}
	args := []any{}
	for _, col := range []string{"title", "description", "content", "kind", "language", "domain", "project"} {
		if v, ok := fields[col]; ok {
			setClauses = append(setClauses, col+" = ?")
			args = append(args, v)
		}
	}
	// Always bump version and updated_at
	setClauses = append(setClauses, "version = version + 1", "updated_at = datetime('now')")
	args = append(args, entryID)
	query := "UPDATE entries SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("update entry: %w", err)
	}

	// Sync FTS: delete old row, re-insert with current values from entries
	if _, err := tx.ExecContext(ctx, `DELETE FROM entries_fts WHERE rowid = ?`, entryID); err != nil {
		return fmt.Errorf("fts delete: %w", err)
	}
	var ftsTitle, ftsDesc, ftsContent string
	if err := tx.QueryRowContext(ctx, `SELECT title, description, content FROM entries WHERE id = ?`, entryID).Scan(&ftsTitle, &ftsDesc, &ftsContent); err != nil {
		return fmt.Errorf("fts read: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO entries_fts(rowid, title, description, content) VALUES (?, ?, ?, ?)`,
		entryID, ftsTitle, ftsDesc, ftsContent); err != nil {
		return fmt.Errorf("fts insert: %w", err)
	}

	// Handle tags if provided
	if tagsVal, ok := fields["tags"]; ok {
		var tagList []string
		switch v := tagsVal.(type) {
		case []string:
			tagList = v
		case []any:
			for _, t := range v {
				if s, ok := t.(string); ok {
					tagList = append(tagList, s)
				}
			}
		}
		if err := setTags(ctx, tx, entryID, tagList); err != nil {
			return fmt.Errorf("update tags: %w", err)
		}
	}

	// Bump update stats
	now := time.Now().UTC().Format(time.DateTime)
	if _, err := tx.ExecContext(ctx, `UPDATE entry_stats SET updates = updates + 1, last_update_at = ? WHERE entry_id = ?`, now, entryID); err != nil {
		return fmt.Errorf("update stats: %w", err)
	}

	return tx.Commit()
}

// DeleteEntry removes an entry by slug. CASCADE handles entry_tags and entry_stats.
// FTS and entry deletion are done in a single transaction for consistency.
func (d *DB) DeleteEntry(ctx context.Context, slug string) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	var entryID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM entries WHERE slug = ?`, slug).Scan(&entryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("entry not found: %s: %w", slug, ErrNotFound)
		}
		return fmt.Errorf("lookup: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM entries_fts WHERE rowid = ?`, entryID); err != nil {
		return fmt.Errorf("delete fts: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM entries WHERE id = ?`, entryID); err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}
	return tx.Commit()
}

// ListEntries returns entries without content, optionally filtered.
func (d *DB) ListEntries(ctx context.Context, f Filter) ([]Entry, error) {
	query := `SELECT e.id, e.slug, e.title, e.description, e.kind, e.language, e.domain, e.project, e.version, e.created_at, e.updated_at FROM entries e`
	args := []any{}
	wheres := []string{}
	if f.Kind != "" {
		wheres = append(wheres, "e.kind = ?")
		args = append(args, f.Kind)
	}
	if f.Language != "" {
		wheres = append(wheres, "e.language = ?")
		args = append(args, f.Language)
	}
	if f.Domain != "" {
		wheres = append(wheres, "e.domain = ?")
		args = append(args, f.Domain)
	}
	if f.Project != "" {
		wheres = append(wheres, "e.project = ?")
		args = append(args, f.Project)
	}
	if f.Tag != "" {
		query += ` JOIN entry_tags et ON et.entry_id = e.id JOIN tags t ON t.id = et.tag_id`
		wheres = append(wheres, "t.name = ?")
		args = append(args, f.Tag)
	}
	if len(wheres) > 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}
	query += " ORDER BY e.title"

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Slug, &e.Title, &e.Description, &e.Kind, &e.Language, &e.Domain, &e.Project, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		tags, err := getTagsForEntry(ctx, d.db, e.ID)
		if err != nil {
			return nil, fmt.Errorf("get tags for entry %d: %w", e.ID, err)
		}
		e.Tags = tags
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// SearchEntries runs FTS5 search with optional filters, returns entries with snippets (no full content).
func (d *DB) SearchEntries(ctx context.Context, queryStr string, f Filter, limit int) ([]Entry, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	q := `SELECT e.id, e.slug, e.title, e.description, e.kind, e.language, e.domain, e.project, e.version, e.created_at, e.updated_at,
	             snippet(entries_fts, 2, '>>>', '<<<', '...', 32) as snip
	      FROM entries_fts fts
	      JOIN entries e ON e.id = fts.rowid`
	args := []any{}
	wheres := []string{"fts.entries_fts MATCH ?"}
	args = append(args, queryStr)

	if f.Kind != "" {
		wheres = append(wheres, "e.kind = ?")
		args = append(args, f.Kind)
	}
	if f.Language != "" {
		wheres = append(wheres, "e.language = ?")
		args = append(args, f.Language)
	}
	if f.Domain != "" {
		wheres = append(wheres, "e.domain = ?")
		args = append(args, f.Domain)
	}
	if f.Project != "" {
		wheres = append(wheres, "e.project = ?")
		args = append(args, f.Project)
	}
	if f.Tag != "" {
		q += ` JOIN entry_tags et ON et.entry_id = e.id JOIN tags t ON t.id = et.tag_id`
		wheres = append(wheres, "t.name = ?")
		args = append(args, f.Tag)
	}
	q += " WHERE " + strings.Join(wheres, " AND ")
	q += " ORDER BY rank LIMIT ?"
	args = append(args, limit)

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	now := time.Now().UTC().Format(time.DateTime)
	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Slug, &e.Title, &e.Description, &e.Kind, &e.Language, &e.Domain, &e.Project, &e.Version, &e.CreatedAt, &e.UpdatedAt, &e.Snippet); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		tags, err := getTagsForEntry(ctx, d.db, e.ID)
		if err != nil {
			return nil, fmt.Errorf("get tags for entry %d: %w", e.ID, err)
		}
		e.Tags = tags
		entries = append(entries, e)
		// Bump search stats (best-effort)
		if _, err := d.db.ExecContext(ctx, `UPDATE entry_stats SET searches = searches + 1, last_search_at = ? WHERE entry_id = ?`, now, e.ID); err != nil {
			slog.Debug("update search stats", "err", err, "entry_id", e.ID)
		}
	}
	return entries, rows.Err()
}

// GetEntriesByContext returns full entries matching the given filters (language, domain, kind, tags, project).
func (d *DB) GetEntriesByContext(ctx context.Context, f Filter, limit int) ([]Entry, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q := `SELECT e.id, e.slug, e.title, e.description, e.content, e.kind, e.language, e.domain, e.project, e.version, e.created_at, e.updated_at FROM entries e`
	args := []any{}
	wheres := []string{}

	if f.Kind != "" {
		wheres = append(wheres, "e.kind = ?")
		args = append(args, f.Kind)
	}
	if f.Language != "" {
		wheres = append(wheres, "e.language = ?")
		args = append(args, f.Language)
	}
	if f.Domain != "" {
		wheres = append(wheres, "e.domain = ?")
		args = append(args, f.Domain)
	}
	if f.Project != "" {
		wheres = append(wheres, "e.project = ?")
		args = append(args, f.Project)
	}
	// Tags filter: entry must have ALL specified tags
	if len(f.Tags) > 0 {
		for i, tag := range f.Tags {
			alias := fmt.Sprintf("et%d", i)
			talias := fmt.Sprintf("t%d", i)
			q += fmt.Sprintf(` JOIN entry_tags %s ON %s.entry_id = e.id JOIN tags %s ON %s.id = %s.tag_id`, alias, alias, talias, talias, alias)
			wheres = append(wheres, talias+".name = ?")
			args = append(args, tag)
		}
	} else if f.Tag != "" {
		q += ` JOIN entry_tags et ON et.entry_id = e.id JOIN tags t ON t.id = et.tag_id`
		wheres = append(wheres, "t.name = ?")
		args = append(args, f.Tag)
	}

	if len(wheres) > 0 {
		q += " WHERE " + strings.Join(wheres, " AND ")
	}
	q += " ORDER BY e.title LIMIT ?"
	args = append(args, limit)

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("get by context: %w", err)
	}
	defer rows.Close()

	now := time.Now().UTC().Format(time.DateTime)
	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Slug, &e.Title, &e.Description, &e.Content, &e.Kind, &e.Language, &e.Domain, &e.Project, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		tags, err := getTagsForEntry(ctx, d.db, e.ID)
		if err != nil {
			return nil, fmt.Errorf("get tags for entry %d: %w", e.ID, err)
		}
		e.Tags = tags
		entries = append(entries, e)
		// Bump read stats (best-effort)
		if _, err := d.db.ExecContext(ctx, `UPDATE entry_stats SET reads = reads + 1, last_read_at = ? WHERE entry_id = ?`, now, e.ID); err != nil {
			slog.Debug("update read stats", "err", err, "entry_id", e.ID)
		}
	}
	return entries, rows.Err()
}

// ListTags returns all tags with their entry counts.
func (d *DB) ListTags(ctx context.Context) ([]Tag, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT t.name, COUNT(et.entry_id) as cnt FROM tags t JOIN entry_tags et ON et.tag_id = t.id GROUP BY t.id ORDER BY cnt DESC, t.name`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.Name, &t.Count); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// GetStats returns usage statistics for an entry.
func (d *DB) GetStats(ctx context.Context, slug string) (*EntryStats, error) {
	s := &EntryStats{}
	err := d.db.QueryRowContext(ctx,
		`SELECT es.reads, es.searches, es.updates, es.last_read_at, es.last_search_at, es.last_update_at
		 FROM entry_stats es JOIN entries e ON e.id = es.entry_id WHERE e.slug = ?`, slug,
	).Scan(&s.Reads, &s.Searches, &s.Updates, &s.LastReadAt, &s.LastSearchAt, &s.LastUpdateAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entry not found: %s: %w", slug, ErrNotFound)
		}
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return s, nil
}

// AllEntries returns all entries with full content and tags (for export).
func (d *DB) AllEntries(ctx context.Context) ([]Entry, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, slug, title, description, content, kind, language, domain, project, version, created_at, updated_at
		 FROM entries ORDER BY slug`,
	)
	if err != nil {
		return nil, fmt.Errorf("all entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Slug, &e.Title, &e.Description, &e.Content, &e.Kind, &e.Language, &e.Domain, &e.Project, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		tags, err := getTagsForEntry(ctx, d.db, e.ID)
		if err != nil {
			return nil, fmt.Errorf("get tags for entry %d: %w", e.ID, err)
		}
		e.Tags = tags
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// IsLocked returns true if the database write lock is active.
func (d *DB) IsLocked(ctx context.Context) (bool, error) {
	var active int
	if err := d.db.QueryRowContext(ctx, `SELECT active FROM lock WHERE id = 1`).Scan(&active); err != nil {
		return false, fmt.Errorf("check lock: %w", err)
	}
	return active == 1, nil
}

// Lock activates the write lock with the given token. Fails if already locked.
func (d *DB) Lock(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token must not be empty")
	}
	locked, err := d.IsLocked(ctx)
	if err != nil {
		return err
	}
	if locked {
		return fmt.Errorf("database is already locked: %w", ErrLocked)
	}
	hashed := hashToken(token)
	_, err = d.db.ExecContext(ctx, `UPDATE lock SET active = 1, token = ? WHERE id = 1`, hashed)
	return err
}

// Unlock deactivates the write lock. The provided token must match the one used to lock.
func (d *DB) Unlock(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token must not be empty")
	}
	locked, err := d.IsLocked(ctx)
	if err != nil {
		return err
	}
	if !locked {
		return fmt.Errorf("database is not locked")
	}
	var storedHash string
	if err := d.db.QueryRowContext(ctx, `SELECT token FROM lock WHERE id = 1`).Scan(&storedHash); err != nil {
		return fmt.Errorf("read lock: %w", err)
	}
	if storedHash != hashToken(token) {
		return fmt.Errorf("invalid token")
	}
	_, err = d.db.ExecContext(ctx, `UPDATE lock SET active = 0, token = '' WHERE id = 1`)
	return err
}

// --- helpers ---

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// querier is satisfied by *sql.DB and *sql.Tx for context-aware queries.
type querier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

// setTags replaces all tags for an entry within a transaction.
func setTags(ctx context.Context, tx *sql.Tx, entryID int64, tags []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM entry_tags WHERE entry_id = ?`, entryID); err != nil {
		return err
	}
	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO tags (name) VALUES (?)`, tagName); err != nil {
			return err
		}
		var tagID int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?`, tagName).Scan(&tagID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO entry_tags (entry_id, tag_id) VALUES (?, ?)`, entryID, tagID); err != nil {
			return err
		}
	}
	return nil
}

// getTagsForEntry returns all tag names for a given entry.
func getTagsForEntry(ctx context.Context, q querier, entryID int64) ([]string, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT t.name FROM tags t JOIN entry_tags et ON et.tag_id = t.id WHERE et.entry_id = ? ORDER BY t.name`, entryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	if tags == nil {
		tags = []string{}
	}
	return tags, rows.Err()
}
