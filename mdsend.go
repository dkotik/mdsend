package mdsend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/dkotik/mdsend/markdown"
	"github.com/dkotik/mdsend/media"
	"github.com/oklog/ulid/v2"
	"github.com/yuin/goldmark/parser"
)

type FileReadError struct {
	Path  string
	Cause error
}

func (err FileReadError) Error() string {
	return fmt.Sprintf(
		"file <%s> cannot be read: %v",
		err.Path,
		err.Cause,
	)
}

func (err FileReadError) Unwrap() error {
	return err.Cause
}

func NewFileReadError(p string, cause error) error {
	return FileReadError{
		Path:  p,
		Cause: cause,
	}
}

type IdentifierGenerator interface {
	GenerateID() (string, error)
}

type IdentifierGeneratorFunc func() (string, error)

func (f IdentifierGeneratorFunc) GenerateID() (string, error) {
	return f()
}

type Mailer interface {
	SendMail(context.Context, Message) (string, error)
}

type Loader interface {
	LoadLetter(context.Context, string) (Letter, []Attachment, error)
}

type Defaults struct {
	LetterIdentifierGenerator IdentifierGenerator
	// Language                  language.Tag
	MediaContraints media.Constraints
	Schedule        Schedule
}

type loader struct {
	LetterIdentifierGenerator IdentifierGenerator
	FileSystem                fs.FS
	// DefaultLanguage           language.Tag
	DefaultMediaContraints media.Constraints
	DefaultSchedule        Schedule
	Parser                 parser.Parser
}

func New(fs fs.FS, options Defaults) (_ Loader, err error) {
	if fs == nil {
		return nil, errors.New("nil file system")
	}
	options.MediaContraints = options.MediaContraints.WithDefaults()
	if err = options.MediaContraints.Validate(); err != nil {
		return nil, err
	}
	if err = options.Schedule.Validate(); err != nil {
		return nil, err
	}
	// if options.Language.String() == "und" {
	// 	options.Language = language.English
	// }
	// if !locale.IsValidLanguageTag(options.Language) {
	// 	return nil, fmt.Errorf("invalid language choice: %s", options.Language.String())
	// }
	if options.LetterIdentifierGenerator == nil {
		options.LetterIdentifierGenerator = IdentifierGeneratorFunc(func() (string, error) {
			id, err := ulid.New(ulid.Now(), ulid.DefaultEntropy())
			if err != nil {
				return "", err
			}
			return id.String(), nil
		})
	}
	return loader{
		LetterIdentifierGenerator: options.LetterIdentifierGenerator,
		FileSystem:                fs,
		// DefaultLanguage:           options.Language,
		DefaultMediaContraints: options.MediaContraints,
		DefaultSchedule:        options.Schedule,
		Parser:                 markdown.NewParser(markdown.DefaultLightTheme), // to options
	}, nil
}

