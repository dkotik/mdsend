/*
Package list models a composite mailing list that is assembled from multiple
data sources.
*/
package list

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type List struct {
	contacts []*Contact
	sources  []Source
}

func NewList(ctx context.Context, sources ...string) (l *List, err error) {
	l = &List{}
	for _, s := range sources {
		switch ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(s), ".")); ext {
		case "JSON":
			if err = l.AddSource(ctx, sourceJSON(s)); err != nil {
				return nil, err
			}
		case "YAML":
			if err = l.AddSource(ctx, sourceYAML(s)); err != nil {
				return nil, err
			}
		default:
			return nil, NewUnsupportedSourceError(ctx, ext)
		}
	}
	return l, nil
}

func (l *List) AddSource(ctx context.Context, s Source) (err error) {
	data, err := s.Load(ctx)
	if err != nil {
		return err
	}
	contacts := make([]*Contact, len(data))
	for i, one := range data {
		contacts[i] = &Contact{
			data:   one,
			source: s,
			list:   l,
		}
	}
	l.contacts = slices.Concat(l.contacts, contacts)
	l.sources = append(l.sources, s)
	return nil
}

func (l *List) ReplaceSource(original, replacement Source) (err error) {
	// search := original.Location()
	for i, current := range l.sources {
		if original == current {
			if replacement == nil {
				return errors.New("<nil> replacement source")
			}
			l.sources[i] = replacement
			for _, contact := range l.contacts {
				if contact.source == original {
					contact.source = replacement
					// panic(contact.source.Location())
				}
			}
			return nil
		}
	}
	return fmt.Errorf("unknown source: %s", original.Location())
}

func (l *List) AddContact(toSource Source) (*Contact, error) {
	for _, current := range l.sources {
		if toSource == current {
			contact := &Contact{
				data:   make(map[string]any),
				source: toSource,
				list:   l,
			}
			l.contacts = append(l.contacts, contact)
			return contact, nil
		}
	}
	return nil, fmt.Errorf("unknown source: %s", toSource.Location())
}
