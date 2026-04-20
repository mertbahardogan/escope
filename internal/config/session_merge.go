package config

import "strings"

// CanonicalSessionHostKey normalizes the URL used as the primary sessions map key.
func CanonicalSessionHostKey(h string) string {
	h = strings.TrimSpace(h)
	for strings.HasSuffix(h, "/") {
		h = strings.TrimSuffix(h, "/")
	}
	return h
}

func candidateSessionKeys(canon, raw string) []string {
	seen := make(map[string]bool)
	var out []string
	add := func(k string) {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			return
		}
		seen[k] = true
		out = append(out, k)
	}
	add(canon)
	add(raw)
	if raw != "" && !strings.HasSuffix(raw, "/") {
		add(raw + "/")
	}
	if canon != "" && !strings.HasSuffix(canon, "/") {
		add(canon + "/")
	}
	return out
}

// PickMergedHostSession merges default_index and calculator from all session entries
// that belong to this connection (exact keys, slash variants, and same canonical URL).
func PickMergedHostSession(hc HostConfig, canon, raw string) HostSessionData {
	var merged HostSessionData
	if hc.Sessions == nil {
		return merged
	}
	seen := make(map[string]bool)
	try := func(d HostSessionData) {
		if merged.DefaultIndex == "" && d.DefaultIndex != "" {
			merged.DefaultIndex = d.DefaultIndex
		}
		if merged.Calculator == nil && d.Calculator != nil {
			merged.Calculator = d.Calculator
		}
	}
	for _, k := range candidateSessionKeys(canon, raw) {
		if d, ok := hc.Sessions[k]; ok {
			try(d)
			seen[k] = true
		}
	}
	for k, d := range hc.Sessions {
		if seen[k] {
			continue
		}
		if CanonicalSessionHostKey(k) == canon {
			try(d)
		}
	}
	return merged
}

// PruneDuplicateSessionKeys keeps data under canon and removes alias keys for the same host.
func PruneDuplicateSessionKeys(hc *HostConfig, canon, raw string) {
	if hc.Sessions == nil {
		return
	}
	for k := range hc.Sessions {
		if k != canon && CanonicalSessionHostKey(k) == canon {
			delete(hc.Sessions, k)
		}
	}
	for _, k := range candidateSessionKeys(canon, raw) {
		if k != canon {
			delete(hc.Sessions, k)
		}
	}
}
