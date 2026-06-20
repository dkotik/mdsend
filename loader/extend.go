package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/markdown"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func (l loader) extend(ctx context.Context, letter mdsend.Letter, rootDirectory string, known map[string]struct{}) (_ mdsend.Letter, err error) {
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

	for _, p := range extends {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "/") {
			p = path.Clean(p)
		} else {
			p = path.Join(rootDirectory, p)
		}
		if _, ok := known[p]; ok {
			return letter, fmt.Errorf("infinite import cycle detected: %s", p)
		}
		known[p] = struct{}{}
		data, err := l.getFile(ctx, p)
		if err != nil {
			return letter, err
		}

		switch ext := strings.ToLower(path.Ext(p)); ext {
		case ".json":
			var frontmatter map[string]any
			if err := json.Unmarshal(data, &frontmatter); err != nil {
				return letter, err
			}
			if frontmatter == nil {
				continue
			}
			mergeLeft(frontmatter, letter.Frontmatter)
			letter.Frontmatter = frontmatter
			continue
		case ".yaml", ".yml":
			var frontmatter map[string]any
			if err := yaml.Unmarshal(data, &frontmatter); err != nil {
				return letter, err
			}
			if frontmatter == nil {
				continue
			}
			mergeLeft(frontmatter, letter.Frontmatter)
			letter.Frontmatter = frontmatter
			continue
		case ".toml":
			var frontmatter map[string]any
			if err := toml.Unmarshal(data, &frontmatter); err != nil {
				return letter, err
			}
			if frontmatter == nil {
				continue
			}
			mergeLeft(frontmatter, letter.Frontmatter)
			letter.Frontmatter = frontmatter
			continue
		case ".cue":
			var frontmatter map[string]any
			if err := cuecontext.New().CompileBytes(data).Decode(&frontmatter); err != nil {
				return letter, err
			}
			if frontmatter == nil {
				continue
			}
			mergeLeft(frontmatter, letter.Frontmatter)
			letter.Frontmatter = frontmatter
			continue
		case ".md", ".markdown":
			// drop down
		default:
			return letter, fmt.Errorf("unsupported extension file type: %s", ext)
		}

		frontmatterRaw, body, delimeter, err := markdown.Cut(data)
		if err != nil {
			return letter, err
		}
		frontmatter, err := markdown.ParseFrontmatter(frontmatterRaw, delimeter)
		if err != nil {
			return letter, err
		}
		subLetter := mdsend.Letter{
			Frontmatter: frontmatter,
			Content:     string(body),
		}
		subLetter, err = l.extend(ctx, subLetter, path.Dir(p), known)
		if err != nil {
			return letter, err
		}
		most, tail, _ := markdown.SplitOnLastHorizontalRule([]byte(subLetter.Content))
		// panic(tail)
		letter.Content = string(most) + letter.Content + "\n\n" + string(tail)
		if subLetter.Frontmatter != nil {
			mergeLeft(subLetter.Frontmatter, letter.Frontmatter)
		}
	}

	return letter, nil
}

func mergeLeft(a, b map[string]any) {
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
