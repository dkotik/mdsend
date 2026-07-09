package internal

func MergeLeft(a, b map[string]any) {
	var (
		existing any
		ok       bool
	)
	for k, v := range b {
		// k = strings.ToLower(k)
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
				MergeLeft(existing, v)
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
