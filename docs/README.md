# Technical Overview

`pocket` is a minimal Go CLI that implements a terminal-based file clipboard.
It stores file/directory references in `~/.pocketrc` (JSON) and can later
**copy** or **move** them to the current working directory.

## Architecture

```
main.go           → entry point
cmd/root.go       → Cobra command definition
pocket/pocket.go  → Core logic (store, load, copy, move, delete)
~/.pocketrc       → JSON array of paths (user-local persistent state)
```

## State

All state lives in `~/.pocketrc`. The file is a plain JSON array of
string paths:

```json
["/home/user/report.pdf", "/home/user/docs/"]
```

The file is created on first use with permissions `0600`.

## Build

```bash
# From the project root
go build -o pocket .

# Cross-compile (example)
GOOS=linux GOARCH=amd64 go build -o pocket-linux-amd64 .
```

## Dependencies

- [Cobra](https://github.com/spf13/cobra) — CLI framework
- Go standard library only for core logic (no external deps for copy/move)

## Flags Reference

| Short | Long       | Arg  | Action                              |
|-------|------------|------|-------------------------------------|
| (none)| (none)     | path(s) | Add paths to clipboard            |
| `-r`  | `--release`| —    | Copy all items to current directory |
| `-c`  | `--cut`    | —    | Move instead of copy (with `-r`)    |
| `-l`  | `--list`   | —    | Show numbered list                  |
| `-d N`| `--delete N`| int | Remove item #N from clipboard       |

## License

EUPL v1.2 — see [LICENSE](../LICENSE).
