---
title: "Rust Async and Tokio"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, async, tokio, concurrency]
description: "Async Rust and Tokio runtime patterns for AI agents"
---

# Rust Async and Tokio

Use async/await with the Tokio runtime for I/O-bound concurrency. Keep the event loop non-blocking and offload CPU-bound work when needed.

- Mark async functions with `async fn` and run them with an executor (e.g. `tokio::spawn`, `#[tokio::main]`). Use `.await` for futures; avoid blocking the thread (no `std` blocking calls in async code).
- Prefer `tokio` types for I/O: `tokio::fs`, `tokio::net`, `tokio::sync`. Use channels (`mpsc`, `broadcast`) for communication between tasks.
- For CPU-heavy work, spawn blocking tasks with `tokio::task::spawn_blocking` so the async runtime is not starved.
- Use `select!` and `tokio::time::timeout` for cancellation and time limits. Prefer structured concurrency (e.g. scoped tasks) where applicable.
