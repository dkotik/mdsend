package echobox

import (
	"testing"
)

func TestWordWrap(t *testing.T) {
	cases := []struct {
		Length uint8
		In     string
		Out    []string
	}{
		// {
		// 	Length: 1,
		// 	In:     `12345678`,
		// 	Out:    []string{`1`, `2`, `3`, `4`, `5`, `6`, `7`, `8`},
		// },
		{
			Length: 8,
			In:     `one two three four five six seven eight nine ten`,
			Out: []string{
				`one two`,
				`three`,
				`four`,
				`five six`,
				`seven`,
				`eight`,
				`nine ten`,
			},
		},
	}

	for _, c := range cases {
		result := WordWrap(c.In, c.Length)
		for i, line := range result {
			if uint8(len(line)) > c.Length {
				t.Log("Line:", c.In)
				t.Fatalf("word wrap failed: line %d %q exceeds length %d", i, line, c.Length)
			}
		}

		// if len(result) != len(c.Out) {
		// 	t.Logf("Expected %d pieces: %+v", len(c.Out), c.Out)
		// 	t.Logf("Received %d pieces: %+v", len(result), result)
		// 	t.Fatal("number of pieces does not match")
		// }

		for i, line := range result {
			if line != c.Out[i] {
				t.Logf("Expected: %q", c.Out[i])
				t.Logf("Received: %q", line)
				t.Fatalf("line %d does not match", i)
			}
		}
	}
}
