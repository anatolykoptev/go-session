# go-session Roadmap

## Phase 1: Core (this release)

- [x] Message, Session, Fact, ToolCall types
- [x] Store interface
- [x] InMemoryStore (TTL, maxMessages)
- [x] FileStore (atomic JSON, auto-load)
- [x] Compactor (summary + fact extraction, multi-part)
- [x] Intent filter (NeedsMemoryContext)
- [x] Stale cleanup with archive callback
- [x] Full test suite (80%+ coverage)

## Phase 2: Redis backend

- [ ] `redis/` sub-package with RedisStore
- [ ] Integration tests with testcontainers

## Phase 3: MemDB archival bridge

- [ ] `archive/memdb.go` — archive session to MemDB on cleanup
- [ ] Configurable archive format (facts + last N messages)

## Phase 4: Consumer migration

- [ ] dozor: replace `internal/agent/session_store.go` + `compaction.go`
- [ ] vaelor: replace `pkg/session/manager.go`
- [ ] go-hully: replace `internal/session/`
- [ ] picoclaw: replace `pkg/session/manager.go`

## Non-goals

- LLM provider integration (caller provides SummarizeFn)
- Vector/embedding storage (that's MemDB's job)
- Message routing or bus (that's the caller's domain)
