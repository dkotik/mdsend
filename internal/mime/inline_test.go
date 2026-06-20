package mime

import "testing"

func TestInlineImageDetection(t *testing.T) {
	references := []string{
		"1",
		"2",
		"3xa",
		"4a",
	}

	i := 0
	for ref := range eachInlineReferenceInHTML([]byte(`
		<img src='cid:1' />
		<img src="cid:2" />
		<img src=cid:3xa />
		<img src=nocid >
		<img src=cid:4a >
	`)) {
		if ref != references[i] {
			t.Log("A:", references[i])
			t.Log("B:", ref)
			t.Error("references do not match:", i+1)
		}
		i++
		if i == len(references) {
			break
		}
	}
	if i < len(references) {
		t.Errorf("expected %d references, got %d", len(references), i)
	}
}
