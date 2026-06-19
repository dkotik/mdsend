package loader

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/mail"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"github.com/dkotik/mdsend"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type Recipients iter.Seq2[map[string]any, error]

func eachEntryFromFileCSV(data []byte) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		r := csv.NewReader(bytes.NewReader(data))
		headers, err := r.Read()
		if err != nil {
			yield(nil, err)
			return
		}
		headerCount := len(headers)

		for {
			row, err := r.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				yield(nil, err)
				continue
			}

			entry := make(map[string]any)
			for i, value := range row[:min(len(row), headerCount)] {
				entry[headers[i]] = value
			}

			if !yield(entry, nil) {
				return
			}
		}
	}
}

func eachEntryFromFileJSON(data []byte) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		d := json.NewDecoder(bytes.NewReader(data))
		var entries []any
		if err := d.Decode(&entries); err != nil {
			yield(nil, err)
			return
		}
		for _, entry := range entries {
			if !yield(entry, nil) {
				return
			}
		}
	}
}

func eachEntryFromFileYAML(data []byte) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		d := yaml.NewDecoder(bytes.NewReader(data))
		var entries []any
		if err := d.Decode(&entries); err != nil {
			yield(nil, err)
			return
		}
		for _, entry := range entries {
			if !yield(entry, nil) {
				return
			}
		}
	}
}

func eachEntryFromFileTOML(data []byte) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		d := toml.NewDecoder(bytes.NewReader(data))
		var entries map[string]any
		if err := d.Decode(&entries); err != nil {
			yield(nil, err)
			return
		}

		for key, entry := range entries {
			list, ok := entry.([]any)
			if !ok {
				yield(nil, fmt.Errorf("TOML entry is not a list: %s", key))
				return
			}
			for _, entry := range list {
				if !yield(entry, nil) {
					return
				}
			}
		}
	}
}

func eachEntryFromFileCue(data []byte) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		ctx := cuecontext.New()
		v := ctx.CompileBytes(data)
		var entries []any
		if err := v.Decode(&entries); err != nil {
			yield(nil, err)
			return
		}
		for _, entry := range entries {
			if !yield(entry, nil) {
				return
			}
		}
	}
}

func (l loader) eachEntryFromFile(p string, cache Cache) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		ctx := context.Background()
		data, err := cache.Pull(ctx, p)
		if err != nil {
			yield(nil, err)
			return
		}
		if data != nil {
			return // prevent circular imports
		}
		file, err := l.FS.Open(p)
		if err != nil {
			yield(nil, err)
			return
		}
		if data, err = io.ReadAll(file); err != nil {
			yield(nil, err)
			return
		}
		if err = file.Close(); err != nil {
			yield(nil, err)
			return
		}
		if err = cache.Push(ctx, p, data); err != nil {
			yield(nil, err)
			return
		}
		// b := &bytes.Buffer{}
		// hash := xxhash.New()
		// if _, err = io.Copy(io.MultiWriter(b, hash), file); err != nil {
		// 	yield(nil, err)
		// 	return
		// }
		// key := fmt.Sprintf("%x", hash.Sum64())
		// data, ok := l.Cache[key]
		// if ok {
		// 	return // only process each file one
		// }
		// data = b.Bytes()
		// l.Cache[key] = data

		ext := strings.ToLower(filepath.Ext(p))
		switch ext {
		case ".csv":
			for recipient, err := range eachEntryFromFileCSV(data) {
				if !yield(recipient, err) {
					return
				}
			}
		case ".json":
			for recipient, err := range eachEntryFromFileJSON(data) {
				if !yield(recipient, err) {
					return
				}
			}
		case ".yaml", ".yml":
			for recipient, err := range eachEntryFromFileYAML(data) {
				if !yield(recipient, err) {
					return
				}
			}
		case ".toml":
			for recipient, err := range eachEntryFromFileTOML(data) {
				if !yield(recipient, err) {
					return
				}
			}
		case ".cue":
			for recipient, err := range eachEntryFromFileCue(data) {
				if !yield(recipient, err) {
					return
				}
			}
		default:
			yield(nil, fmt.Errorf("unsupported file format for a recipient list: %s", p))
			return
		}
	}
}

func (l loader) eachRecipientFromEntry(
	entry any,
	rootDirectory string,
	cache Cache,
) Recipients {
	return func(yield func(map[string]any, error) bool) {
		ok := false
		switch v := entry.(type) {
		case string:
			v = strings.TrimSpace(v)
			if v == "" {
				return
			}
			switch v[0] {
			case '.':
				v = path.Join(rootDirectory, v)
				subRoot := path.Dir(v)
				for entry, err := range l.eachEntryFromFile(v, cache) {
					if err != nil {
						yield(nil, err)
						continue
					}
					for recipient, err := range l.eachRecipientFromEntry(entry, subRoot, cache) {
						if !yield(recipient, err) {
							return
						}
					}
				}
				return
			case '/', '\\':
				v = path.Clean(v)
				subRoot := path.Dir(v)
				for entry, err := range l.eachEntryFromFile(v, cache) {
					if err != nil {
						yield(nil, err)
						continue
					}
					for recipient, err := range l.eachRecipientFromEntry(entry, subRoot, cache) {
						if !yield(recipient, err) {
							return
						}
					}
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
					mdsend.FieldNameName:  address.Name,
					mdsend.FieldNameEmail: address.Address,
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
			for _, v := range v {
				for v, err := range l.eachRecipientFromEntry(v, rootDirectory, cache) {
					if !yield(v, err) {
						return
					}
				}
			}
		case nil:
		default:
			yield(nil, fmt.Errorf("data type %T is not supported for recipient list: %s", v, v))
		}
	}
}

func (l loader) eachRecipient(frontmatter map[string]any, rootDirectory string) Recipients {
	return func(yield func(map[string]any, error) bool) {
		cache := NewMapCache()
		to, ok := frontmatter[mdsend.FieldNameTo]
		if ok {
			for recipient, err := range l.eachRecipientFromEntry(to, rootDirectory, cache) {
				if !yield(recipient, err) {
					return
				}
			}
		}

		cc, ok := frontmatter[mdsend.FieldNameCarbonCopy]
		if ok {
			for recipient, err := range l.eachRecipientFromEntry(cc, rootDirectory, cache) {
				if !yield(recipient, err) {
					return
				}
			}
		}

		bcc, ok := frontmatter[mdsend.FieldNameBlindCarbonCopy]
		if ok {
			for recipient, err := range l.eachRecipientFromEntry(bcc, rootDirectory, cache) {
				if !yield(recipient, err) {
					return
				}
			}
		}
	}
}
