package markdown

type Message struct {
	Path        string
	Frontmatter map[string]any
	Content     string
}
