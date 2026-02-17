---
title: "Go Standard Library"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, stdlib, packages, io]
description: "Key Go standard library packages for AI agents"
---

# Go Standard Library

Prefer the standard library for common tasks. Know the main packages so you can suggest the right tool.

- **io / os**: `io.Reader`, `io.Writer`, `io.ReadAll`, `os.ReadFile`, `os.WriteFile`. Use `bufio` for buffered I/O and scanning. Use `os.Exec` and `exec.Cmd` for subprocesses.
- **net/http**: `http.ListenAndServe`, `http.Handler`, `http.HandlerFunc`, `httptest` for tests. Use `http.Client` with timeout and optional transport. Use `context` in requests.
- **encoding/json**: `json.Marshal`/`Unmarshal`, struct tags. Use `json.Decoder` for streams. Handle number types and optional fields with care.
- **sync**: `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Once`, `sync.Map` (when needed). Prefer channels for coordination when it fits. Use `sync.Pool` for reuse of short-lived objects.
- **testing**: `testing.T`, `testing.B`, table-driven tests, `httptest`, `iotest`. Use `testing/quick` or fuzz for property-style tests when useful.
