# escope — AI Assistant Rules

Applies to Cursor, Copilot, Claude Code, etc.

## Rules

1. Cobra `--help`
    - Always fill: `Short`, `Long`, `Example` (if needed)
    - Session/persistence: explain in `Long` that data lives in the **host config file** under the `sessions` map keyed by **Elasticsearch host URL**. Mention **ctrl+s** for calculator snapshot save. Do not hard-code file paths or product names like “TUI” in user-facing strings.
    - For `escope calculator`: no flags = **restore last ctrl+s snapshot if present**, else built-in defaults; **`--from-cluster`** = live cluster pre-fill (optional default index from `escope index use`); **`--snapshot`** = stored snapshot only (error if missing); document mutual exclusion with **`--clear`** / the other source flag. Fields: **data nodes**, **rps**, **RAM/disk per data node (GiB)**; cache-fit metrics are computed in the UI.

2. README
    - Update Command Reference table
    - Add examples in existing format/sections
    - Do not change structure unnecessarily

3. Makefile (`test-commands`)
    - Add only non-blocking commands (`--help`, `--clear`, etc.)
    - Do not add interactive session commands

4. Persistence
    - Use **`internal/config`** `HostConfig.Sessions` for default index and calculator snapshot; host URL scoping; `--clear` removes the relevant block
    - Calculator initial load: **`--from-cluster`** vs **`--snapshot`** vs **defaults (no flags)** are distinct; do not conflate in docs

5. Code
    - No unnecessary refactor
    - Follow existing style
    - Do not add new markdown files (except README / this file) unless requested
    - Do not add comments unless explicitly requested
    - Do not use magic string/number, move to constants classes

6. Updates
    - Keep order
    - Add, don’t duplicate
    - Keep rules concise
