---
title: "Rust Project Structure"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, structure, modules, cargo]
description: "Organizing Rust crates and modules"
---

# Rust Project Structure

Organize code into crates and modules for clarity and reuse. Use Cargo workspaces for multi-crate repos.

- Use `mod` to split a crate into modules; expose public API with `pub` and `pub use`. Prefer one main theme per module (e.g. `error`, `config`, `handlers`). Use `mod.rs` or `module_name.rs` plus `module_name/` for submodules.
- Put library code in `src/lib.rs` and binary entry in `src/main.rs`; binaries can depend on the library with `use crate_name::...`. Use `src/bin/` for multiple binaries.
- Use `internal` or `crate` visibility to keep helpers non-public. Re-export important types at the crate root for a clean public API.
- In workspaces, use path dependencies and shared `Cargo.toml` settings. Keep crate boundaries clear: core logic in a lib, binaries and optional features in separate crates.
