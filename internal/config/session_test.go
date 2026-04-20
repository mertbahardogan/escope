package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mertbahardogan/escope/internal/constants"
)

func TestPutHostSessionRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	host := "http://localhost:9200"
	data := HostSessionData{
		DefaultIndex: "my-index",
		Calculator: &CalculatorSession{
			Fields: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
			Focus:  2,
			Scroll: 1,
		},
	}
	if err := PutHostSession(host, data); err != nil {
		t.Fatal(err)
	}
	got, ok := GetHostSession(host)
	if !ok {
		t.Fatal("expected session")
	}
	if got.DefaultIndex != "my-index" {
		t.Fatal(got.DefaultIndex)
	}
	if got.Calculator == nil || len(got.Calculator.Fields) != 9 {
		t.Fatal("calculator")
	}
	if err := ClearCalculator(host); err != nil {
		t.Fatal(err)
	}
	got, ok = GetHostSession(host)
	if !ok {
		t.Fatal("need default index still")
	}
	if got.DefaultIndex != "my-index" || got.Calculator != nil {
		t.Fatalf("%+v", got)
	}
	if err := ClearDefaultIndex(host); err != nil {
		t.Fatal(err)
	}
	if _, ok := GetHostSession(host); ok {
		t.Fatal("expected no session")
	}

	// file exists
	path := filepath.Join(tmp, constants.ConfigFilePath)
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
