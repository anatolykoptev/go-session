# go-session Quality Standards

## Coverage Target: 90%+

This is a shared library — bugs here affect 4+ projects. High coverage is mandatory.

## Required Test Suites

### message_test.go
- [ ] Message JSON round-trip (all fields)
- [ ] Message JSON omitempty (empty ToolCalls not serialized)
- [ ] ToolCall with nested FunctionCall serialization
- [ ] Empty message serialization

### session_test.go
- [ ] NewSession sets timestamps
- [ ] AddMessage appends and updates timestamp
- [ ] MessageCount returns correct count
- [ ] CompactMessages extracts oldest, keeps last N
- [ ] CompactMessages with fewer than keepLast returns nil
- [ ] TruncateHistory removes oldest
- [ ] Clear resets messages, summary, facts
- [ ] GetFacts returns copy (not reference)
- [ ] AddFacts appends to existing

### memory_test.go (InMemoryStore)
- [ ] GetOrCreate creates new session
- [ ] GetOrCreate returns existing session
- [ ] AddMessage to new key auto-creates session
- [ ] GetHistory returns ordered copy
- [ ] GetHistory for unknown key returns nil
- [ ] GetSummary / SetSummary round-trip
- [ ] GetFacts / AddFacts round-trip
- [ ] MessageCount for existing and missing keys
- [ ] CompactMessages extracts and truncates
- [ ] TruncateHistory keeps last N
- [ ] Clear resets session
- [ ] Delete removes session entirely
- [ ] Save is no-op (returns nil)
- [ ] ListStale returns expired sessions
- [ ] ListStale ignores fresh sessions
- [ ] TTL: expired entries return nil on GetHistory
- [ ] MaxMessages: auto-truncates on AddMessage
- [ ] Concurrent AddMessage safety (goroutine test)
- [ ] Concurrent GetHistory during writes

### file_test.go (FileStore)
- [ ] Save creates JSON file
- [ ] Save uses atomic write (tmp + rename)
- [ ] Load restores sessions from directory
- [ ] Load skips corrupt JSON files
- [ ] Load skips non-JSON files
- [ ] Delete removes file from disk
- [ ] File naming sanitizes special characters
- [ ] Empty directory handled gracefully
- [ ] Concurrent Save from multiple goroutines

### compact_test.go (Compactor)
- [ ] Compact triggers when MessageCount >= Threshold
- [ ] Compact is no-op under threshold
- [ ] Compact removes old messages, keeps last N
- [ ] Summary mode: SummarizeFn called with conversation text
- [ ] Fact extraction mode: facts parsed from bullet list
- [ ] Fact extraction handles malformed LLM output
- [ ] Multi-part: large history split, summarized separately, merged
- [ ] Multi-part threshold (>10 messages)
- [ ] Oversized message guard: skips messages > maxTokens
- [ ] SummarizeFn error: compaction aborted, messages restored
- [ ] Existing summary included in prompt for incremental compaction
- [ ] Existing facts preserved across compactions

### intent_test.go
- [ ] Short messages (<5 runes) filtered out
- [ ] Greetings (RU/EN) filtered out
- [ ] Thanks filtered out
- [ ] Substantive questions pass through
- [ ] Force keywords ("помнишь", "remember") always pass
- [ ] Mixed: greeting + substantive content passes (long enough)
- [ ] Empty string filtered out
- [ ] Unicode handling (emoji, CJK)

### cleanup_test.go
- [ ] ListStale finds expired sessions
- [ ] Cleanup calls archive callback before deletion
- [ ] Cleanup deletes stale sessions
- [ ] Cleanup preserves fresh sessions
- [ ] Archive callback error doesn't prevent deletion
- [ ] Zero retention disables cleanup

## Lint / Vet / Build

- `golangci-lint run` clean (v2 config)
- `go vet ./...` clean
- `go build ./...` clean
- No external dependencies in core package (only stdlib)
- `redis/` sub-package: only `github.com/redis/go-redis/v9`

## Benchmarks (optional, Phase 2)

- BenchmarkAddMessage (target: <100ns)
- BenchmarkGetHistory (target: <500ns for 100 messages)
- BenchmarkFileStoreSave (target: <1ms)
- BenchmarkCompactMessages (target: <1μs for 50 messages)
