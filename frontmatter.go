package mdsend

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	FieldNameSubject         = "subject"
	FieldNameFrom            = "from"
	FieldNameName            = "name"
	FieldNameEmail           = "email"
	FieldNameAttachments     = "attachments"
	FieldNameSendAfter       = "send_after"
	FieldNameMediaContraints = "media_constraints"
	FieldNameMediaQuality    = "quality"
	FieldNameMediaResolution = "resolution"
	FieldNameMediaWidth      = "width"
	FieldNameMediaHeight     = "height"
	// FieldNameExpireAfter     = "expiration"
)

func getIntFromMap(m map[string]interface{}, key string, defaultValue int) (int, error) {
	switch v := m[key].(type) {
	case nil:
		return defaultValue, nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case int16:
		return int(v), nil
	// case int8:
	// 	return int(v), nil
	// case uint8:
	// 	return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("invalid type for %s: %T", key, v)
	}
}

var rePercent = regexp.MustCompile(`^\s*(\d+)\s*\%\s*$`)

func getPercentageFromMap(m map[string]interface{}, key string, defaultValue int) (int, error) {
	switch v := m[key].(type) {
	case nil:
		return defaultValue, nil
	case string:
		m := rePercent.FindStringSubmatch(v)
		if m == nil {
			return defaultValue, fmt.Errorf("invalid percentage for %s: %s", key, v)
		}
		percent, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, fmt.Errorf("invalid percentage for %s: %s", key, v)
		}
		if percent > 100 {
			return 0, fmt.Errorf("invalid percentage for %s: %s", key, v)
		}
		return percent, nil
	default:
		return 0, fmt.Errorf("invalid type for %s: %T", key, v)
	}
}
