package mdsend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"github.com/dkotik/mdsend/internal"
	"github.com/dkotik/mdsend/internal/media"
	"github.com/dkotik/mdsend/markdown"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func injectPathPrefix(
	frontmatter map[string]any,
	rootDirectory string,
) (err error) {
	templates, err := getTemplates(frontmatter)
	if err != nil {
		return err
	}
	updated := false
	for i, t := range templates {
		if media.IsPathLocal(t) {
			updated = true
			templates[i] = path.Join(rootDirectory, t)
		}
	}
	if updated {
		frontmatter[FieldNameTemplates] = any(templates).([]any)
	}

	if attachments, ok := frontmatter[FieldNameAttachments]; ok {
		frontmatter[FieldNameAttachments] = media.AppendPathPrefixToExternalRecipientListEntries(rootDirectory, attachments)
	}
	if to, ok := frontmatter[FieldNameTo]; ok {
		frontmatter[FieldNameTo] = media.AppendPathPrefixToExternalRecipientListEntries(rootDirectory, to)
	}
	if cc, ok := frontmatter[FieldNameCarbonCopy]; ok {
		frontmatter[FieldNameCarbonCopy] = media.AppendPathPrefixToExternalRecipientListEntries(rootDirectory, cc)
	}
	if bcc, ok := frontmatter[FieldNameBlindCarbonCopy]; ok {
		frontmatter[FieldNameBlindCarbonCopy] = media.AppendPathPrefixToExternalRecipientListEntries(rootDirectory, bcc)
	}
	return nil
}

// use [media.NewCyclicalImportPreventingFileSystem] only
func extend(
	ctx context.Context,
	letter Letter,
	rootDirectory string,
	fs fs.FS,
) (_ Letter, err error) {
	select {
	case <-ctx.Done():
		return letter, ctx.Err()
	default:
	}

	ext, ok := letter.Frontmatter[FieldNameExtends]
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

		file, err := fs.Open(p)
		if err != nil {
			return letter, err
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return letter, errors.Join(err, file.Close())
		}
		if err = file.Close(); err != nil {
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
			if rootDirectory != "" && rootDirectory != "." {
				if err = injectPathPrefix(frontmatter, rootDirectory); err != nil {
					return letter, err
				}
			}
			internal.MergeLeft(frontmatter, letter.Frontmatter)
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
			if rootDirectory != "" && rootDirectory != "." {
				if err = injectPathPrefix(frontmatter, rootDirectory); err != nil {
					return letter, err
				}
			}
			internal.MergeLeft(frontmatter, letter.Frontmatter)
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
			if rootDirectory != "" && rootDirectory != "." {
				if err = injectPathPrefix(frontmatter, rootDirectory); err != nil {
					return letter, err
				}
			}
			internal.MergeLeft(frontmatter, letter.Frontmatter)
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
			if rootDirectory != "" && rootDirectory != "." {
				if err = injectPathPrefix(frontmatter, rootDirectory); err != nil {
					return letter, err
				}
			}
			internal.MergeLeft(frontmatter, letter.Frontmatter)
			letter.Frontmatter = frontmatter
			continue
		case ".md", ".markdown":
			// drop down
		default:
			return letter, fmt.Errorf("unsupported extension file type: %s", ext)
		}

		subLetter, err := NewLetter(data)
		if err != nil {
			return letter, err
		}
		subLetter, err = extend(ctx, subLetter, path.Dir(p), fs)
		if err != nil {
			return letter, err
		}

		if rootDirectory != "" && rootDirectory != "." {
			b := &bytes.Buffer{}
			if err = markdown.CopyWithRelativePathPrefix(b, []byte(subLetter.Content), rootDirectory); err != nil {
				return letter, err
			}
			if err = injectPathPrefix(subLetter.Frontmatter, rootDirectory); err != nil {
				return letter, err
			}
		}

		most, tail, _ := markdown.SplitOnLastHorizontalRule([]byte(subLetter.Content))
		most = bytes.TrimSpace(most)
		tail = bytes.TrimSpace(tail)
		b := &bytes.Buffer{}
		b.Grow(len(most) + len(letter.Content) + len(tail) + 200)

		if len(most) > 0 {
			_, _ = fmt.Fprintf(b, "\n\n<!-- begin content from %s --->\n\n", p)
			_, _ = io.Copy(b, bytes.NewReader(most))
			_, _ = fmt.Fprintf(b, "\n\n<!-- end content from %s --->\n\n", p)
		}

		_, _ = io.WriteString(b, letter.Content)

		if len(tail) > 0 {
			_, _ = fmt.Fprintf(b, "\n\n<!-- begin footer from %s --->\n\n", p)
			_, _ = io.Copy(b, bytes.NewReader(tail))
			_, _ = fmt.Fprintf(b, "\n\n<!-- end footer from %s --->\n\n", p)
		}
		letter.Content = b.String()
		if subLetter.Frontmatter != nil {
			internal.MergeLeft(subLetter.Frontmatter, letter.Frontmatter)
			letter.Frontmatter = subLetter.Frontmatter
		}
	}

	return letter, nil
}
