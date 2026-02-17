---
title: "Go HTTP Middleware"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, middleware, http, handlers]
description: "Writing and chaining HTTP middleware in Go"
---

# Go HTTP Middleware

Middleware wraps handlers to add logging, auth, recovery, or CORS. Use a consistent pattern: accept `http.Handler` and return `http.Handler`; call the next handler in the chain.

- Type: `func(next http.Handler) http.Handler`. Inside, return `http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ... next.ServeHTTP(w, r) })`. Run logic before and/or after the next handler. Do not forget to call `next.ServeHTTP` unless you intentionally short-circuit (e.g. auth failure).
- Chain with a helper: `chain := func(h http.Handler) http.Handler { return logging(recovery(cors(h))) }`. Apply once to your root handler or router. Use `r.Context()` to pass request-scoped values (e.g. request ID, user) set by earlier middleware.
- Keep middleware focused: one concern per middleware. Log requests and response status/size; recover panics and return 500; add CORS headers; validate auth and return 401/403. Test middleware by wrapping a test handler and asserting on calls and response.
