package loaders

import (
	"fmt"
	"io"
)

// const bl = 13
// const bl = 200
const bl = 2048

// NewDocumentChunkReader returns a reader with pre-loaded buffer.
func NewDocumentChunkReader(r io.Reader, boundary []byte) *DocumentChunkReader {
	l := len(boundary)
	if l > bl/4 {
		panic(fmt.Errorf(`DocumentChunkReader will be inefficient with boundary length %d`, l))
	}
	d := &DocumentChunkReader{
		r:  r,
		d:  boundary,
		dl: l,
	}
	d.cb, _ = io.ReadAtLeast(r, d.b[:], bl)
	return d
}

// DocumentChunkReader provides a reader interface to a document reader.
// Seperates the stream into chunks by detecting provided boundary.
type DocumentChunkReader struct {
	r  io.Reader // Source
	d  []byte    // Signature that splits the documents from the Reader.
	dl int       // boundary length: len(Boundary)
	dc int       // boundary byte counter
	b  [bl]byte  // reading buffer
	ca int       // Cursor position in the buffer, used for chunking
	cb int       // last position of the buffer
}

// Next returns "true" if there is another document that can be read in the stream.
func (d *DocumentChunkReader) Next() bool {
	if d.cb == 0 {
		return false
	}
	return true
}

func (d *DocumentChunkReader) lookAhead() bool {
	length := d.cb - d.ca
	examine := d.dl - d.dc // how many bytes to examine
	if length < examine {  // re-fill buffer if it less than bytes needed
		// panic(`moved bytes`)
		for i := 0; i < length; i++ {
			d.b[i] = d.b[i+d.ca] // move all bytes to the front of the buffer
		}
		d.ca = 0
		d.cb, _ = io.ReadFull(d.r, d.b[length:]) // fill the rest of the buffer
		length += d.cb
	}
	if length < examine {
		return false // do not look further than bytes you got
	}

	for i := 0; i < examine; i++ {
		if d.b[d.ca+i] != d.d[d.dc+i] { // break if boundary char does not match
			d.dc = 0 // reset boundary counter, it is no longer needed
			return false
		}
	}
	return true
}

func (d *DocumentChunkReader) Read(b []byte) (n int, err error) {
	for ; n < len(b); n++ {
		if d.ca == d.cb { // reached top boundary
			// fmt.Printf("\n { REFILLING BUFFER %d/%d }\n", d.ca, d.cb)
			d.cb, err = io.ReadAtLeast(d.r, d.b[:], bl)
			if d.cb == 0 { // nothing new to read
				return n, err
			}
			d.ca = 0
		}
		b[n] = d.b[d.ca]
		d.ca++
		if b[n] != d.d[d.dc] { // reset counter if boundary pattern ends
			d.dc = 0
		}
		if b[n] == d.d[d.dc] {
			// fmt.Printf(" || %q <> %q || %d\n", b[n], d.d[d.dc], d.dc)
			d.dc++            // buffer byte matched boundary byte
			if d.dc == d.dl { // the entire boundary was located! CUT!
				d.dc = 0 // reset boundary counter
				// fmt.Print("\n=========== PRINTED AHEAD +================\n")
				return n - d.dl + 1, io.EOF
			}
		}
	}
	if d.dc > 0 && d.lookAhead() { // deal with the boundary leftover
		// fmt.Print("=========== LOAKEAD CAUGHT +================\n")
		d.ca += d.dl - d.dc // go forward in the buffer
		n -= d.dc           // don't write partial boundary to reader
		d.dc = 0            // reset boundary counter
		return n, io.EOF
	}
	return n, nil
}
