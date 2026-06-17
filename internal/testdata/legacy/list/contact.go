package list

import (
	"context"
)

type Contact struct {
	data   map[string]any
	source Source
	list   *List
}

func (c *Contact) Delete(ctx context.Context) error {
	var contacts []map[string]any
	for i, current := range c.list.contacts {
		if current.source != c.source {
			continue // contact is from another source
		}
		if current == c {
			c.list.contacts = append( // drop contact
				c.list.contacts[:i],
				c.list.contacts[i+1:]...,
			)
			continue
		}
		contacts = append(contacts, current.data)
	}
	return c.source.Save(ctx, contacts)
}

func (c *Contact) save(ctx context.Context) error {
	if len(c.data) < 1 {
		return NewEmptyContactError(ctx)
	}

	var contacts []map[string]any
	for _, current := range c.list.contacts {
		if current.source != c.source {
			continue // contact is from another source
		}
		contacts = append(contacts, current.data)
	}
	return c.source.Save(ctx, contacts)
}
