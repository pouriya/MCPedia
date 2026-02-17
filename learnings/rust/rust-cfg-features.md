---
title: "Rust cfg and Features"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, features, cfg, conditional]
description: "Conditional compilation and Cargo features in Rust"
---

# Rust cfg and Features

Use `#[cfg(...)]` for conditional compilation and Cargo features to make dependencies and code optional. Keep feature sets small and documented.

- Use `#[cfg(target_os = "linux")]`, `#[cfg(feature = "json")]`, `#[cfg(test)]` to include or exclude code. Use `cfg!()` macro for runtime checks in expressions. Use `#[cfg_attr(feature = "x", derive(Serialize))]` for conditional derives.
- In `Cargo.toml`, define features under `[features]` (e.g. `default = ["std"]`, `full = []`). Add optional dependencies with `optional = true` and list them in features (e.g. `serde = ["dep:serde"]`). Document features in the crate docs and README.
- Avoid feature creep; prefer a small set of coherent features. Do not use features to work around semver (e.g. breaking changes). Use `--no-default-features` in CI or when testing minimal builds. Ensure the crate compiles with default features and with key feature combinations.
