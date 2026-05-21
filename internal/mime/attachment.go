package mime

type Attachment struct {
	Name        string
	ContentType string
	Content     []byte
}
