package mdsend

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

func (l Letter) Validate() (err error) {
	// if l.ID == "" {
	// 	return errors.New("letter has no ID")
	// }
	if strings.TrimSpace(l.Content) == "" {
		return ErrNoContent
	}
	if _, err = l.GetSubject(); err != nil {
		return err
	}
	if _, err = l.GetFrom(); err != nil {
		return err
	}
	if _, err = l.GetSchedule(); err != nil {
		return err
	}
	return nil
}

func (l Letter) IsValid(cxt context.Context, logger *slog.Logger) (ok bool) {
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
	if ok {
		if err = l.Validate(); err != nil {
			logger.ErrorContext(cxt, "letter validation failed:", slog.Any("error", err))
			return false
		}
	}
	return ok
}
