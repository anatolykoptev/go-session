package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// FileStore wraps InMemoryStore with JSON file persistence.
type FileStore struct {
	*InMemoryStore
	dir string
}

// NewFileStore creates a file-backed store, loading existing sessions.
func NewFileStore(dir string, opts Options) Store {
	if err := os.MkdirAll(dir, 0750); err != nil {
		panic("go-session: cannot create dir: " + err.Error())
	}

	fs := &FileStore{
		InMemoryStore: NewInMemoryStore(opts),
		dir:           dir,
	}
	fs.loadAll()
	return fs
}

func (fs *FileStore) loadAll() {
	entries, err := os.ReadDir(fs.dir)
	if err != nil {
		return
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(fs.dir, e.Name()))
		if err != nil {
			continue
		}

		var sess Session
		if err := json.Unmarshal(data, &sess); err != nil {
			continue // skip corrupt
		}

		fs.mu.Lock()
		fs.sessions[sess.Key] = &sess
		fs.mu.Unlock()
	}
}

func sanitizeKey(key string) string {
	r := strings.NewReplacer("/", "_", ":", "_", " ", "_")
	return r.Replace(key)
}

func (fs *FileStore) filename(key string) string {
	return filepath.Join(fs.dir, sanitizeKey(key)+".json")
}

// Save persists a session to disk atomically.
func (fs *FileStore) Save(key string) error {
	fs.mu.RLock()
	sess, ok := fs.sessions[key]
	if !ok {
		fs.mu.RUnlock()
		return nil
	}

	data, err := json.Marshal(sess)
	fs.mu.RUnlock()
	if err != nil {
		return err
	}

	tmp := fs.filename(key) + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, fs.filename(key))
}

// Delete removes from memory and disk.
func (fs *FileStore) Delete(key string) error {
	if err := fs.InMemoryStore.Delete(key); err != nil {
		return err
	}
	path := fs.filename(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
