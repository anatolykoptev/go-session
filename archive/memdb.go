// Package archive provides archival backends for go-session cleanup.
package archive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	session "github.com/anatolykoptev/go-session"
)

const (
	defaultBaseURL   = "http://127.0.0.1:8080"
	defaultUserID    = "memos"
	defaultCubeID    = "memos"
	defaultKeepLastN = 5
	syncAddPath      = "/api/v1/sync/add"
)

// Config holds MemDB archival configuration.
type Config struct {
	// BaseURL is the MemDB API base URL. Defaults to http://127.0.0.1:8080.
	BaseURL string
	// UserID is the MemDB user ID. Defaults to "memos".
	UserID string
	// CubeID is the MemDB cube ID. Defaults to "memos".
	CubeID string
	// KeepLastN is the number of last messages to include in the archive. Defaults to 5.
	KeepLastN int
	// HTTPClient is an optional custom HTTP client.
	HTTPClient *http.Client
}

func (c *Config) applyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL
	}
	if c.UserID == "" {
		c.UserID = defaultUserID
	}
	if c.CubeID == "" {
		c.CubeID = defaultCubeID
	}
	if c.KeepLastN <= 0 {
		c.KeepLastN = defaultKeepLastN
	}
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
}

type syncAddRequest struct {
	UserID   string            `json:"user_id"`
	CubeID   string            `json:"cube_id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

// NewMemDBArchiver returns an archiveFn suitable for passing to session.Cleanup.
// It serialises each stale session as a text document and POSTs it to MemDB.
func NewMemDBArchiver(cfg Config) func(*session.Session) error {
	cfg.applyDefaults()

	return func(s *session.Session) error {
		content := buildContent(s, cfg.KeepLastN)

		body := syncAddRequest{
			UserID:  cfg.UserID,
			CubeID:  cfg.CubeID,
			Content: content,
			Metadata: map[string]string{
				"source":      "go-session",
				"session_key": s.Key,
				"archived_at": time.Now().UTC().Format(time.RFC3339),
			},
		}

		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("archive: marshal request: %w", err)
		}

		url := cfg.BaseURL + syncAddPath
		resp, err := cfg.HTTPClient.Post(url, "application/json", bytes.NewReader(data)) //nolint:noctx // archiveFn signature has no ctx
		if err != nil {
			return fmt.Errorf("archive: post to memdb: %w", err)
		}
		defer resp.Body.Close() //nolint:errcheck // best-effort close

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("archive: memdb returned %s", resp.Status)
		}
		return nil
	}
}

// buildContent formats session state into a human-readable document for MemDB.
func buildContent(s *session.Session, keepLastN int) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Session archive for key: %s\n", s.Key)

	if s.Summary != "" {
		fmt.Fprintf(&sb, "\nSummary: %s\n", s.Summary)
	}

	if len(s.Facts) > 0 {
		sb.WriteString("\nFacts:\n")
		for _, f := range s.Facts {
			fmt.Fprintf(&sb, "- %s\n", f.Content)
		}
	}

	msgs := s.Messages
	if len(msgs) > keepLastN {
		msgs = msgs[len(msgs)-keepLastN:]
	}

	if len(msgs) > 0 {
		fmt.Fprintf(&sb, "\nLast %d messages:\n", len(msgs))
		for _, m := range msgs {
			fmt.Fprintf(&sb, "%s: %s\n", m.Role, m.Content)
		}
	}

	return sb.String()
}
