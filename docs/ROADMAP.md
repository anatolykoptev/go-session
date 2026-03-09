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

- [x] `redis/` sub-package with RedisStore
- [x] Integration tests (skip if no Redis)

## Phase 3: MemDB archival bridge

- [x] `archive/memdb.go` — archive session to MemDB on cleanup
- [x] Configurable archive format (facts + last N messages)
- [x] Tests with httptest mock

## Phase 4: Consumer migration

- [x] go-hully: replaced `internal/session/` with go-session adapter
- [x] dozor: replaced `internal/agent/session_store.go` + `compaction.go`
- [x] vaelor: replaced `pkg/session/manager.go` + intent filter + cleanup
- [ ] picoclaw: replace `pkg/session/manager.go`

## Non-goals

- LLM provider integration (caller provides SummarizeFn)
- Vector/embedding storage (that's MemDB's job)
- Message routing or bus (that's the caller's domain)
