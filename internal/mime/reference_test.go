package mime

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/textproto"
)

// https://go.dev/play/p/Ifztb4dKFW2

//  multipart/mixed
//  |- multipart/alternative
//  |  |- text/plain
//  |  `- multipart/related
//  |     |- text/html
//  |     `- image/png
//  `- attachments..

func main() {
	body := &bytes.Buffer{}

	// Write mail header
	fmt.Fprintf(body, "From: %s\r\n", "Bob <bob@example.com>")
	fmt.Fprintf(body, "To: %s\r\n", "Alice <alice@example.com>")
	fmt.Fprintf(body, "Subject: %s\r\n", "A MIME mail example")
	mw := multipart.NewWriter(body)
	fmt.Fprintf(body, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(body, "Content-Type: multipart/mixed\r\n")
	fmt.Fprintf(body, "Content-Type: boundary=%s\r\n", mw.Boundary())
	fmt.Fprint(body, "\r\n")
	fmt.Fprint(body, "This is a MIME v1.0 message\r\n\r\n")

	aw := multipart.NewWriter(body)
	_, err := mw.CreatePart(textproto.MIMEHeader{"Content-Type": {"multipart/alternative", "boundary=" + aw.Boundary()}})
	if err != nil {
		log.Fatal(err)
	}

	w, err := aw.CreatePart(textproto.MIMEHeader{"Content-Type": {"plain/text"}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, "Here should appear the HTML version of the mail body\r\n")

	rw := multipart.NewWriter(body)
	_, err = aw.CreatePart(textproto.MIMEHeader{"Content-Type": {"multipart/related", "boundary=" + rw.Boundary()}})
	if err != nil {
		log.Fatal(err)
	}

	w, err = rw.CreatePart(textproto.MIMEHeader{"Content-Type": {"plain/html"}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, "Here should appear the HTML version of the mail body\r\n")

	w, err = rw.CreatePart(textproto.MIMEHeader{"Content-Type": {"image/png"}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, "Here should appear the Base64 encoded png image\r\n")

	rw.Close()
	aw.Close()

	w, err = mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"application/octet-stream"},
		"Content-Disposition": {"attachment; filename=" + "myText.txt"}})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, "Here should appear the attached file\r\n")

	mw.Close()

	fmt.Println(body.String())
}
