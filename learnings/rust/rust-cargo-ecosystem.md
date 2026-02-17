---
title: "Rust Cargo and Ecosystem"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, cargo, clippy, fmt, dependencies]
description: "Cargo, tooling, and dependency management for Rust"
---

# Rust Cargo and Ecosystem

Use Cargo for building, testing, and dependency management. Keep projects consistent with standard tooling and dependency hygiene.

- Use `cargo build`, `cargo test`, `cargo run`, and `cargo check`. Use `--release` for optimized builds. Add commonly used dev-dependencies (e.g. `cargo test`, `cargo clippy`) to `Cargo.toml`.
- Run `cargo fmt` (and enforce in CI) for consistent style. Run `cargo clippy` and address warnings; enable lints in `Cargo.toml` or `clippy.toml` as needed.
- Prefer minimal, well-maintained dependencies. Pin versions in `Cargo.toml`; use `cargo update` deliberately. Prefer `crates.io` and documented, widely used crates.
- Use workspaces for multi-crate repos. Use `[patch]` or path dependencies for local development; avoid committing large binaries or generated code.
