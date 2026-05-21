package markdown

import "context"

type Sender interface {
	Send(context.Context, Message) error
}
