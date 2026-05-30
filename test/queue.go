package test

import (
	"testing"

	"github.com/dkotik/mdsend"
)

func Queue(q mdsend.Queue) func(*testing.T) {
	l1 := mdsend.Letter{
		ID: "firstLetter",
	}
	l2 := mdsend.Letter{
		ID: "secondLetter",
	}
	return func(t *testing.T) {
		var (
			ctx = t.Context()
			err error
		)
		if err = q.CreateLetter(ctx, l1, nil, nil); err != nil {
			t.Fatal("unable to create the first letter:", err)
		}
		t.Cleanup(func() {
			if err = q.DeleteLetter(ctx, l1.ID); err != nil {
				t.Fatal("unable to delete the first letter:", err)
			}
		})
		lcomp, err := q.RetrieveLetter(ctx, l1.ID)
		if err != nil {
			t.Fatal("unable to retrieve first letter:", err)
		}
		t.Run("retrieved first letter matches", LettersAreEqual(l1, lcomp))

		if err = q.CreateLetter(ctx, l2, nil, nil); err != nil {
			t.Fatal("unable to create the second letter:", err)
		}
		t.Cleanup(func() {
			if err = q.DeleteLetter(ctx, l2.ID); err != nil {
				t.Fatal("unable to delete the second letter:", err)
			}
		})
		lcomp, err = q.RetrieveLetter(ctx, l2.ID)
		if err != nil {
			t.Fatal("unable to retrieve second letter:", err)
		}
		t.Run("retrieved second letter matches", LettersAreEqual(l2, lcomp))
	}
}
