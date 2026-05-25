---
name: our-pr-checklist
description: Acme team's pull request checklist. Use before opening a PR or marking work as done, to ensure description, tests, screenshots, and reviewer assignment all meet team standards.
---

# Acme Team PR Checklist

Run through this checklist before opening or marking a PR ready for review.

## PR description must include

- **What**: one-paragraph summary of the change
- **Why**: link to the Linear/Jira ticket or RFC
- **How to test**: explicit steps a reviewer can follow locally
- **Screenshots/recordings**: for any UI change
- **Risk**: low / medium / high, with reasoning if medium or high

## Code requirements

- All new code has tests (unit minimum; integration if it crosses a service boundary)
- `npm run lint` and `npm run typecheck` pass locally
- No `console.log`, `print`, `dbg!`, or commented-out code left in
- No new dependencies without justification in the PR description

## Reviewers

- Tag at least one engineer from the owning team (see `CODEOWNERS`)
- For changes to `infra/`, `auth/`, or `billing/`: tag a senior engineer
- For schema migrations: tag the data team

## Size

- Target under 400 lines changed. If larger, split into stacked PRs or explain why splitting isn't possible in the description.

## Before requesting review

Confirm the branch is rebased on the latest `main` and CI is green.
