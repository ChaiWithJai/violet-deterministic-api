# Skill: Platform Verification

Verify behavior on the actual local platform and avoid claims about untested environments.

## Workflow
1. Record machine context (OS, architecture, runtime versions).
2. List what can and cannot be validated locally.
3. Execute explicit verification steps.
4. Record expected vs actual outcomes.

## Mandatory checks for this repo
1. `docker compose -f docker-compose.demo.yml config`
2. `docker compose -f docker-compose.demo.yml up -d`
3. `GET /v1/health`
4. deterministic behavior checks for `POST /v1/decisions` and `POST /v1/replay`

## Output
Update runbook evidence in `docs/runbooks/local-smoke.md` with date-stamped results.
