# go-session Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Shared Go library for per-key conversation session management, replacing duplicated stores in dozor, vaelor, go-hully, and picoclaw.

**Architecture:** Interface-first design with pluggable backends (InMemory, File, Redis). Core has zero external deps. Compactor uses caller-provided SummarizeFn (no LLM coupling). Multi-part summarization for large histories, fact extraction, oversized message guard.

**Tech Stack:** Go 1.22+, stdlib only (redis sub-package uses go-redis/v9)

---

### Task 1: Session struct methods

**Files:**
- Modify: `session.go`
- Test: `session_test.go` (already written)

Implement `NewSession`, `AddMessage`, `MessageCount`, `CompactMessages`, `TruncateHistory`, `Clear`, `GetFacts`, `AddFacts` methods on `Session`.

### Task 2: InMemoryStore

**Files:**
- Create: `memory.go`
- Test: `memory_test.go` (already written)

Implement `InMemoryStore` with `Options{TTL, MaxMessages}`. All Store interface methods. Thread-safe via `sync.RWMutex`. TTL check on reads, maxMessages cap on writes.

### Task 3: FileStore

**Files:**
- Create: `file.go`
- Test: `file_test.go` (already written)

Wraps InMemoryStore + JSON persistence. Atomic writes (tmp+rename). Auto-load from dir on startup. Key sanitization for filenames. Delete removes file from disk.

### Task 4: Compactor

**Files:**
- Create: `compact.go`
- Test: `compact_test.go` (already written)

Implement `Compactor` struct with `Compact(ctx, key)`. Summary mode (prose) and fact extraction mode (bullet list parsing). Multi-part: split at >10 messages, summarize halves, merge. Oversized message guard (MaxTokensGuard). Existing summary/facts preserved.

### Task 5: Intent filter

**Files:**
- Create: `intent.go`
- Test: `intent_test.go` (already written)

`NeedsMemoryContext(text)` — merged logic from vaelor (force/skip keywords) and dozor (prefix + rune-length filtering).

### Task 6: Cleanup

**Files:**
- Create: `cleanup.go`
- Test: `cleanup_test.go` (already written)

`Cleanup(store, maxAge, archiveFn)` — finds stale sessions, calls archive callback, deletes. Zero maxAge disables.

### Task 7: Lint, vet, final verification

Run `go vet ./...`, `go build ./...`, verify all tests pass, check coverage.
