# go-session

> **DEPRECATED — use `github.com/anatolykoptev/go-kit/session` instead.**
>
> Promoted into go-kit as the `session/` subpackage in **go-kit v0.44.0**
> (2026-05-04). The API surface is byte-identical; migration is a single
> import rewrite plus `go mod tidy`.
>
> This repo (`go-session`) is **frozen at v0.5.0** for compatibility
> with consumers that haven't migrated yet. New features go to
> `go-kit/session`.

## Migration

```diff
-import session "github.com/anatolykoptev/go-session"
+import session "github.com/anatolykoptev/go-kit/session"

-import sessionredis "github.com/anatolykoptev/go-session/redis"
+import sessionredis "github.com/anatolykoptev/go-kit/session/redis"
```

```bash
go get github.com/anatolykoptev/go-kit@latest
go mod tidy
```

The `archive/` subpackage (MemDB archival bridge) is **NOT migrated**
— it was unused at consolidation time, and bringing memdb-go as a
transitive dep into go-kit was deemed too heavy. If you need it, copy
the file into your service or vendor it explicitly.

## What this repo was

A small library for managing per-conversation chat history with:

- `Session{Key, Messages, Summary, Facts, Created, Updated}` and
  `Store` interface (memory / file / Redis backends).
- `Compactor` — calls a `SummarizeFn` (typically an LLM) to compress
  old turns into a `Summary` once the message count exceeds a
  threshold.
- `Fact` extraction — opt-in factoid distillation during compaction.
- Per-key locking, TTL/eviction (`cleanup.go`), intent filter
  (`intent.go`).

All of that now lives at `github.com/anatolykoptev/go-kit/session`,
unchanged.
