---
title: "Rust Serde and JSON"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, serde, json, serialization]
description: "Serialization and deserialization with Serde in Rust"
---

# Rust Serde and JSON

Use Serde for serialization and deserialization. Combine with `serde_json` for JSON in APIs, configs, and storage.

- Derive `Serialize` and `Deserialize` with `#[derive(serde::Serialize, serde::Deserialize)]`. Use `serde_json::to_string`/`from_str` or `to_vec`/`from_slice` for JSON bytes. Use `serde_json::Value` when the shape is dynamic.
- Use attributes for control: `#[serde(rename = "...")]`, `#[serde(default)]`, `skip_serializing_if`, `with` for custom (de)serialization. Use `Option` for optional fields; avoid panics on missing or unknown fields in configs.
- Validate after deserialization when needed (e.g. business rules). Return clear errors for invalid input. Prefer typed structs over raw `Value` when the schema is known.
- In web APIs, use `axum::Json` with Serde types; handle errors and return appropriate status codes and error bodies.
