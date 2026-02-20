# Violet Rails Deprecation Plan

## Preconditions
1. 30-day SLO compliance on new service.
2. Deterministic replay mismatch <= 0.5% for migrated flows.
3. Security controls complete (authn/authz/tenant isolation).

## Migration strategy
1. New tenants on new service only.
2. Existing tenants migrate via export/import + verification reports.
3. Violet enters maintenance-only mode.
4. Final sunset with rollback archive and read-only access window.
