package main

import (
	_ "github.com/mertbahardogan/escope/cmd/analyze"
	_ "github.com/mertbahardogan/escope/cmd/check"
	_ "github.com/mertbahardogan/escope/cmd/cluster"
	_ "github.com/mertbahardogan/escope/cmd/config"
	"github.com/mertbahardogan/escope/cmd/core"
	_ "github.com/mertbahardogan/escope/cmd/index"
	_ "github.com/mertbahardogan/escope/cmd/lucene"
	_ "github.com/mertbahardogan/escope/cmd/node"
	_ "github.com/mertbahardogan/escope/cmd/segments"
	_ "github.com/mertbahardogan/escope/cmd/shard"
	_ "github.com/mertbahardogan/escope/cmd/sort"
	_ "github.com/mertbahardogan/escope/cmd/system"
	_ "github.com/mertbahardogan/escope/cmd/termvectors"
	_ "github.com/mertbahardogan/escope/cmd/upgrade"
)

func main() {
	core.Execute()
}
