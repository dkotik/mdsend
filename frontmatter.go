package mdsend

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	FieldNameID                        = "id"
	FieldNameSeed                      = "seed"
	FieldNameExtends                   = "extends"
	FieldNameDatabase                  = "queue"
	FieldNameSubject                   = "subject"
	FieldNameFrom                      = "from"
	FieldNameReplyTo                   = "reply_to"
	FieldNameAttachments               = "attachments"
	FieldNameAttachmentName            = "name"
	FieldNameAttachmentLocation        = "location"
	FieldNameTemplates                 = "templates"
	FieldNameHeaders                   = "headers"
	FieldNameLanguage                  = "language"
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
