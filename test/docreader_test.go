package tests

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/dkotik/mdsend/loaders"
)

const doc = `---
1 sds fsdf
---
[doc]
---
[2]
---

[3]
[many more]

---

---





[above one is empty]


~~
---
finally||?`

func TestDoc(t *testing.T) {
	it := 0
	// handle, err := os.Open(``)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer handle.Close()
	// r := NewDocumentChunkReader(handle, []byte("\n---\n"))
	r := loaders.NewDocumentChunkReader(bytes.NewBuffer([]byte(doc)), []byte("\n---\n"))
	for r.Next() {
		it++
		if it > 25 {
			log.Fatal(`MAX ITERATIONS REACHED`)
		}
		b := make([]byte, 9)
		for {
			n, err := r.Read(b[:])
			// fmt.Print(string(b), "|", n, r.Error)
			fmt.Printf(`%q`, b[:n])
			// fmt.Print(string(b[:n]), "^[", n, "]")
			if err != nil || n == 0 {
				break
			}
		}
		// b := bytes.NewBuffer(nil)
		// io.Copy(b, r)
		// fmt.Print("\n`````````````\n", b.String(), "$")
		fmt.Print("\n............os.ReadFull............\n")
	}
	// t.Fail()
}
