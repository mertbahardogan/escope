package util

import (
	"fmt"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/models"
	"strconv"
	"strings"
)

func FormatBytes(bytes int64) string {
	if bytes == 0 {
		return constants.ZeroByteString
	}

	if bytes < constants.BytesInKB {
		return fmt.Sprintf("%d%s", bytes, constants.ByteSuffix)
	}

	var div int64
	var suffix string

	if bytes >= constants.BytesInTB {
		div = constants.BytesInTB
		suffix = constants.TeraSuffix
	} else if bytes >= constants.BytesInGB {
		div = constants.BytesInGB
		suffix = constants.GigaSuffix
	} else if bytes >= constants.BytesInMB {
		div = constants.BytesInMB
		suffix = constants.MegaSuffix
	} else {
		div = constants.BytesInKB
		suffix = constants.KiloSuffix
	}

	size := float64(bytes) / float64(div)

	if size < constants.TenThreshold {
		return fmt.Sprintf("%.1f%s", size, suffix)
	}
	return fmt.Sprintf("%.0f%s", size, suffix)
}

func IsSystemIndex(name string) bool {
	return strings.HasPrefix(name, constants.DotPrefix) ||
		strings.HasPrefix(name, constants.KibanaPrefix) ||
		strings.HasPrefix(name, constants.APMPrefix) ||
		strings.HasPrefix(name, constants.SecurityPrefix) ||
		strings.HasPrefix(name, constants.MonitoringPrefix) ||
		strings.HasPrefix(name, constants.WatcherPrefix) ||
		strings.HasPrefix(name, constants.ILMPrefix) ||
		strings.HasPrefix(name, constants.SLMPrefix) ||
		strings.HasPrefix(name, constants.TransformPrefix)
}

func ConvertShardName(prirep string) string {
	if prirep == constants.PrimaryShortString {
		return constants.PrimaryString
	}
	return constants.ReplicaString
}

func FormatDocsCount(count int64) string {
	if count == 0 {
		return constants.DashString
	}
	str := fmt.Sprintf("%d", count)
	for i := len(str) - constants.DocsCountSeparator; i > 0; i -= constants.DocsCountSeparator {
		str = str[:i] + "." + str[i:]
	}

	return str
}

func GetStringField(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func GetIntField(data map[string]interface{}, key string) int {
	if value, ok := data[key]; ok {
		if num, ok := value.(float64); ok {
			return int(num)
		}
		if num, ok := value.(int); ok {
			return num
		}
		if str, ok := value.(string); ok {
			if num, err := strconv.Atoi(str); err == nil {
				return num
			}
		}
	}
	return 0
}

func SortShardsByTypeAndIndex(shards []models.ShardInfo) {
	if len(shards) == 0 {
		return
	}

	for i := 0; i < len(shards)-1; i++ {
		for j := 0; j < len(shards)-i-1; j++ {
			shouldSwap := false

			if shards[j].Prirep != shards[j+1].Prirep {
				if shards[j].Prirep == constants.ReplicaShortString && shards[j+1].Prirep == constants.PrimaryShortString {
					shouldSwap = true
				}
			} else {
				if shards[j].Index > shards[j+1].Index {
					shouldSwap = true
				}
			}

			if shouldSwap {
				shards[j], shards[j+1] = shards[j+1], shards[j]
			}
		}
	}
}

func CalculatePercentage(used, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) * constants.HundredMultiplier / float64(total)
}

func ParsePercentString(percentStr string) (float64, error) {
	cleanStr := strings.TrimSuffix(strings.TrimSpace(percentStr), "%")
	return strconv.ParseFloat(cleanStr, 64)
}
