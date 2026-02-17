---
title: "Go HTTP Handlers"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, http, handlers, api]
description: "Writing HTTP handlers and APIs in Go"
---

# Go HTTP Handlers

Use the standard `net/http` package or a minimal framework. Keep handlers thin: parse input, call business logic, write response and status.

- Handlers have signature `func(w http.ResponseWriter, r *http.Request)`. Read body with `io.ReadAll(r.Body)` or a JSON decoder; set `Content-Type` and status with `w.Header().Set` and `w.WriteHeader`. Use `http.Error` for error responses.
- Use `r.Context()` for cancellation and timeouts; pass it to DB and downstream calls. Parse path and query with `mux` or `chi` if you need routing; use `r.URL.Query().Get` and path params from the router.
- Return appropriate status codes: 200/201 for success, 400 for bad input, 401/403 for auth, 404 for not found, 500 for server errors. Return JSON (or other format) consistently; use a helper for JSON responses and errors.
- Keep handlers short; delegate to services or packages. Use middleware for logging, auth, recovery, and CORS. Test handlers with `httptest.ResponseRecorder` and `httptest.NewRequest`.
