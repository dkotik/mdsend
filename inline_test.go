package mdsend

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/dkotik/mdsend/markdown"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
)

func TestInlineAttachmentInjectionToTemplate(t *testing.T) {
	tmpl, err := inlineTemplateAttachments(
		[]byte(`
			<html><head><title>Test</title></head>
			<body><img src="test1" alt="t1"> <hr /> <img src="test2" alt="t2" /></body></html>
		`),
		func(a AttachmentSource) (string, error) {
			return "removed:" + a.Location + "::", nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(tmpl))
	// t.Fatal("woo")

	content, err := html.Parse(bytes.NewReader(tmpl))
	if err != nil {
		t.Fatal(err)
	}
	sources := []string{}
	for node := range content.Descendants() {
		if node.Type != html.ElementNode || node.Data != "img" {
			continue
		}
		for _, attr := range node.Attr {
			if attr.Key != "src" {
				continue
			}
			sources = append(sources, attr.Val)
		}
	}
	if len(sources) != 2 {
		t.Fatal("expected 2 sources, got", len(sources))
	}
	if sources[0] != "removed:test1::" {
		t.Fatal("src should be test1", sources)
	}
	if sources[1] != "removed:test2::" {
		t.Fatal("src should be test2", sources)
	}
}

func TestInlineMarkdownAttachments(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("markdown", "testdata", "inline.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty data")
	}

	parser := markdown.NewParser(markdown.DefaultLightTheme)
	destinations := []string{}
	result, err := inlineMarkdownAttachments(
		parser,
		data,
		func(a AttachmentSource) (string, error) {
			destinations = append(destinations, a.Location)
			return "removed:" + a.Location + "::", nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(destinations) != 3 {
		t.Fatal("expected 3 destinations, got", len(destinations))
	}
	if destinations[0] != "ref1URL" {
		t.Error("dest should be ref1URL", destinations[0])
	}
	if destinations[1] != "img2URL" {
		t.Error("dest should be img2URL", destinations[1])
	}
	if destinations[2] != "ref3URL" {
		t.Error("dest should be ref3URL", destinations[2])
	}

	tree := parser.Parse(text.NewReader([]byte(result)))
	recovered := []string{}
	err = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if img, ok := n.(*ast.Image); ok {
			d := string(img.Destination)
			if d != "removed:"+destinations[len(recovered)]+"::" {
				t.Errorf("expected %s, got %s", "removed:"+destinations[len(recovered)]+"::", d)
			}
			recovered = append(recovered, d)
		}
		return ast.WalkContinue, nil
	})
}
