# Legacy component inventory

## Purpose and status

This inventory is a migration gate, not a claim that every file below has been reviewed line by line. Paths are categorized from the checked-in tree and service declarations. Before a legacy component is brought into the supported runtime, its actual behavior, data flows, external network access, and tests must be inspected.

The supported runtime remains `cmd/simulation-lab` plus `internal/simulation`. It accepts synthetic awareness events only, binds to loopback, and uses the Go standard library.

## Runtime and user-interface inventory

| Area | Repository paths | Status |
|---|---|---|
| Supported Go lab | `cmd/simulation-lab/`, `internal/simulation/` | Supported and CI-tested |
| Legacy Go entrypoints | `cmd/phantom-proxy-v14/`, `cmd/phantom-proxy/`, `cmd/gendert/` | Audit before use; not part of the supported launch path |
| Python API and background jobs | `api/`, `api/celery_app.py`, `api/app/tasks.py` | Legacy runtime; migration incomplete |
| Python AI service | `ai_service/` | Legacy runtime; migration incomplete |
| Python maintenance and tests | `backup.py`, `tests/test_all.py` | Behavior must be specified and tested before replacement |
| Web frontend | `frontend/` including Next.js configuration and `package.json` | Node.js runtime; separate from supported Go lab |
| Multi-runtime deployment | `docker-compose.yml`, root and service Dockerfiles, `requirements.txt` files | Legacy deployment path; not a Go-only deployment |

## Components requiring explicit security review

The following areas must not be copied into the synthetic lab merely to make the migration compile. The path names identify review areas, not a verdict that every implementation is unsafe.

| Paths or packages | Gate before migration |
|---|---|
| `core/proxy/`, `internal/session/`, `internal/auth/`, `internal/browser/` | Review request routing, session material, browser control, and data retention. Only synthetic test data and explicitly authorized systems may be supported. |
| `internal/credentialstuffing/`, `internal/evasion/`, `internal/c2/`, `internal/vishing/` | Do not port credential attacks, stealth/evasion, remote-command control, or impersonation workflows into the supported product. Replace with harmless stubs or remove from the eventual supported build. |
| `internal/domain/`, `internal/email/`, `internal/campaign/`, `internal/decentral/` | Review external side effects, domain/email operations, campaign targets, and scope enforcement. Require explicit allowlists and dry-run behavior for any retained administrative utility. |
| `internal/ai/`, `internal/ganobf/`, `internal/mlopt/` | Review prompts, generated code or content, subprocesses, network access, and model dependencies before deciding whether each capability belongs in the supported runtime. |

## Migration order

1. Keep the safe lab as the only documented default entrypoint.
2. For each legacy subsystem, record its public interfaces, dependencies, side effects, and test coverage before deciding whether it can be retained.
3. Move only benign, fully specified functionality into a Go package with bounded inputs, context-aware cancellation, explicit network scope, and tests.
4. Keep external effects behind opt-in adapters, disabled by default. Do not expose credential/session collection, evasion, or remote-command execution through the supported runtime.
5. Delete legacy files only after references, deployment manifests, documentation, and tests are checked. Do not delete by filename alone.
6. Claim “pure Go” only when a clean checkout can build, run, test, and administer the supported product without Python, Node.js, or Celery.

## Release gate

The current CI job validates the simulation package and its CLI only. It is not a security audit and does not mean the legacy entrypoints or multi-runtime stack are ready for production.
