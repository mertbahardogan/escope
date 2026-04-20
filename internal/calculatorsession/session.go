// Package calculatorsession stores calculator snapshots in the host config (sessions map).
package calculatorsession

import (
	"errors"
	"strconv"
	"strings"

	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/connection"
)

const fieldCount = 10

// State is calculator-only data (no host field; host is connection.CurrentHost()).
type State struct {
	Fields []string
	Focus  int
	Scroll int
}

// ReadState loads saved calculator state when YAML has data for the current host.
func ReadState() (*State, bool) {
	raw, ok := connection.SessionHostURL()
	if !ok {
		return nil, false
	}
	canon := config.CanonicalSessionHostKey(raw)
	hc, err := config.Load()
	if err != nil {
		return nil, false
	}
	merged := config.PickMergedHostSession(hc, canon, raw)
	if merged.Calculator == nil {
		return nil, false
	}
	c := merged.Calculator
	fields := migrateCalculatorFields(c.Fields)
	if fields == nil || len(fields) != fieldCount {
		return nil, false
	}
	focus := c.Focus
	if focus < 0 || focus >= fieldCount {
		focus = 0
	}
	scroll := c.Scroll
	if scroll < 0 {
		scroll = 0
	}
	return &State{Fields: fields, Focus: focus, Scroll: scroll}, true
}

func migrateCalculatorFields(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	if len(raw) == fieldCount {
		return normalizeFields(raw)
	}
	if len(raw) == 12 {
		out := []string{
			raw[0], raw[1], raw[2], raw[3], raw[4], raw[5], raw[6], raw[7],
			raw[10], raw[11],
		}
		return normalizeFields(out)
	}
	if len(raw) == 9 {
		total := atoiDef(raw[0], 3)
		masters := atoiDef(raw[1], 0)
		dataNodes := total - masters
		if dataNodes < 0 {
			dataNodes = 0
		}
		readRPS := atofDef(raw[6], 50)
		writeRPS := atofDef(raw[7], 50)
		if readRPS > 150 {
			readRPS /= 60.0
		}
		if writeRPS > 30 {
			writeRPS /= 60.0
		}
		out := []string{
			itoa(dataNodes),
			itoa(masters),
			raw[2],
			raw[3],
			raw[4],
			raw[5],
			ftoa(readRPS),
			ftoa(writeRPS),
			"64",
			"2000",
		}
		return normalizeFields(out)
	}
	return normalizeFields(raw)
}

func atoiDef(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func atofDef(s string, def float64) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return n
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func ftoa(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func normalizeFields(s []string) []string {
	out := make([]string, fieldCount)
	for i := range out {
		if i < len(s) {
			out[i] = strings.TrimSpace(s[i])
		}
	}
	return out
}

// Write merges calculator state for the current host into the config file.
// It preserves default_index on the same host entry and consolidates duplicate session keys (e.g. with/without trailing slash).
func Write(st *State) error {
	if st == nil {
		return errors.New("nil state")
	}
	raw, ok := connection.SessionHostURL()
	if !ok {
		return errors.New("no Elasticsearch host configured")
	}
	canon := config.CanonicalSessionHostKey(raw)
	fields := migrateCalculatorFields(st.Fields)
	if fields == nil || len(fields) != fieldCount {
		return errors.New("invalid field count")
	}
	focus := st.Focus
	if focus < 0 || focus >= fieldCount {
		focus = 0
	}
	scroll := st.Scroll
	if scroll < 0 {
		scroll = 0
	}
	hc, err := config.Load()
	if err != nil {
		return err
	}
	merged := config.PickMergedHostSession(hc, canon, raw)
	merged.Calculator = &config.CalculatorSession{
		Fields: fields,
		Focus:  focus,
		Scroll: scroll,
	}
	if hc.Sessions == nil {
		hc.Sessions = make(map[string]config.HostSessionData)
	}
	hc.Sessions[canon] = merged
	config.PruneDuplicateSessionKeys(&hc, canon, raw)
	return config.Save(hc)
}

// Clear removes calculator session data for the current host (keeps default index if set).
func Clear() error {
	raw, ok := connection.SessionHostURL()
	if !ok {
		return errors.New("no Elasticsearch host configured")
	}
	return config.ClearCalculator(raw)
}
