# Seed learnings into mcpedia from per-language markdown files (export format).
# Override for custom binary or database:
#   make build mcpedia_exe=./mcpedia mcpedia_db=./mcpedia.db

mcpedia_exe ?= mcpedia
mcpedia_db ?= mcpedia.db

LANGUAGES := rust golang

.PHONY: build
build:
	@set -e; \
	for lang in $(LANGUAGES); do \
		dir="$$lang"; \
		if [ -d "$$dir" ]; then \
			for f in $$dir/*.md; do \
				[ -f "$$f" ] || continue; \
				$(mcpedia_exe) import --db $(mcpedia_db) --file "$$f"; \
			done; \
		fi; \
	done
