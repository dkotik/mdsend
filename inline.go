package mdsend

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/dkotik/mdsend/markdown"
	"github.com/yuin/goldmark/parser"
	"golang.org/x/net/html"
)

var (
	reImageTagHTML = regexp.MustCompile(`<img[^>]+>`)
)

func inlineTemplateAttachments(
	source []byte,
	replacer func(AttachmentSource) (string, error),
) (_ []byte, err error) {
	index := reImageTagHTML.FindAllIndex(source, -1)
	if index == nil {
		return source, nil
	}
	content := make([]byte, 0, len(source))
	cutoff := 0

	for _, imageRange := range index {
		content = append(content, source[cutoff:imageRange[0]]...)
		cutoff = imageRange[1]

		tokenizer := html.NewTokenizer(bytes.NewReader(source[imageRange[0]:imageRange[1]]))

		for {
			tokenType := tokenizer.Next()
			if tokenType == html.ErrorToken {
				if tokenizer.Err() == io.EOF {
					break
				}
				return source, fmt.Errorf("invalid <img> tag: %w", tokenizer.Err())
			}

			if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
				token := tokenizer.Token()
				if token.Data == "img" {
					content = append(content, []byte("<img ")...)
					attachment := AttachmentSource{}
					for i, attr := range token.Attr {
						switch attr.Key {
						case "alt":
							if attachment.Name == "" {
								attachment.Name = attr.Val
							}
						case "title":
							attachment.Name = attr.Val
						case "src":
							attachment.Location = attr.Val
							token.Attr[i].Val, err = replacer(attachment)
							if err != nil {
								return source, err
							}
						}
					}

					for _, attr := range token.Attr {
						content = append(content, []byte(attr.Key+"=\"")...)
						content = append(content, []byte(html.EscapeString(attr.Val))...)
						content = append(content, []byte("\" ")...)
					}
					content = append(content, []byte("/>")...)
				}
			}
		}
	}

	content = append(content, source[cutoff:]...)
	return content, nil
}

func inlineMarkdownAttachments(
	parser parser.Parser,
	source []byte,
	replacer func(AttachmentSource) (string, error),
) (_ string, err error) {
	index := slices.Collect(markdown.IndexImages(parser, source))
	if len(index) == 0 {
		return string(source), nil
	}
	content := make([]byte, 0, len(source))
	cutoff := 0

	for _, imgDestination := range index {
		content = append(content, source[cutoff:imgDestination.LocationAt]...)
		cutoff = imgDestination.LocationAt + len(imgDestination.Location)

		attachment := AttachmentSource{
			Name:     imgDestination.Title,
			Location: imgDestination.Location,
		}
		if attachment.Name == "" {
			attachment.Name = imgDestination.Alt
		}
		destination, err := replacer(attachment)
		if err != nil {
			return string(source), err
		}
		content = append(content, []byte(markdown.Escape(destination))...)
	}

	content = append(content, source[cutoff:]...)
	return string(content), nil
}
