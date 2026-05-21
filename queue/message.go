package queue

import (
	"errors"
	"net/mail"
	"time"
)

type Message struct {
	ID      string
	Path    string
	Subject string
	From    mail.Address
	To      mail.Address
	// Template string
	Meta      map[string]any
	PlainText string
	HTML      string
	// Recipients []map[string]any
	// Attachments  []Attachment
	CreatedAt    time.Time
	DeliveredAt  time.Time
	DeliverAfter time.Time
}

func (m Message) Validate() error {
	if m.ID == "" {
		return errors.New("message ID is empty")
	}
	// if len(m.Recipients) == 0 {
	// 	return errors.New("message does not have any recipients")
	// }
	if m.CreatedAt.IsZero() {
		return errors.New("message creation time is zero")
	}
	// if m.Subject == "" {
	// 	return errors.New("message subject is empty")
	// }
	// if m.Template == "" {
	// 	return errors.New("template is empty")
	// }
	return nil
}
