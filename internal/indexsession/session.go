package indexsession

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/connection"
)

func ReadSelectedIndex() (string, bool) {
	raw, ok := connection.SessionHostURL()
	if !ok {
		return "", false
	}
	canon := config.CanonicalSessionHostKey(raw)
	hc, err := config.Load()
	if err != nil {
		return "", false
	}
	merged := config.PickMergedHostSession(hc, canon, raw)
	idx := strings.TrimSpace(merged.DefaultIndex)
	if idx == "" {
		return "", false
	}
	return idx, true
}

func WriteSelectedIndex(index string) error {
	index = strings.TrimSpace(index)
	if index == "" {
		return errors.New("index name is empty")
	}
	raw, ok := connection.SessionHostURL()
	if !ok {
		return errors.New("no Elasticsearch host configured")
	}
	canon := config.CanonicalSessionHostKey(raw)
	hc, err := config.Load()
	if err != nil {
		return err
	}
	merged := config.PickMergedHostSession(hc, canon, raw)
	merged.DefaultIndex = index
	if hc.Sessions == nil {
		hc.Sessions = make(map[string]config.HostSessionData)
	}
	hc.Sessions[canon] = merged
	config.PruneDuplicateSessionKeys(&hc, canon, raw)
	return config.Save(hc)
}

func Clear() error {
	raw, ok := connection.SessionHostURL()
	if !ok {
		return nil
	}
	return config.ClearDefaultIndex(raw)
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
