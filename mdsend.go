package mdsend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"path"

	"github.com/dkotik/mdsend/internal/html"
	"github.com/dkotik/mdsend/internal/media"
	"github.com/dkotik/mdsend/markdown"
	"github.com/oklog/ulid/v2"
)

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
	LoadLetter(context.Context, string) (Letter, iter.Seq2[Attachment, error], error)
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
	}, nil
}

func (loader loader) loadLetterFromFile(
	ctx context.Context,
	p string,
	rootDirectory string,
) (letter Letter, err error) {
	file, err := loader.FileSystem.Open(p)
	if err != nil {
		return letter, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return letter, errors.Join(err, file.Close())
	}
	if err = file.Close(); err != nil {
		return letter, err
	}

	letter, err = newLetter(data)
	if err != nil {
		return letter, err
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
			if errors.Is(err, fs.ErrNotExist) {
				data, ok := html.LoadEmbeddedTemplate(t)
				if !ok {
					return letter, fmt.Errorf("template does not exist: %s", t)
				}
				letter.Templates = append(letter.Templates, Attachment{
					Name:        t,
					Content:     data,
					ContentType: media.ContentTypeTextHTML,
				})
				continue
			}
			return letter, err
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return letter, errors.Join(err, file.Close())
		}
		if err = file.Close(); err != nil {
			return letter, err
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
		return a, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	b, err := io.ReadAll(file)
	if err != nil {
		return a, err
	}
	a, err = NewAttachment(b, constraints)
	if err != nil {
		return a, err
	}
	a.Name = source.Name
	return a, err
}

func (loader loader) LoadLetter(ctx context.Context, p string) (Letter, iter.Seq2[Attachment, error], error) {
	rootDirectory := path.Dir(p)
	letter, err := loader.loadLetterFromFile(ctx, p, rootDirectory)
	if err != nil {
		return letter, nil, fmt.Errorf("unable to load file %q: %w", p, err)
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
	return letter,
		func(yield func(Attachment, error) bool) {
			for source, err := range letter.EachAttachmentSource() {
				if err != nil {
					yield(Attachment{}, fmt.Errorf("unable to decode attachment source %q: %w", source.Location, err))
					return
				}
				attachment, err := loader.LoadAttachment(
					ctx,
					source,
					constraints,
				)
				if err != nil {
					yield(Attachment{}, fmt.Errorf("unable to load attachment %q: %w", source.Location, err))
					return
				}
				attachment.LetterID = letter.ID
				if !yield(attachment, nil) {
					return
				}
			}
		}, nil
}
