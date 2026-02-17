---
title: "Rust Web APIs with Axum"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, axum, http, api]
description: "Building HTTP APIs in Rust with Axum"
---

# Rust Web APIs with Axum

Axum is a popular async web framework for Rust on top of Tokio and Tower. Use it for type-safe, composable HTTP APIs.

- Define handlers as async functions taking extractors (e.g. `Json`, `Path`, `Query`, `State`) and returning types that implement `IntoResponse`. Use `Result<T, E>` and map errors to status codes.
- Use the router with `.route()` and `.nest()`; apply middleware (e.g. logging, CORS, auth) with `.layer()`. Share app state via `State` and Arc.
- Prefer `axum::Json` for JSON bodies; use `serde` for serialization. Validate input and return clear error responses (4xx/5xx) with appropriate body.
- Run with Tokio; use `tower_http` or similar for timeouts, compression, and tracing. Test with `axum::test` or integration tests against a test server.
