package list

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

const (
	fieldName  = "name"
	fieldEmail = "email"
	fieldPhone = "phone"
	fieldNotes = "notes"
)

func (c *Contact) Name() string {
	name, ok := c.data[fieldName]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s", name)
}

func (c *Contact) SetName(ctx context.Context, s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return NewEmptyNameError(ctx)
	}
	c.data[fieldName] = s
	return c.save(ctx)
}

func (c *Contact) Email() string {
	email, ok := c.data[fieldEmail]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s", email)
}

func (c *Contact) SetEmail(ctx context.Context, s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return NewEmptyEmailError(ctx)
	}
	re, err := regexp.Compile(`^[^\@]+\@[^\@]{1,128}$`)
	if err != nil {
		return err
	}
	if !re.MatchString(s) {
		return NewInvalidEmailError(ctx)
	}
	c.data[fieldEmail] = s
	return c.save(ctx)
}

func (c *Contact) Phone() string {
	phone, ok := c.data[fieldPhone]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s", phone)
}

func (c *Contact) SetPhone(ctx context.Context, s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return c.DeleteField(ctx, fieldPhone)
	}
	re, err := regexp.Compile(`^\+?[0-9\(\)\-\s]{4,128}$`)
	if err != nil {
		return err
	}
	if !re.MatchString(s) {
		return NewInvalidPhoneError(ctx)
	}
	c.data[fieldPhone] = s
	return c.save(ctx)
}

func (c *Contact) Notes() string {
	notes, ok := c.data[fieldNotes]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s", notes)
}

func (c *Contact) SetNotes(ctx context.Context, s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return c.DeleteField(ctx, fieldNotes)
	}
	c.data[fieldNotes] = s
	return c.save(ctx)
}

func (c *Contact) DeleteField(ctx context.Context, key string) error {
	delete(c.data, key)
	return c.save(ctx)
}
