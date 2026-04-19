package indexsession

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mertbahardogan/escope/internal/connection"
)

type payload struct {
	Host  string `json:"host"`
	Index string `json:"index"`
}

func sessionFilePath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "escope")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "index-session.json"), nil
}

// ReadSelectedIndex returns the index or alias saved for the current connection host.
func ReadSelectedIndex() (string, bool) {
	path, err := sessionFilePath()
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var p payload
	if err := json.Unmarshal(data, &p); err != nil {
		return "", false
	}
	host := connection.CurrentHost()
	if host == "" || p.Host != host || strings.TrimSpace(p.Index) == "" {
		return "", false
	}
	return strings.TrimSpace(p.Index), true
}

// WriteSelectedIndex stores the index or alias for the current connection host.
func WriteSelectedIndex(index string) error {
	index = strings.TrimSpace(index)
	if index == "" {
		return errors.New("index name is empty")
	}
	host := connection.CurrentHost()
	if host == "" {
		return errors.New("no Elasticsearch host configured")
	}
	path, err := sessionFilePath()
	if err != nil {
		return err
	}
	p := payload{Host: host, Index: index}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Clear removes stored index selection.
func Clear() error {
	path, err := sessionFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DescribeCurrent prints human-readable state for the use command (no trailing newline).
func DescribeCurrent() (string, error) {
	host := connection.CurrentHost()
	if host == "" {
		return "", fmt.Errorf("no Elasticsearch host configured")
	}
	if idx, ok := ReadSelectedIndex(); ok {
		return fmt.Sprintf("Selected index for %s: %s", host, idx), nil
	}
	return fmt.Sprintf("No index selected for %s (use: escope index use <index-or-alias>)", host), nil
}
