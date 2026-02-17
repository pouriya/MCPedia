---
title: "Rust CLI Tools"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, cli, argparse, stdio]
description: "Building command-line tools in Rust"
---

# Rust CLI Tools

Build fast, user-friendly CLI tools with clear arguments, help text, and error messages. Use `clap` for parsing and structured output for scripts.

- Use `clap` (derive or builder) for subcommands, flags, and options. Enable `derive` feature and document args with help strings. Validate and convert types (e.g. paths, numbers) during parsing.
- Read from stdin when appropriate; write to stdout for results and stderr for logs and errors. Use exit codes: 0 for success, non-zero for failure. Consider `atty` or similar to detect TTY and adjust behavior.
- Use `anyhow` or custom errors for failures; print clear messages to stderr and exit with a non-zero code. Use colored output (e.g. `termcolor`, `owo-colors`) sparingly and only when output is a TTY.
- For large outputs, stream when possible. Support `--quiet`, `--verbose`, and config files if the tool grows. Add shell completions (e.g. `clap_complete`) for better UX.
