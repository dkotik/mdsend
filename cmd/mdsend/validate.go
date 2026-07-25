package main

import (
	"context"
	"errors"
	"fmt"
	"iter"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/media"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v3"
	"golang.org/x/text/language"
)

type validator interface {
	Validate(context.Context, string) iter.Seq[mdsend.ValidationResult]
}

func cmdValidate(ctx context.Context, c *cli.Command) error {
	if !c.Args().Present() {
		return errors.New(`no Markdown letters selected to validate`)
	}
	fs := media.NewUnsafeUnconstrainedFileSystem()
	fs = media.NewCyclicalImportPreventingFileSystem(fs)
	loader, err := mdsend.New(fs, mdsend.Defaults{})
	if err != nil {
		return err
	}

	validator := loader.(validator)
	bundle := i18n.NewBundle(language.English)
	localizer := i18n.NewLocalizer(bundle, "en")
	for _, p := range c.Args().Slice() {
		for report := range validator.Validate(ctx, p) {
			if report.LetterPath != "" {
				fmt.Printf("Letter <%s>:\n", report.LetterPath)
			}
			fmt.Printf("  %s:\n", localizer.MustLocalize(report.Subject))
			for _, err := range report.Errors {
				fmt.Printf("    Error: %s\n", localizer.MustLocalize(err))
			}
			for _, err := range report.Warnings {
				fmt.Printf("    %s\n", localizer.MustLocalize(err))
			}
		}
	}

	// conn, err := sqlite.OpenConn(
	// 	":memory:?foreign_keys=true",
	// 	sqlite.OpenCreate, sqlite.OpenReadWrite,
	// )
	// if err != nil {
	// 	return err
	// }
	// defer func() {
	// 	err = errors.Join(err, conn.Close())
	// }()
	// queue, err := sqliteQ.New(conn, "")
	// if err != nil {
	// 	return err
	// }
	// queue, tx, err := queue.BeginTransaction(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer tx.Close(&err)

	if c.IsSet(flagQueue.Name) {
		// validate letters and messages in the queue
		return nil
	}
	return nil
}
