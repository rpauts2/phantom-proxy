# Go-only migration plan

## Current state

The repository is **not yet a pure-Go application**. The Go module is present, but the current repository also contains Python API/AI/Celery services, a Next.js frontend, Python maintenance scripts, and a Docker Compose stack that starts several non-Go runtimes. The existing v14 command also contains placeholder implementations, so a successful compile must not be treated as proof that the advertised features work.

This plan deliberately leaves `config.yaml` unchanged, as requested.

A first isolated Go-only component now exists at `cmd/simulation-lab` with its event store and HTTP API in `internal/simulation`. It binds to a loopback IP only, stores a bounded in-memory history, accepts only `visit`, `click`, and `training_complete` events, and rejects unknown JSON fields. It is a safe prototype, not yet a replacement for the existing application. Repository tests have been added, but could not be executed from the current environment because GitHub could not be reached from the local shell.

## Target

One Go runtime and one canonical Go entrypoint for the application. External infrastructure such as a database or a separately installed local model may remain optional adapters, but no Python, Node.js, or Celery process should be required to build, run, test, or administer the core application.

## Phases

1. **Inventory and baseline**
   - Keep `cmd/phantom-proxy-v14` as the provisional entrypoint until its call graph and behavior are verified.
   - Run `go test -race ./...`, `go vet ./...`, and `go build ./cmd/phantom-proxy-v14` in an environment with Go 1.24+.
   - Record failures instead of claiming a clean build when they have not been run.
   - Mark stubs and duplicated entrypoints; don't silently route production traffic to placeholder implementations.

2. **Go-only developer workflow**
   - Make Go module download, formatting, tests, vetting, and build the default Makefile workflow.
   - Keep the old multi-runtime Compose stack explicitly labeled as legacy until it is replaced.
   - Align the documented Go version with `go.mod` (currently Go 1.24.0).

3. **Safe application boundary**
   - Make the supported product a controlled security-awareness/simulation lab.
   - Use synthetic identities and seeded test events only. Record visits, clicks, and training completion, not real passwords, session cookies, or bearer tokens.
   - Enforce an explicit lab allowlist, scope expiry, audit trail, and an emergency stop.
   - Do not enable credential capture, evasion, or remote-command execution as part of the migration.

4. **Replace non-Go services**
   - Move supported API routes into a Go `net/http` service with explicit request/response types, bounded request sizes, timeouts, and health/readiness endpoints.
   - Replace Celery tasks with a bounded Go worker queue using contexts and graceful shutdown.
   - Replace Python backup/health scripts with Go subcommands only where their behavior is specified and tested.
   - Replace the Next.js dependency with a Go-served static control panel, or document that the UI is an optional separate product rather than claiming the entire stack is pure Go.
   - Keep LLM access behind a small Go interface. An external model server can be optional, but it must not be required for core startup.

5. **Reliability and release gates**
   - Add unit tests for config parsing (without editing the live config file), API authorization, input validation, shutdown, migrations, and simulation scope enforcement.
   - Add CI for formatting, vet, tests, dependency review, and build only after the supported safe application boundary is established.
   - Pin runtime/build versions and verify the final image contains only the intended Go binary, certificates, and explicitly required data files.

## Definition of done

- A clean checkout builds with Go 1.24+ and no Python/Node/Celery installation.
- `make test`, `make lint`, and `make build` use Go tooling only.
- The supported launch path starts one canonical Go binary.
- Legacy Compose services are either removed or clearly optional and no longer presented as part of the Go-only product.
- Tests demonstrate that simulation scope restrictions and shutdown behavior work.
- No claim of full Go-only status is made until the above checks have actually passed.

## Run the safe simulation prototype

```sh
make lab
```

The service listens on `127.0.0.1:8088` by default and refuses non-loopback bind addresses.

- `GET /healthz`: process health.
- `GET /api/v1/events`: recent synthetic events.
- `POST /api/v1/events`: add an event with `campaign_id`, `participant_id`, and `type`. The only accepted event types are `visit`, `click`, and `training_complete`.

Example event:

```sh
curl -X POST http://127.0.0.1:8088/api/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"campaign_id":"demo","participant_id":"person_01","type":"click"}'
```

Events are held in memory and disappear when the process exits. This prototype intentionally does not collect credentials, cookies, tokens, arbitrary payloads, or external redirect targets.
