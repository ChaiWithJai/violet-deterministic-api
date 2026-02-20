# Skill: Codebase Archaeology

Analyze repository history to understand architecture, hotspots, ownership, and risk before major changes.

## When to use
1. New contributor onboarding.
2. Pre-migration planning.
3. Before large refactors.

## Workflow
1. Identify hotspots.
```bash
git log --pretty=format: --name-only | sort | uniq -c | sort -rn | head -30
```
2. Identify ownership and bus factor.
```bash
git shortlog -sn
git shortlog -sn -- path/to/file
```
3. Trace subsystem evolution.
```bash
git log --follow --oneline -- path/to/file
git log --oneline --grep='refactor\|rewrite\|migration\|deprecat'
```
4. Map recurring bug zones.
```bash
git log --oneline --grep='fix\|bug\|incident\|hotfix' --all | head -80
rg -n 'TODO|FIXME|HACK|XXX' .
```

## Required artifacts
1. `docs/research/codebase-archaeology-report.md` with hotspots and risk map.
2. ADR candidates list for major turning points.
3. Migration risk notes referenced by ticket IDs.
