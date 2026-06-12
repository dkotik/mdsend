package queue

import (
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
