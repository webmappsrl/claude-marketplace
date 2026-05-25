---
name: our-code-style
description: Apply Acme team's code style conventions. Use whenever writing or reviewing TypeScript/Python code in any Acme repository, especially when creating new files, refactoring, or before committing.
---

# Acme Team Code Style

Apply these conventions to all code written in Acme repositories.

## TypeScript

- Use `type` aliases over `interface` except for public API contracts
- Always prefer `const` over `let`; never use `var`
- Named exports only — no default exports in shared modules
- Imports grouped: stdlib, third-party, internal (`@/...`), relative
- File names: `kebab-case.ts` for modules, `PascalCase.tsx` for React components

## Python

- Python 3.11+ syntax (PEP 604 unions: `str | None`, not `Optional[str]`)
- Always type-annotate function signatures
- Format with `ruff format`, lint with `ruff check --fix`
- Docstrings: Google style, only on public functions and classes

## Naming

- Variables and functions: descriptive, no abbreviations (use `request` not `req`)
- Booleans prefixed: `is_`, `has_`, `should_`, `can_`
- Constants `SCREAMING_SNAKE_CASE`

## Comments

- Comments explain *why*, not *what*. If the code needs a "what" comment, refactor first.
- TODO format: `# TODO(jira-ticket-id): description`

## Before suggesting code

1. Read the existing file (if any) to match its conventions
2. Check `package.json` / `pyproject.toml` for the project's lint config — those rules override these defaults
