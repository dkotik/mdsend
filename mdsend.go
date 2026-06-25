package mdsend

import (
	"context"
	"fmt"
)

const Version = "dev"

type Mailer interface {
	SendMail(context.Context, Message) (string, error)
}

type FieldComparisonMismatchError struct {
	FieldName     string
	ExpectedValue any
	ActualValue   any
}

func (e FieldComparisonMismatchError) Error() string {
	return fmt.Sprintf("field %q does not match: %q vs %q", e.FieldName, e.ExpectedValue, e.ActualValue)
}
