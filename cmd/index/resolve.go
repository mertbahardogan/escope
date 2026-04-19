package index

import (
	"fmt"
	"strings"

	"github.com/mertbahardogan/escope/internal/indexsession"
)

func resolveIndexName(flagValue string) string {
	if s := strings.TrimSpace(flagValue); s != "" {
		return s
	}
	if idx, ok := indexsession.ReadSelectedIndex(); ok {
		return idx
	}
	return ""
}

func printIndexNameRequired() {
	fmt.Println("Error: no index specified.")
	fmt.Println("Use --name <index-or-alias>, or select once with: escope index use <index-or-alias>")
}
