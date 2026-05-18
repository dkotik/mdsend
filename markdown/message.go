package markdown

import (
	"bytes"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/cespare/xxhash/v2"
	"github.com/pelletier/go-toml"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

type Message struct {
	ID          string
	Path        string
	Directory   string
	Attachments map[string]string
	Frontmatter map[string]any
	Content     string
}

func NewMessage(p string) (m Message, err error) {
	m.Path, err = filepath.Abs(p)
	if err != nil {
		return m, fmt.Errorf("unable to locate absolute path for %q: %w", m.Path, err)
	}

	raw, err := os.ReadFile(m.Path)
	if err != nil {
		return
	}
	frontmatter, content, delimeter, err := Cut(raw)
	if err != nil {
		return
	}
	switch delimeter {
	case delimeterYAML:
		if err = yaml.NewDecoder(bytes.NewReader(frontmatter)).Decode(&m.Frontmatter); err != nil {
			return m, fmt.Errorf("invalid YAML front-matter: %w", err)
		}
	case delimeterTOML:
		if err = toml.NewDecoder(bytes.NewReader(frontmatter)).Decode(&m.Frontmatter); err != nil {
			return m, fmt.Errorf("invalid TOML front-matter: %w", err)
		}
	default:
		return m, fmt.Errorf("unsupported front-matter delimeter: %s", string(delimeter))
	}
	if m.Frontmatter == nil {
		m.Frontmatter = make(map[string]any)
	}

	m.Attachments = make(map[string]string)
	m.Directory = filepath.Dir(m.Path)
	addAttachment := func(original string) {
		p := strings.TrimSpace(original)
		if p == "" {
			return
		}
		p = filepath.Clean(p)
		if !filepath.IsAbs(p) {
			p = filepath.Join(m.Directory, p)
		}
		m.Attachments[original] = p
	}

	switch attachments := m.Frontmatter[AttachmentsKey].(type) {
	case nil: // ignore
	case []any:
		for _, attachment := range attachments {
			p, ok := attachment.(string)
			if !ok {
				return m, fmt.Errorf("front-matter attachment %q is not a path: %T", attachment, attachment)
			}
			addAttachment(p)
		}
	default:
		return m, fmt.Errorf("front-matter attachments are not a list: %T", attachments)
	}

	ast.Walk(
		New().Parser().Parse(text.NewReader(content)),
		ast.Walker(func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				if n, ok := n.(*ast.Image); ok {
					addAttachment(string(n.Destination))
					return ast.WalkSkipChildren, nil
				}
			}
			return ast.WalkContinue, nil
		}))
	m.Content = string(content)

	id, ok := m.Frontmatter[IdempotentIdentifierKey]
	if ok {
		m.ID = strings.TrimSpace(fmt.Sprintf("%s", id))
	}
	if m.ID == "" {
		hash := xxhash.New()
		if _, err = hash.WriteString(m.Path); err != nil {
			return m, err
		}
		if _, err = hash.Write(frontmatter); err != nil {
			return m, err
		}
		if _, err = hash.Write(content); err != nil {
			return m, err
		}
		m.ID = base58.Encode(hash.Sum(nil))
	}
	m.Frontmatter[IdempotentIdentifierKey] = m.ID
	return
}

// EachAttachment returns a unique absolute path for
// each letter attachment.
func (m Message) EachAttachment() iter.Seq[string] {
	return func(yield func(string) bool) {
		known := make(map[string]struct{})
		ok := false
		for _, absolutePath := range m.Attachments {
			if _, ok = known[absolutePath]; ok {
				continue
			}
			known[absolutePath] = struct{}{}
			if !yield(absolutePath) {
				return
			}
		}
	}
}
