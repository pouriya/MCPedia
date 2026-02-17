---
title: "Go Modules"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, modules, go.mod, dependencies]
description: "Go modules and dependency management"
---

# Go Modules

Use Go modules for dependency management. Keep `go.mod` and `go.sum` in version control and use reproducible builds.

- Initialize with `go mod init <module path>`. The module path should match the repo (e.g. `github.com/org/repo`). Add dependencies by importing in code and running `go build` or `go mod tidy`.
- Use `go get package@version` to add or upgrade; use `go mod tidy` to drop unused and fix requirements. Pin versions in `go.mod`; avoid `replace` except for local development or forks.
- Use semantic versions for releases; consumers use `go get module@v1.2.3`. Follow semver for public modules: no breaking changes without a major version bump. Use `internal/` for private packages.
- Run `go mod verify` and keep `go.sum` committed. In CI, use `go mod download` and cache the module cache. Prefer minimal dependencies and well-maintained libraries.
