---
title: "Rust Logging and Tracing"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, logging, tracing, observability]
description: "Structured logging and tracing in Rust"
---

# Rust Logging and Tracing

Use the `log` crate for application logging and `tracing` for structured spans and events in async and service code.

- Initialize a logger (e.g. `env_logger`, `tracing_subscriber`) early in `main`. Use `log::info!`, `error!`, `debug!`, etc., or `tracing::info!` with optional key-value fields. Avoid logging in hot paths or at excessive verbosity in production.
- With tracing: use `#[tracing::instrument]` on functions to create spans; use `tracing::info_span!` and `event!` for structured context. Propagate context in async with `tracing::Instrument` or layer support. Use trace IDs for request correlation.
- Do not log secrets or PII. Use structured fields (key-value) so logs can be queried. Configure level and format (e.g. JSON) per environment. In libraries, use `log` or `tracing` as a facade so the application chooses the implementation.
