package loader

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
)

type MarkdownConfiguration struct {
	Parser   parser.Parser
	Renderer renderer.Renderer
}

type Configuration struct {
	RootDirectory string
	FS            fs.FS
	Markdown      MarkdownConfiguration
}

func (c Configuration) withDefaults() (Configuration, error) {
	if c.RootDirectory == "" {
		var err error
		c.RootDirectory, err = os.Getwd()
		if err != nil {
			return c, fmt.Errorf("unable to determine root directory: %w", err)
		}
	}
	if c.FS == nil {
		// Create an unconstrained FS starting at the system root
		c.FS = os.DirFS("/")
	}

	return c, nil
}
