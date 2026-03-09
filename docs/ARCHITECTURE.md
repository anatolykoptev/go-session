# go-session Architecture

## Purpose

Shared Go library for per-key conversation session management across AI agent projects.
Replaces duplicated session stores in vaelor, dozor, go-hully, and picoclaw.

## Design Principles

1. **Interface-first** — pluggable backends (memory, file, redis)
2. **Zero external dependencies** for core — only stdlib
3. **Optional integrations** — redis, MemDB archival via sub-packages
4. **Generic Message type** — not tied to any specific LLM provider
5. **Thread-safe** — all operations safe for concurrent use

## Architecture

```
go-session/
├── message.go          # Message type (Role, Content, ToolCallID, ToolCalls)
├── session.go          # Session struct (Key, Messages, Summary, Facts, timestamps)
├── store.go            # Store interface + errors
├── memory.go           # InMemoryStore (map + RWMutex, TTL eviction)
├── file.go             # FileStore (JSON files, atomic writes)
├── compact.go          # Compactor interface + multi-part summarization
├── intent.go           # NeedsMemoryContext (force/skip keywords filtering)
├── cleanup.go          # StaleCleanup (TTL-based, archive callback)
├── redis/
│   └── store.go        # RedisStore (go-redis, TTL, JSON serialization)
└── archive/
    └── memdb.go        # MemDB archival bridge (facts + last N messages)
```

## Core Types

```go
// Message is a provider-agnostic chat message.
type Message struct {
    Role       string     `json:"role"`
    Content    string     `json:"content"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a single tool invocation.
type ToolCall struct {
    ID       string        `json:"id"`
    Name     string        `json:"name,omitempty"`
    Args     string        `json:"arguments,omitempty"`
    Function *FunctionCall `json:"function,omitempty"`
}

// FunctionCall is the OpenAI-style nested function call format.
type FunctionCall struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}

// Fact is a single extracted fact from compacted conversation.
type Fact struct {
    Content     string    `json:"content"`
    ExtractedAt time.Time `json:"extracted_at"`
}

// Session holds conversation state for a single key.
type Session struct {
    Key      string    `json:"key"`
    Messages []Message `json:"messages"`
    Summary  string    `json:"summary,omitempty"`
    Facts    []Fact    `json:"facts,omitempty"`
    Created  time.Time `json:"created"`
    Updated  time.Time `json:"updated"`
}
```

## Store Interface

```go
type Store interface {
    // GetOrCreate returns existing session or creates a new one.
    GetOrCreate(key string) *Session

    // AddMessage appends a message to session history.
    AddMessage(key string, msg Message)

    // GetHistory returns ordered copy of session messages.
    GetHistory(key string) []Message

    // GetSummary returns the compaction summary.
    GetSummary(key string) string

    // SetSummary stores a compaction summary.
    SetSummary(key, summary string)

    // GetFacts returns extracted facts.
    GetFacts(key string) []Fact

    // AddFacts appends new facts.
    AddFacts(key string, facts []Fact)

    // MessageCount returns number of messages.
    MessageCount(key string) int

    // CompactMessages extracts oldest messages, keeping keepLast.
    CompactMessages(key string, keepLast int) []Message

    // TruncateHistory removes oldest, keeping keepLast.
    TruncateHistory(key string, keepLast int)

    // Clear resets a session.
    Clear(key string)

    // Delete removes a session entirely.
    Delete(key string) error

    // Save persists a session.
    Save(key string) error

    // ListStale returns keys where Updated is older than maxAge.
    ListStale(maxAge time.Duration) []string
}
```

## Compactor Interface

```go
// SummarizeFn is called by the compactor to get an LLM summary.
// The library does NOT import any LLM provider — the caller provides this.
type SummarizeFn func(ctx context.Context, prompt string) (string, error)

// Compactor manages session compaction with configurable strategy.
type Compactor struct {
    Store        Store
    Summarize    SummarizeFn
    Threshold    int  // trigger compaction when MessageCount >= this
    KeepLast     int  // messages to retain after compaction
    ExtractFacts bool // true = bullet-point facts, false = prose summary
    MultiPart    bool // split large histories into parts before summarizing
}
```

## Data Flow

```
User message → Store.AddMessage()
                  ↓
            MessageCount >= Threshold?
                  ↓ yes
            Compactor.Compact(ctx, key)
                  ↓
            CompactMessages(key, keepLast) → removed messages
                  ↓
            SummarizeFn(prompt) → summary/facts
                  ↓
            SetSummary() / AddFacts()
                  ↓
            Save()

Session expires → ListStale(maxAge) → archiveFn(session) → Delete()
```

## Backend Details

### InMemoryStore
- `map[string]*Session` + `sync.RWMutex`
- Optional TTL: entries expire after configurable duration
- Optional maxMessages: auto-truncate on AddMessage

### FileStore
- Wraps InMemoryStore + JSON persistence
- Atomic writes (tmp file + rename)
- Auto-load on startup from directory
- File naming: `<sanitized-key>.json`

### RedisStore (sub-package `redis/`)
- JSON serialization to Redis strings
- TTL via Redis EXPIRE
- Configurable key prefix
- Requires `github.com/redis/go-redis/v9`

## Consumers

| Project | Current | Migration to go-session |
|---------|---------|------------------------|
| dozor | `internal/agent/session_store.go` | FileStore + Compactor |
| vaelor | `pkg/session/manager.go` | FileStore + Compactor(facts) |
| go-hully | `internal/session/{store,redis}.go` | InMemoryStore or RedisStore |
| picoclaw | `pkg/session/manager.go` | FileStore + Compactor(multipart) |
