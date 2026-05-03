# escope — AI Assistant Rules

Applies to Cursor, Copilot, Claude Code, etc.

## Global

- No unnecessary refactor; follow existing style. Touch foundational setup (`internal/connection`, `internal/elastic` client configuration) only for a clear bug fix or required behavior change, not for cosmetic edits.
- Do not introduce new environment variables for app configuration, secrets, or agent/shell automation unless the user explicitly asks; prefer the host config file and CLI flags.
- User-facing CLI strings, errors, `Cobra` help text, and edits to project markdown (README, `AGENTS.md`) are **English only** unless the user requests another language.
- Reuse `internal/connection` for shared connection state and `ApplyPersistentConnection` (same rules as root persistent flags). Do not add connection APIs for a single feature; keep feature-only helpers in that feature’s package (e.g. `internal/record` for daemon argv and log identity).
- Reuse `internal/ui` for tabular CLI output instead of ad-hoc printing.
- Do not add new markdown files (except README and this file) unless requested.
- Do not add comments unless explicitly requested.
- Do not use magic strings or numbers in code; move them to `internal/constants` (or the project’s existing constants pattern). Use **one** constant per Elasticsearch JSON field value (no duplicate names like `StoreField` and `StoreFieldKey` for the same key). Avoid meaningless numeric suffixes in names (`Thing2`, `Field3`); qualify by domain instead (e.g. shard listing vs node stats).
- When editing docs: keep section order, add rather than duplicate, stay concise.

## Commands (Cobra)

- Always fill `Short`, `Long`, and `Example` when a command is user-facing.
- In `Long`, for any feature that persists data: say it lives in the **host config file** and how it is keyed (e.g. host URL); do **not** hard-code file paths in user-facing strings.

## Host config and `sessions` (`internal/config`)

- Use `HostConfig.Sessions` keyed by **Elasticsearch host URL** for:
  - default index / alias (`escope index use`)
  - calculator snapshot (ctrl+s)
- `--clear` on the relevant command removes only that feature’s block where applicable.
- **Record** sampler state (PID + log path) is **not** in `sessions`; it uses a **separate small file** alongside the host config path.

## CLI output (tables vs TUI)

- Use a dedicated TUI (e.g. Bubble Tea) only when the command is intentionally built as an interactive full-screen experience.
- For ordinary list/detail output, follow sibling commands: shared table/formatter helpers under `internal/ui`, not ad-hoc column printing.

## Calculator (`escope calculator`)

- Document in `Long` / README: **ctrl+s** saves snapshot to `sessions` for that host URL.
- **No flags** = restore last ctrl+s snapshot if present, else built-in defaults.
- **`--from-cluster`** = live cluster pre-fill (optional default index from `escope index use`).
- **`--snapshot`** = stored snapshot only (error if missing).
- Document mutual exclusion: **`--clear`** must not combine with the other source flags.
- Fields called out in docs: **data nodes**, **rps**, **RAM/disk per data node (GiB)**; cache-fit metrics are computed in the UI.

## Record (`escope record`)

- **`start`**: same global connection flags as other Elasticsearch commands; **`stop`**: no cluster required.
- Session file holds PID + log path; not under `sessions` (see Host config above).
- **`--interval`**: sample period in **seconds**; default from constants.
- Child-process argv and log-header cluster labels are **`internal/record` only**, not `internal/connection`.

## README

- Update the Command Reference table when commands or important flags change.
- Add examples in the existing format and sections; do not change overall README structure unless necessary.

## Makefile (`test-commands`)

- Add only non-blocking targets (e.g. `--help`, `--clear`).
- Do not add interactive session or long-running TUI commands.
