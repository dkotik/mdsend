package mdsend

import (
	"context"
	"errors"
	"io"
	"iter"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/dkotik/mdsend/internal/locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type ValidationResult struct {
	LetterPath string
	Subject    *i18n.LocalizeConfig
	Errors     []*i18n.LocalizeConfig
	Warnings   []*i18n.LocalizeConfig
}

func (r ValidationResult) IsMeaningful() bool {
	return (len(r.Errors) + len(r.Warnings)) > 0
}

func (r ValidationResult) withNewSubject(subject *i18n.LocalizeConfig) ValidationResult {
	r.Subject = subject
	r.Errors = r.Errors[:0]
	r.Warnings = r.Warnings[:0]
	return r
}

func (r ValidationResult) withError(err error) ValidationResult {
	// translate if able, otherwise:
	r.Errors = append(r.Errors, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			Other: "Operation failed: {{ .Error }}.",
		},
		TemplateData: map[string]any{
			"Error": err.Error(),
		},
	})
	return r
}

func (r ValidationResult) withWarning(warning *i18n.LocalizeConfig) ValidationResult {
	r.Warnings = append(r.Warnings, warning)
	return r
}

func (l Letter) Validate() (err error) {
	if strings.TrimSpace(l.Content) == "" {
		return ErrNoContent
	}
	if _, err = l.GetSubject(); err != nil {
		return err
	}
	if _, err = l.GetFrom(); err != nil {
		return err
	}
	if _, err = l.GetLanguage(); err != nil {
		return err
	}
	if _, err = l.GetSchedule(); err != nil {
		return err
	}
	return nil
}

func (loader loader) Validate(ctx context.Context, p string) iter.Seq[ValidationResult] {
	return func(yield func(ValidationResult) bool) {
		v := ValidationResult{
			LetterPath: p,
			Subject: &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					Other: "Letter contents",
				},
			},
			Errors:   make([]*i18n.LocalizeConfig, 0, 4),
			Warnings: make([]*i18n.LocalizeConfig, 0, 4),
		}
		file, err := loader.FileSystem.Open(p)
		if err != nil {
			yield(v.withError(NewFileReadError(p, err)))
			return
		}
		data, err := io.ReadAll(file)
		if err != nil {
			yield(v.withError(NewFileReadError(p, errors.Join(err, file.Close()))))
			return
		}
		if len(data) == 0 {
			yield(v.withError(NewFileReadError(p, errors.New("file is empty"))))
			return
		}
		if err = file.Close(); err != nil {
			yield(v.withError(NewFileReadError(p, err)))
			return
		}
		letter, err := newLetter(data)
		if err != nil {
			yield(v.withError(err))
			return
		}
		rootDirectory := filepath.Base(p)
		letter, err = extend(ctx, letter, rootDirectory, loader.FileSystem)
		if err != nil {
			v = v.withError(err)
		}
		if v.IsMeaningful() {
			yield(v)
		}

		v = v.withNewSubject(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				Other: "Letter frontmatter",
			},
		})
		language, err := letter.GetLanguage()
		if err != nil {
			v = v.withError(err)
		}
		if !locale.IsValidLanguageTag(language) {
			v = v.withWarning(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					Other: "Letter does not contain a valid \"language\" frontmatter field. Example: language: \"en\".",
				},
			})
		}

		subject, err := letter.GetSubject()
		if err != nil {
			v = v.withError(err)
		}
		if len(subject) < 5 {
			v = v.withWarning(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					Other: "Letter subject is too short.",
				},
			})
		} else if len(subject) > 80 {
			v = v.withWarning(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					Other: "Letter subject is too long.",
				},
			})
		}
		if v.IsMeaningful() {
			yield(v)
		}

		v = v.withNewSubject(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				Other: "Letter templates",
			},
		})
		templates, err := getTemplates(letter.Frontmatter, rootDirectory)
		if err != nil {
			v = v.withError(err)
		} else {
			for _, t := range templates {
				file, err := loader.FileSystem.Open(t)
				if err != nil {
					v = v.withError(NewFileReadError(t, err))
					continue
				}
				if _, err = io.ReadAll(file); err != nil {
					v = v.withError(NewFileReadError(p, err))
					continue
				}
				if err = file.Close(); err != nil {
					v = v.withError(NewFileReadError(p, err))
					continue
				}
			}
		}
		if v.IsMeaningful() {
			yield(v)
		}

		v = v.withNewSubject(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				Other: "Letter headers",
			},
		})
		known := make(map[string]int)
		headers, err := letter.GetHeaders()
		if err != nil {
			v = v.withError(err)
		} else {
			for _, h := range headers {
				known[h.Name] = known[h.Name] + 1
			}
			for _, count := range known {
				if count < 2 {
					continue
				}
				v = v.withWarning(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						Other: "There are {{.}} canonical header duplicates.",
					},
					TemplateData: count - 1,
					PluralCount:  count - 1,
				})
			}
		}
		if len(known) == 0 {
			v = v.withWarning(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					Other: "There are no letter headers set in the frontmatter. Headers <List-Id> and <List-Unsubscribe> should be set.",
				},
			})
		} else {
			ok := false
			if _, ok = known["List-Id"]; !ok {
				v = v.withWarning(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						Other: "Letter headers in missing a <List-Id>.",
					},
				})
			}
			if _, ok = known["List-Unsubscribe"]; !ok {
				v = v.withWarning(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						Other: "Letter headers in missing a <List-Unsubscribe>.",
					},
				})
			} else {
				for _, header := range headers {
					if header.Name == "List-Unsubscribe" {
						options := strings.Split(header.Value, ",")
						for _, header := range options {
							header = strings.TrimSpace(header)
							if header == "" {
								v = v.withWarning(&i18n.LocalizeConfig{
									DefaultMessage: &i18n.Message{
										Other: "<List-Unsubscribe> header is empty.",
									},
								})
								continue
							}
							if len(header) < 4 {
								v = v.withWarning(&i18n.LocalizeConfig{
									DefaultMessage: &i18n.Message{
										Other: "<List-Unsubscribe> header is not invalid.",
									},
								})
								continue
							}
							if header[0] != '<' {
								v = v.withWarning(&i18n.LocalizeConfig{
									DefaultMessage: &i18n.Message{
										Other: "<List-Unsubscribe> option does not start with a '<'.",
									},
								})
							}
							if header[len(header)-1] != '>' {
								v = v.withWarning(&i18n.LocalizeConfig{
									DefaultMessage: &i18n.Message{
										Other: "<List-Unsubscribe> option does not end with a '>'.",
									},
								})
							}
						}
					}
				}
			}
		}
		if v.IsMeaningful() {
			yield(v)
		}
	}
}

func (l Letter) isValid(cxt context.Context, logger *slog.Logger) (ok bool) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	var err error
	if strings.TrimSpace(l.Content) == "" {
		logger.ErrorContext(cxt, "letter has no content")
		ok = false
	}
	if _, err = l.GetSubject(); err != nil {
		logger.ErrorContext(cxt, "defective subject:", slog.Any("error", err))
		ok = false
	}
	if _, err = l.GetFrom(); err != nil {
		logger.ErrorContext(cxt, "invalid sender address:", slog.Any("error", err))
		ok = false
	}
	if _, err = l.GetSchedule(); err != nil {
		logger.ErrorContext(cxt, "invalid schedule:", slog.Any("error", err))
		ok = false
	}
	if ok {
		if err = l.Validate(); err != nil {
			logger.ErrorContext(cxt, "letter validation failed:", slog.Any("error", err))
			return false
		}
	}
	return ok
}
