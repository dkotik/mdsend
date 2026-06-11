package test

import (
	"fmt"
	"iter"
	"reflect"
	"testing"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal"
)

func Queue(q mdsend.Queue) func(*testing.T) {
	l1 := mdsend.Letter{
		ID: "firstLetter",
		Frontmatter: map[string]any{
			"subject": "first letter",
		},
		Content:   "first letter",
		CreatedAt: internal.MockTime,
		SentAt:    internal.MockTime.Add(time.Hour * 70),
	}

	l2 := mdsend.Letter{
		ID: "secondLetter",
		Frontmatter: map[string]any{
			"subject": "second letter",
		},
		Content:   "second letter",
		CreatedAt: internal.MockTime.Add(time.Hour * 50),
		SentAt:    internal.MockTime.Add(time.Hour * 170),
	}

	dispatches := []mdsend.Dispatch{
		{
			LetterID: l1.ID,
			Subject:  "",
			Text:     "",
			HTML:     "",
		},
		{
			LetterID: l1.ID,
			Subject:  "",
			Text:     "",
			HTML:     "",
		},
	}

	attachments := []mdsend.Attachment{
		{LetterID: l1.ID, Content: []byte("attachment content1")},
		{LetterID: l1.ID, Content: []byte("attachment content2")},
	}
	return func(t *testing.T) {
		var (
			ctx = t.Context()
			err error
		)
		if err = q.CreateLetter(ctx, l1, attachments, dispatches); err != nil {
			t.Fatal("unable to create the first letter:", err)
		}
		defer func() {
			if err = q.DeleteLetter(ctx, l1.ID); err != nil {
				t.Fatal("unable to delete the first letter:", err)
			}
			for l1, err = range q.ListLetters(ctx, mdsend.Cursor{Batch: 1}) {
				if err != nil {
					t.Fatal("unable to collect a list of letters:", err)
				}
				t.Fatal("There is still a letter left over:", l1)
			}

			t.Run("there are no dispatches left over", IteratorIsEmpty(q.ListDispatches(ctx, mdsend.ChildCursor{
				ParentID: l1.ID,
				Cursor: mdsend.Cursor{
					ItemID: "",
					Batch:  5,
				},
			})))

			t.Run("there are no attachments left over", IteratorIsEmpty(q.ListAttachments(ctx, l1.ID)))
		}()

		lcomp, err := q.RetrieveLetter(ctx, l1.ID)
		if err != nil {
			t.Fatal("unable to retrieve first letter:", err)
		}
		t.Run("retrieved first letter matches", LettersAreEqual(l1, lcomp))

		l1attachments := make([]mdsend.Attachment, 0, len(attachments))
		for l1attachment, err := range q.ListAttachments(ctx, l1.ID) {
			if err != nil {
				t.Fatal("unable to list attachments for first letter:", err)
			}
			l1attachments = append(l1attachments, l1attachment)
		}
		if !reflect.DeepEqual(l1attachments, attachments) {
			// t.Log("expected attachments:", attachments)
			// t.Log("actual attachments:", l1attachments)
			t.Fatal("attachment lists do not match")
		}

		l1dispatches := make([]mdsend.Dispatch, 0, 1)
		for l1dispatch, err := range q.ListDispatches(ctx, mdsend.ChildCursor{
			ParentID: l1.ID,
			Cursor: mdsend.Cursor{
				ItemID: "",
				Batch:  1,
			},
		}) {
			if err != nil {
				t.Fatal("unable to list dispatches for first letter:", err)
			}
			l1dispatches = append(l1dispatches, l1dispatch)
		}
		if len(l1dispatches) != len(dispatches) {
			t.Fatal("dispatch count mismatch: expected", len(dispatches), "got", len(l1dispatches))
		}
		for i, d := range l1dispatches {
			d.ID = dispatches[i].ID // copy the ID from the expected dispatch
			t.Run(fmt.Sprintf("dispatch #%d is the same", i+1), DispatchesAreEqual(d, dispatches[i]))
			if err = q.CompleteDispatch(ctx, d.ID); err != nil {
				t.Fatalf("unable to complete dispatch %d: %v", i+1, err)
			}
		}

		if err = q.CreateLetter(ctx, l2, nil, nil); err != nil {
			t.Fatal("unable to create the second letter:", err)
		}
		defer func() {
			if err = q.DeleteLetter(ctx, l2.ID); err != nil {
				t.Fatal("unable to delete the second letter:", err)
			}

			t.Run("there are no dispatches left over", IteratorIsEmpty(q.ListDispatches(ctx, mdsend.ChildCursor{
				ParentID: l2.ID,
				Cursor: mdsend.Cursor{
					ItemID: "",
					Batch:  5,
				},
			})))

			t.Run("there are no attachments left over", IteratorIsEmpty(q.ListAttachments(ctx, l2.ID)))
		}()

		lcomp, err = q.RetrieveLetter(ctx, l2.ID)
		if err != nil {
			t.Fatal("unable to retrieve second letter:", err)
		}
		t.Run("retrieved second letter matches", LettersAreEqual(l2, lcomp))

		// test letter listing
		ok := false
		next, stop := iter.Pull2(q.ListLetters(ctx, mdsend.Cursor{Batch: 1}))
		if next == nil {
			t.Fatal("no letters found")
		}
		lcomp, err, ok = next()
		if !ok {
			t.Fatal("no letters found, when the first letter was expected")
		}
		if err != nil {
			t.Fatal("unable to retrieve the first letter:", err)
		}
		t.Run("retrieved first letter matches", LettersAreEqual(l1, lcomp))
		lcomp, err, ok = next()
		if !ok {
			t.Fatal("no letters found, when the second letter was expected")
		}
		if err != nil {
			t.Fatal("unable to retrieve the second letter:", err)
		}
		t.Run("retrieved second letter matches", LettersAreEqual(l2, lcomp))
		stop()
	}
}
