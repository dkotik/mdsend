package address

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"net/mail"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func eachEntryFromFileCSV(ctx context.Context, data []byte) iter.Seq2[any, error] {
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

func eachEntryFromFileJSON(ctx context.Context, data []byte) iter.Seq2[any, error] {
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

func eachEntryFromFileYAML(ctx context.Context, data []byte) iter.Seq2[any, error] {
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

func eachEntryFromFileTOML(ctx context.Context, data []byte) iter.Seq2[any, error] {
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

func eachEntryFromFileCue(ctx context.Context, data []byte) iter.Seq2[any, error] {
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

func eachEntryFromFile(ctx context.Context, p string, fs fs.FS) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		file, err := fs.Open(p)
		if err != nil {
			yield(nil, err)
			return
		}
		data, err := io.ReadAll(file)
		if err != nil {
			yield(nil, err)
			return
		}
		if err = file.Close(); err != nil {
			yield(nil, err)
			return
		}

		ext := strings.ToLower(filepath.Ext(p))
		switch ext {
		case ".csv":
			for recipient, err := range eachEntryFromFileCSV(ctx, data) {
				if !yield(recipient, err) {
					return
				}
			}
		case ".json":
			for recipient, err := range eachEntryFromFileJSON(ctx, data) {
				if !yield(recipient, err) {
					return
				}
			}
		case ".yaml", ".yml":
			for recipient, err := range eachEntryFromFileYAML(ctx, data) {
				if !yield(recipient, err) {
					return
				}
			}
		case ".toml":
			for recipient, err := range eachEntryFromFileTOML(ctx, data) {
				if !yield(recipient, err) {
					return
				}
			}
		case ".cue":
			for recipient, err := range eachEntryFromFileCue(ctx, data) {
				if !yield(recipient, err) {
					return
				}
			}
		default:
			ok, err := isFileExecutable(p)
			if err != nil {
				yield(nil, err)
				return
			}
			if !ok {
				yield(nil, fmt.Errorf("unsupported file format for a recipient list: %s", p))
				return
			}
			for entry, err := range eachEntryFromExecutable(
				ctx,
				p, fs,
			) {
				if !yield(entry, err) {
					return
				}
			}
		}
	}
}

func eachRecipientFromEntry(
	ctx context.Context,
	entry any,
	rootDirectory string,
	fs fs.FS,
) iter.Seq2[map[string]any, error] {
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
				// fmt.Println(rootDirectory, v)
				v = path.Join(rootDirectory, v)
				// fmt.Println(v)
				subRoot := path.Dir(v)
				for entry, err := range eachEntryFromFile(ctx, v, fs) {
					if err != nil {
						yield(nil, err)
						continue
					}
					for recipient, err := range eachRecipientFromEntry(ctx, entry, subRoot, fs) {
						if !yield(recipient, err) {
							return
						}
					}
				}
				return
			case '/', '\\':
				v = path.Clean(v)
				subRoot := path.Dir(v)
				for entry, err := range eachEntryFromFile(ctx, v, fs) {
					if err != nil {
						yield(nil, err)
						continue
					}
					for recipient, err := range eachRecipientFromEntry(ctx, entry, subRoot, fs) {
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
					FieldName:  address.Name,
					FieldEmail: address.Address,
				}, nil) {
					return
				}
			}
		case map[string]any:
			if _, ok = v[FieldEmail]; !ok {
				yield(nil, ErrAbsentEmailAddress)
				return
			}
			if !yield(v, nil) {
				return
			}
		case []any:
			for _, v := range v {
				for v, err := range eachRecipientFromEntry(ctx, v, rootDirectory, fs) {
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

func Each(
	ctx context.Context,
	frontmatter map[string]any,
	rootDirectory string,
	fs fs.FS,
) iter.Seq2[map[string]any, error] {
	return func(yield func(map[string]any, error) bool) {
		to, ok := frontmatter[FieldTo]
		if ok {
			for recipient, err := range eachRecipientFromEntry(ctx, to, rootDirectory, fs) {
				if !yield(recipient, err) {
					return
				}
			}
		}

		cc, ok := frontmatter[FieldCarbonCopy]
		if ok {
			for recipient, err := range eachRecipientFromEntry(ctx, cc, rootDirectory, fs) {
				if !yield(recipient, err) {
					return
				}
			}
		}

		bcc, ok := frontmatter[FieldBlindCarbonCopy]
		if ok {
			for recipient, err := range eachRecipientFromEntry(ctx, bcc, rootDirectory, fs) {
				if !yield(recipient, err) {
					return
				}
			}
		}
	}
}

func EachUnique(
	in iter.Seq2[map[string]any, error],
) iter.Seq2[map[string]any, error] {
	if in == nil {
		panic("empty address source")
	}
	return func(yield func(map[string]any, error) bool) {
		known := make(map[string]struct{}, 64)
		for recipient, err := range in {
			if err != nil {
				yield(recipient, err)
				return
			}
			email, _ := recipient[FieldEmail].(string)
			if email == "" {
				yield(recipient, ErrAbsentEmailAddress)
				return
			}
			if _, ok := known[email]; ok {
				continue // sift out duplicate
			}
			known[email] = struct{}{}
			if !yield(recipient, err) {
				return
			}
		}
	}
}
