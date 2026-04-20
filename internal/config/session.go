package config

import (
	"errors"
)

type CalculatorSession struct {
	Fields []string `yaml:"fields"`
	Focus  int      `yaml:"focus"`
	Scroll int      `yaml:"scroll"`
}

type HostSessionData struct {
	DefaultIndex string             `yaml:"default_index,omitempty"`
	Calculator   *CalculatorSession `yaml:"calculator,omitempty"`
}

func GetHostSession(host string) (HostSessionData, bool) {
	if host == "" {
		return HostSessionData{}, false
	}
	hc, err := Load()
	if err != nil {
		return HostSessionData{}, false
	}
	if hc.Sessions == nil {
		return HostSessionData{}, false
	}
	v, ok := hc.Sessions[host]
	return v, ok
}

// PutHostSession replaces session data for a host. Empty DefaultIndex and nil Calculator removes the host key.
func PutHostSession(host string, data HostSessionData) error {
	if host == "" {
		return errors.New("empty host")
	}
	hc, err := Load()
	if err != nil {
		return err
	}
	if hc.Sessions == nil {
		hc.Sessions = make(map[string]HostSessionData)
	}
	if data.DefaultIndex == "" && data.Calculator == nil {
		delete(hc.Sessions, host)
	} else {
		hc.Sessions[host] = data
	}
	return Save(hc)
}

// ClearDefaultIndex removes the default index for a host, keeping calculator session if any.
func ClearDefaultIndex(host string) error {
	if host == "" {
		return nil
	}
	canon := CanonicalSessionHostKey(host)
	hc, err := Load()
	if err != nil {
		return err
	}
	if hc.Sessions == nil {
		return nil
	}
	merged := PickMergedHostSession(hc, canon, host)
	merged.DefaultIndex = ""
	if merged.Calculator == nil {
		delete(hc.Sessions, canon)
	} else {
		hc.Sessions[canon] = merged
	}
	PruneDuplicateSessionKeys(&hc, canon, host)
	return Save(hc)
}

// ClearCalculator removes calculator data for a host, keeping default index if set.
func ClearCalculator(host string) error {
	if host == "" {
		return nil
	}
	canon := CanonicalSessionHostKey(host)
	hc, err := Load()
	if err != nil {
		return err
	}
	if hc.Sessions == nil {
		return nil
	}
	merged := PickMergedHostSession(hc, canon, host)
	merged.Calculator = nil
	if merged.DefaultIndex == "" {
		delete(hc.Sessions, canon)
	} else {
		hc.Sessions[canon] = merged
	}
	PruneDuplicateSessionKeys(&hc, canon, host)
	return Save(hc)
}
