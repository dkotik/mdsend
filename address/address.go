package address

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

const (
	// RFC 3696 Errata 1690
	MaximumLength = 254

	FieldName            = "name"
	FieldEmail           = "email"
	FieldTo              = "to"
	FieldCarbonCopy      = "cc"
	FieldBlindCarbonCopy = "bcc"
)

/*
<https://ayada.dev/posts/validate-email-address-in-go/>

W3C has provided [recommendation](https://html.spec.whatwg.org/multipage/input.html#email-state-(type=email)) on the valid email address syntax. It also provides a regular expression that can be used to validate the syntax of the email address:
*/
var (
	reValidEmailAddressW3C = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	ErrAbsentEmailAddress  = errors.New("absent email address")
	ErrEmailAddressTooLong = errors.New("email address too long")
	ErrEmailAddressInvalid = errors.New("email address is in the wrong format")
	// ErrEmailAddressDomainInvalid = errors.New("unable to validate email address domain")
)

func New(m map[string]any) (result mail.Address, err error) {
	switch nameRaw := m[FieldName].(type) {
	case nil:
	case string:
		result.Name = strings.TrimSpace(nameRaw)
	default:
		return result, errors.New("invalid name format")
	}

	switch emailRaw := m[FieldEmail].(type) {
	case nil:
		return result, errors.New("no electronic email address specified")
	case string:
		result.Address = strings.TrimSpace(emailRaw)
		if err = ValidateFormat(result.Address); err != nil {
			return result, err
		}
	default:
		return result, errors.New("invalid email format")
	}

	return result, nil
}

func ValidateFormat(emailAddress string) error {
	if len(emailAddress) > MaximumLength {
		return ErrEmailAddressTooLong
	}

	if !reValidEmailAddressW3C.MatchString(emailAddress) {
		return ErrEmailAddressInvalid
	}
	return nil
}
