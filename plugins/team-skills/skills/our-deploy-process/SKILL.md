---
name: our-deploy-process
description: Acme team's deployment workflow. Use when the user mentions deploying, releasing, promoting to staging/production, rolling back, or running migrations.
---

# Acme Team Deploy Process

## Environments

- `dev` — auto-deploys from `main` on every merge
- `staging` — manual promote from a green `dev` build via `./scripts/promote staging`
- `prod` — manual promote from a green `staging` build, requires approval

## Before deploying to prod

1. Confirm staging has been running the candidate build for at least 30 minutes with no error-rate regression in Datadog dashboard `acme-prod-health`
2. Check `#deploys` Slack channel for any active incident or freeze
3. Announce in `#deploys`: "Deploying <PR link> to prod in 5 min"

## Running the deploy

```bash
./scripts/promote prod
```

The script will:
- Tag the release in git
- Trigger the prod pipeline
- Post the deploy status to `#deploys`

## Database migrations

- Migrations run automatically as part of deploy
- For destructive migrations (DROP, ALTER COLUMN with data loss), use the two-phase pattern: deploy the additive change first, backfill, then deploy the destructive change in a separate PR
- Never deploy a migration on Friday after 14:00 CET

## Rollback

If error rate spikes within 15 minutes of deploy:

```bash
./scripts/rollback prod
```

Then post-mortem in `#incidents` within 24 hours using the template in Notion.
