---
title: "Go database/sql"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, sql, database, queries]
description: "Using database/sql for SQL databases in Go"
---

# Go database/sql

Use `database/sql` with a driver (e.g. `lib/pq`, `go-sqlite3`) for SQL databases. Use context for timeouts and cancellation; use prepared statements and connection pooling correctly.

- Open with `sql.Open(driver, dsn)`; get a `*sql.DB` and call `db.Ping()` or `db.PingContext(ctx)` to verify. Use `db.SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime` for pool tuning. Always close with `defer db.Close()`.
- Use `db.QueryContext(ctx, sql, args...)` for multiple rows; defer `rows.Close()` and iterate with `rows.Next()` and `rows.Scan()`. Use `db.QueryRowContext` for at most one row; check `ErrNoRows`. Use `db.ExecContext` for INSERT/UPDATE/DELETE. Use `?` or driver-specific placeholders; never concatenate user input.
- Use prepared statements with `db.PrepareContext` when repeating the same query; use `stmt.QueryContext`/`ExecContext`. For transactions use `db.BeginTx(ctx, opts)` and commit or rollback. Pass context so long-running queries can be cancelled.