func (loader loader) loadLetterFromFile(
	ctx context.Context,
	p string,
	rootDirectory string,
) (letter Letter, err error) {
	file, err := loader.FileSystem.Open(p)
	if err != nil {
		return letter, NewFileReadError(p, err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return letter, NewFileReadError(p, errors.Join(err, file.Close()))
	}
	if err = file.Close(); err != nil {
		return letter, NewFileReadError(p, err)
	}

	letter, err = newLetter(data)
	if err != nil {
		return letter, NewFileReadError(p, err)
	}
	if letter.ID == "" {
		id, err := loader.LetterIdentifierGenerator.GenerateID()
		if err != nil {
			return letter, err
		}
		letter.ID = id
	}
	letter, err = extend(ctx, letter, rootDirectory, loader.FileSystem)
	if err != nil {
		return letter, err
	}

	if _, err = newSubject(letter.Frontmatter[FieldNameSubject]); err != nil {
		if errors.Is(err, ErrNoSubject) {
			// pull the subject from the first heading text
			letter.Frontmatter[FieldNameSubject] = markdown.GetFirstHeadingText([]byte(letter.Content))
			if letter.Frontmatter[FieldNameSubject] == "" {
				return letter, err
			}
		} else {
			return letter, err
		}
	}

	templates, err := getTemplates(letter.Frontmatter, rootDirectory)
	if err != nil {
		return letter, err
	}
	for _, t := range templates {
		// if media.IsPathLocal(t) {
		// 	t = path.Join(path.Dir(p), t)
		// }
		file, err := loader.FileSystem.Open(t)
		if err != nil {
			return letter, NewFileReadError(t, err)
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return letter, NewFileReadError(t, errors.Join(err, file.Close()))
		}
		if err = file.Close(); err != nil {
			return letter, NewFileReadError(t, err)
		}
		letter.Templates = append(letter.Templates, Attachment{
			Name:        t,
			Content:     data,
			ContentType: media.ContentTypeTextHTML,
		})
	}
	return letter, nil
}

func (loader loader) LoadAttachment(
	ctx context.Context,
	source AttachmentSource,
	constraints media.Constraints,
) (a Attachment, err error) {
	select {
	case <-ctx.Done():
		return a, ctx.Err()
	default:
	}

	file, err := loader.FileSystem.Open(source.Location)
	if err != nil {
		return a, NewFileReadError(source.Location, err)
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	b, err := io.ReadAll(file)
	if err != nil {
		return a, NewFileReadError(source.Location, err)
	}
	a, err = NewAttachment(b, constraints)
	if err != nil {
		return a, err
	}
	a.Name = source.Name
	return a, err
}

func (loader loader) LoadLetter(ctx context.Context, p string) (letter Letter, attachments []Attachment, err error) {
	rootDirectory := path.Dir(p)
	letter, err = loader.loadLetterFromFile(ctx, p, rootDirectory)
	if err != nil {
		return letter, nil, NewFileReadError(p, err)
	}
	domain, err := letter.GetDomain()
	if err != nil {
		return letter, nil, NewFileReadError(p, err)
	}
	// language, err := letter.GetLanguage()
	// if err != nil {
	// 	return letter, nil, err
	// }
	// if language.String() == "und" {
	// 	letter.Frontmatter[FieldNameLanguage] = loader.DefaultLanguage.String()
	// }
	constraints, err := letter.GetMediaConstraints()
	if err != nil {
		return letter, nil, err
	}
	if constraints.Quality == 0 {
		constraints.Quality = loader.DefaultMediaContraints.Quality
	}
	if constraints.Width == 0 {
		constraints.Width = loader.DefaultMediaContraints.Width
	}
	if constraints.Height == 0 {
		constraints.Height = loader.DefaultMediaContraints.Height
	}

	// load attachments from the front matter
	for source, err := range letter.EachAttachment() {
		if err != nil {
			return letter, nil, fmt.Errorf("unable to decode attachment source %q: %w", source.Location, err)
		}
		if !filepath.IsAbs(source.Location) {
			source.Location = filepath.Join(rootDirectory, source.Location)
		}
		for _, attachment := range attachments {
			if attachment.Source == source.Location {
				return letter, attachments, fmt.Errorf("duplicate attachment: %s", attachment.Source)
			}
		}
		attachment, err := loader.LoadAttachment(
			ctx,
			source,
			constraints,
		)
		if err != nil {
			return letter, nil, NewFileReadError(source.Location, err)
		}
		attachment.LetterID = letter.ID
		attachments = append(attachments, attachment)
	}

	// load inline attachments from the templates directory
	for i, t := range letter.Templates {
		if !filepath.IsAbs(t.Source) {
			t.Source = filepath.Join(rootDirectory, t.Source)
		}
		letter.Templates[i].Content, err = inlineTemplateAttachments(
			t.Content,
			func(source AttachmentSource) (string, error) {
				if !filepath.IsAbs(source.Location) {
					source.Location = filepath.Join(t.Source, source.Location)
				}
				attachment, err := loader.LoadAttachment(
					ctx,
					source,
					constraints,
				)
				if err != nil {
					return "", NewFileReadError(source.Location, err)
				}
				attachment.ContentID = attachment.Hash + "@" + domain
				for _, a := range attachments {
					if a.ContentID == attachment.ContentID {
						return "cid:" + attachment.ContentID, nil
					}
				}
				attachment.LetterID = letter.ID
				attachments = append(attachments, attachment)
				return "cid:" + attachment.ContentID, nil
			},
		)
	}

	// load inline attachments from the Markdown content
	letter.Content, err = inlineMarkdownAttachments(
		loader.Parser,
		[]byte(letter.Content),
		func(source AttachmentSource) (string, error) {
			if !filepath.IsAbs(source.Location) {
				source.Location = filepath.Join(rootDirectory, source.Location)
			}
			attachment, err := loader.LoadAttachment(
				ctx,
				source,
				constraints,
			)
			if err != nil {
				return "", NewFileReadError(source.Location, err)
			}
			attachment.ContentID = attachment.Hash + "@" + domain
			for _, a := range attachments {
				if a.ContentID == attachment.ContentID {
					return "cid:" + attachment.ContentID, nil
				}
			}
			attachment.LetterID = letter.ID
			attachments = append(attachments, attachment)
			return "cid:" + attachment.ContentID, nil
		},
	)

	return letter, attachments, err
}
