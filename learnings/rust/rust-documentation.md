---
title: "Rust Documentation"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, docs, comments, examples]
description: "Documenting Rust crates and APIs"
---

# Rust Documentation

Document public API with doc comments. Use `cargo doc` to generate and read the output; add examples that run with `cargo test`.

- Use `///` for item docs and `//!` for crate or module docs. First line is a short summary; add a blank line and then details. Use Markdown: code in backticks, `# Examples` section, `# Panics`, `# Errors`, `# Safety` for unsafe.
- Use `# Examples` with code blocks; mark with ````ignore` or ````no_run` if the example does not run as-is. Use ````rust` and run with `cargo test` to keep doc examples compile- and run-correct. Document parameters and return values with `# Arguments` and inline backticks.
- Re-export and document at the crate root for a clear public API. Use `#[doc(hidden)]` for internal items that must be public but are not part of the stable API. Keep docs up to date when changing behavior.
