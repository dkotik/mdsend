package tealist_test

import (
	"context"
	"testing"

	"github.com/dkotik/mdsend/list"
	"github.com/dkotik/mdsend/list/tealist"
)

func TestTeaList(t *testing.T) {
	ctx := context.Background()
	l, err := list.NewList(
		ctx,
		"../testdata/source.json",
		"../testdata/source.yaml",
	)
	if err != nil {
		t.Error(err)
	}

	bl := &tealist.List{*l}
	t.Log(bl.View())
}
