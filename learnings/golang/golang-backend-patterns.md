---
title: "Go Backend Patterns"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, backend, api, services]
description: "Patterns for backend services and APIs in Go"
---

# Go Backend Patterns

Structure backend services with clear layers: handlers, business logic, and data access. Use context, interfaces, and consistent error handling.

- Keep HTTP handlers thin: parse request, validate input, call a service or use case, write response. Put business logic in packages that do not depend on HTTP. Use interfaces for repositories and external services so you can test and swap implementations.
- Use context for request scope and cancellation. Pass context from the handler into DB queries, HTTP client calls, and goroutines. Use timeouts (e.g. `context.WithTimeout`) for external calls and respect them in your code.
- Return errors from services; let the handler or middleware translate them to HTTP status codes and log. Use structured logging (e.g. slog, zap) with request ID or trace ID from context. Use middleware for auth, logging, and recovery.
- Use connection pooling for DB and HTTP clients. Configure timeouts, retries, and backoff for external dependencies. Prefer small, focused packages and avoid circular dependencies.
