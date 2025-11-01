package ui

import (
	"fmt"
	"github.com/mertbahardogan/escope/internal/constants"
)

type IndexDetailFormatter struct{}

func NewIndexDetailFormatter() *IndexDetailFormatter {
	return &IndexDetailFormatter{}
}

func (f *IndexDetailFormatter) FormatIndexDetail(info *IndexDetailInfo) string {
	var output string
	output += "\n"
	output += fmt.Sprintf("Search Rate: %s%s\n", info.SearchRate, constants.ANSIClearLineEnd)
	output += fmt.Sprintf("Index Rate: %s%s\n", info.IndexRate, constants.ANSIClearLineEnd)
	output += fmt.Sprintf("Query Time: %s%s\n", info.AvgQueryTime, constants.ANSIClearLineEnd)
	output += fmt.Sprintf("Index Time: %s%s\n", info.AvgIndexTime, constants.ANSIClearLineEnd)
	return output
}

type IndexDetailInfo struct {
	Name         string
	SearchRate   string
	IndexRate    string
	AvgQueryTime string
	AvgIndexTime string
	CheckCount   int
}
