package loader

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dkotik/mdsend"
)

func (l loader) loadExtensions(letter mdsend.Letter, rootDirectory string) (_ mdsend.Letter, err error) {
	ext, ok := letter.Frontmatter[mdsend.FieldNameExtends]
	if !ok {
		return letter, nil
	}
	var extends []string
	switch ext := ext.(type) {
	case string:
		extends = append(extends, ext)
	case []any:
		for _, e := range ext {
			if s, ok := e.(string); ok {
				extends = append(extends, s)
			} else {
				return letter, fmt.Errorf("unsupported extension file type: %v(%T)", e, e)
			}
		}
	default:
		return letter, fmt.Errorf("unsupported extension file type: %v(%T)", ext, ext)
	}

	return letter, errors.New("impl")
}

func mergeLeft(a, b map[string]any) {
	var (
		existing any
		ok       bool
	)
	for k, v := range b {
		k = strings.ToLower(k)
		existing, ok = a[k]
		if !ok { // simplest
			a[k] = v
			continue
		}

		switch existing := existing.(type) {
		case []any:
			switch v := v.(type) {
			case nil:
				continue // skip nil values
			case []any:
				a[k] = append(existing, v...)
			default:
				a[k] = append(existing, v)
			}
		case map[string]any:
			switch v := v.(type) {
			case nil:
				continue // skip nil values
			case map[string]any:
				mergeLeft(existing, v)
				// a[k] = mergeLeft(existing, v)
			default:
				a[k] = v
			}
		default:
			if v == nil {
				continue // skip nil values
			}
			a[k] = v
		}
	}
}

/*
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
*/
