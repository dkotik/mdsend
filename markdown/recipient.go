package markdown

import (
	"fmt"
	"iter"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

const RecipientFilePrefix = "file://"

func (m Message) EachRecipient() iter.Seq2[map[string]any, error] {
	return func(yield func(map[string]any, error) bool) {
		// to
		for recipient, err := range eachRecipient(m.Frontmatter[ToKey], m.Directory) {
			if !yield(recipient, err) {
				return
			}
		}

		// cc
		for recipient, err := range eachRecipient(m.Frontmatter[CarbonCopyKey], m.Directory) {
			if !yield(recipient, err) {
				return
			}
		}

		// bcc
		for recipient, err := range eachRecipient(m.Frontmatter[BlindCopyKey], m.Directory) {
			if !yield(recipient, err) {
				return
			}
		}
	}
}

func eachRecipientFromFileYAML(p string) iter.Seq2[map[string]any, error] {
	return func(yield func(map[string]any, error) bool) {
		file, err := os.Open(p)
		if err != nil {
			yield(nil, err)
			return
		}
		defer func() {
			if err = file.Close(); err != nil {
				yield(nil, err)
			}
		}()
		d := yaml.NewDecoder(file)
		var entries []any
		if err = d.Decode(&entries); err != nil {
			yield(nil, err)
			return
		}
		directory := filepath.Dir(p)
		for _, entry := range entries {
			for recipient, err := range eachRecipient(entry, directory) {
				if !yield(recipient, err) {
					return
				}
			}
		}
	}
}

func eachRecipientFromFileTOML(p string) iter.Seq2[map[string]any, error] {
	return func(yield func(map[string]any, error) bool) {
		file, err := os.Open(p)
		if err != nil {
			yield(nil, err)
			return
		}
		defer func() {
			if err = file.Close(); err != nil {
				yield(nil, err)
			}
		}()
		d := toml.NewDecoder(file)
		var entries []any
		if err = d.Decode(&entries); err != nil {
			yield(nil, err)
			return
		}
		directory := filepath.Dir(p)
		for _, entry := range entries {
			for recipient, err := range eachRecipient(entry, directory) {
				if !yield(recipient, err) {
					return
				}
			}
		}
	}
}

func eachRecipient(
	source any,
	directory string,
) iter.Seq2[map[string]any, error] {
	return func(yield func(map[string]any, error) bool) {
		ok := false
		switch v := source.(type) {
		case string:
			v = strings.TrimSpace(v)
			if strings.HasPrefix(v, RecipientFilePrefix) {
				v = strings.TrimPrefix(v, RecipientFilePrefix)
				ext := strings.ToLower(filepath.Ext(v))
				if !filepath.IsAbs(v) {
					v = filepath.Join(directory, v)
				}
				switch ext {
				case ".yaml":
					for recipient, err := range eachRecipientFromFileYAML(v) {
						if !yield(recipient, err) {
							return
						}
					}
				case ".toml":
					for recipient, err := range eachRecipientFromFileTOML(v) {
						if !yield(recipient, err) {
							return
						}
					}
				// TODO: case ".cue":
				default:
					yield(nil, fmt.Errorf("recipient list file format %q is not supported", ext))
				}
				return
			}

			addresses, err := mail.ParseAddressList(v)
			if err != nil {
				yield(nil, fmt.Errorf("invalid address list: %w", err))
				return
			}
			for _, address := range addresses {
				if !yield(map[string]any{
					NameKey:  address.Name,
					EmailKey: address.Address,
				}, nil) {
					return
				}
			}
		case map[string]any:
			if _, ok = v[EmailKey]; !ok {
				yield(nil, fmt.Errorf("contact contains no electronic mail address: %s", v))
				return
			}
			if !yield(v, nil) {
				return
			}
		case []any:
			for v, err := range eachRecipient(v, directory) {
				if !yield(v, err) {
					return
				}
			}
		case nil:
		default:
			yield(nil, fmt.Errorf("data type %T is not supported for recipient list: %s", v, v))
		}
	}
}
