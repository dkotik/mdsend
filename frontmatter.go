package mdsend

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	FieldNameExtends                   = "extends"
	FieldNameQueue                     = "queue"
	FieldNameSubject                   = "subject"
	FieldNameFrom                      = "from"
	FieldNameReplyTo                   = "reply_to"
	FieldNameTo                        = "to"
	FieldNameCarbonCopy                = "cc"
	FieldNameBlindCarbonCopy           = "bcc"
	FieldNameName                      = "name"
	FieldNameEmail                     = "email"
	FieldNameAttachments               = "attachments"
	FieldNameSendAfter                 = "send_after"
	FieldNameMediaContraints           = "media_constraints"
	FieldNameMediaConstraintsQuality   = "quality"
	FieldNameMediaConstrainsResolution = "resolution"
	FieldNameMediaConstraintsWidth     = "width"
	FieldNameMediaConstrainsHeight     = "height"
	FieldNameSchedule                  = "schedule"
	FieldNameScheduleAfter             = "after"
	FieldNameScheduleDelay             = "delay"
	FieldNameScheduleStep              = "step"
	FieldNameScheduleExpire            = "expire"
	FieldNameScheduleFluctuate         = "fluctuate"
	FieldNameListID                    = "list_id"
	FieldNameUnsubscribe               = "unsubscribe"
	FieldNameUnsubscribeEmail          = "unsubscribe_email"
	FieldNameUnsubscribeURL            = "unsubscribe_url"
)

func mergeMaps(ms ...map[string]any) (result map[string]any) {
	switch len(ms) {
	case 0:
		return make(map[string]any)
	case 1:
		return ms[0]
	}
	result = make(map[string]any)

	// copy the first map
	for k, v := range ms[0] {
		result[strings.ToLower(k)] = v
	}

	// override with later maps
	var (
		existing any
		ok       bool
	)
	for _, m := range ms[1:] {
		for k, v := range m {
			k = strings.ToLower(k)
			existing, ok = result[k]
			if !ok { // simplest
				result[k] = v
				continue
			}

			switch existing := existing.(type) {
			case []any:
				switch v := v.(type) {
				case nil:
					continue // skip nil values
				case []any:
					result[k] = append(existing, v...)
				default:
					result[k] = append(existing, v)
				}
			case map[string]any:
				switch v := v.(type) {
				case nil:
					continue // skip nil values
				case map[string]any:
					result[k] = mergeMaps(existing, v)
				default:
					result[k] = v
				}
			default:
				if v == nil {
					continue // skip nil values
				}
				result[k] = v
			}
		}
	}
	return result
}

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
