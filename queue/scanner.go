package queue

import (
	"context"
	"iter"
	"time"

	"github.com/dkotik/mdsend"
)

func NewScanner(
	ctx context.Context,
	queue mdsend.Queue,
	frequency time.Duration,
	batchSize int64,
) (task func() (err error), pendingMessageIDs chan []string) {
	pendingMessageIDs = make(chan []string)
	return func() (err error) {
		defer close(pendingMessageIDs)
		nextLetter := mdsend.Cursor{
			ItemID: "",
			Batch:  1,
		}
		nextMessage := mdsend.ChildCursor{
			ParentID: "",
			Cursor: mdsend.Cursor{
				ItemID: "",
				Batch:  batchSize,
			},
		}

		for {
			for {
				letterPull, letterStop := iter.Pull2[mdsend.Letter, error](queue.ListLetters(ctx, nextLetter))
				letter, err, ok := letterPull()
				letterStop()
				if err != nil {
					return err
				}
				if !ok {
					break
				}

				nextLetter.ItemID = letter.ID
				nextMessage.ParentID = letter.ID

				for {
					batch := make([]string, 0, nextMessage.Batch)
					messagePull, messageStop := iter.Pull2[mdsend.Dispatch, error](queue.ListDispatches(ctx, nextMessage))
					for range nextMessage.Batch {
						message, err, ok := messagePull()
						if err != nil {
							messageStop()
							return err
						}
						if !ok {
							nextMessage.Cursor.ItemID = ""
							break
						}
						nextMessage.Cursor.ItemID = message.ID
						if message.SentAt.IsZero() {
							batch = append(batch, message.ID)
						}
					}
					messageStop()
					select {
					case <-ctx.Done():
						return ctx.Err()
					case pendingMessageIDs <- batch:
						// delivered a batch of discovered messages that were not yet sent
					}

					if nextMessage.Cursor.ItemID == "" {
						break
					}
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(frequency):
					}
				}
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(frequency):
				nextLetter.ItemID = "" // start the scan over
			}
		}
	}, pendingMessageIDs
}
