package list

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

type Source interface {
	Location() string
	Load(context.Context) ([]map[string]any, error)
	Save(context.Context, []map[string]any) error
}

type sourceJSON string

func (s sourceJSON) Location() string {
	return string(s)
}

func (s sourceJSON) Load(_ context.Context) (v []map[string]any, err error) {
	b, err := os.ReadFile(s.Location())
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s sourceJSON) Save(_ context.Context, v []map[string]any) (err error) {
	w, err := NewWriter(s.Location())
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, w.Close())
	}()
	return json.NewEncoder(w).Encode(v)
}

type sourceYAML string

func (s sourceYAML) Location() string {
	return string(s)
}

func (s sourceYAML) Load(_ context.Context) (v []map[string]any, err error) {
	b, err := os.ReadFile(s.Location())
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s sourceYAML) Save(_ context.Context, v []map[string]any) (err error) {
	w, err := NewWriter(s.Location())
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, w.Close())
	}()
	return yaml.NewEncoder(w).Encode(v)
}
