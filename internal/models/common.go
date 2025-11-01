package models

import (
	"github.com/mertbahardogan/escope/internal/constants"
	"strconv"
	"strings"
)

func ParseSize(sizeStr string) int64 {
	if sizeStr == "" || sizeStr == constants.DashString {
		return 0
	}

	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))

	if strings.HasSuffix(sizeStr, "b") {
		sizeStr = strings.TrimSuffix(sizeStr, "b")
	}

	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "kb") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "kb")
	} else if strings.HasSuffix(sizeStr, "mb") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "mb")
	} else if strings.HasSuffix(sizeStr, "gb") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "gb")
	} else if strings.HasSuffix(sizeStr, "tb") {
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "tb")
	}

	value, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0
	}

	return int64(value * float64(multiplier))
}

func FormatBytes(bytes int64) string {
	if bytes == 0 {
		return "0b"
	}

	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + "b"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + string("kmgt"[exp]) + "b"
}
