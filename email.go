package mdsend

import (
	"errors"
	"net"
	"net/mail"
	"regexp"
	"strings"
)

/*
<https://ayada.dev/posts/validate-email-address-in-go/>

W3C has provided [recommendation](https://html.spec.whatwg.org/multipage/input.html#email-state-(type=email)) on the valid email address syntax. It also provides a regular expression that can be used to validate the syntax of the email address:
*/
var (
	reValidEmailAddressW3C = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	ErrEmailAddressTooLong       = errors.New("email address too long")
	ErrEmailAddressInvalid       = errors.New("email address is in the work format")
	ErrEmailAddressDomainInvalid = errors.New("unable to validate email address domain")
)

const MaximumEmailAddressLength = 254 // RFC 3696 Errata 1690

func ValidateEmail(emailAddress string) error {
	if len(emailAddress) > MaximumEmailAddressLength {
		return ErrEmailAddressTooLong
	}

	if !reValidEmailAddressW3C.MatchString(emailAddress) {
		return ErrEmailAddressInvalid
	}

	domain := strings.Split(emailAddress, "@")[1]
	if mx, err := net.LookupMX(domain); err != nil || len(mx) == 0 {
		return errors.Join(ErrEmailAddressDomainInvalid, err)
	}
	return nil
}

func newAddressFromMap(m map[string]any) (result mail.Address, err error) {
	switch nameRaw := m[FieldNameName].(type) {
	case nil:
	case string:
		result.Name = strings.TrimSpace(nameRaw)
	default:
		return result, errors.New("invalid name format")
	}

	switch emailRaw := m[FieldNameName].(type) {
	case nil:
		return result, errors.New("no electronic email address specified")
	case string:
		result.Address = strings.TrimSpace(emailRaw)
		if err = ValidateEmail(result.Address); err != nil {
			return result, err
		}
	default:
		return result, errors.New("invalid email format")
	}

	return result, nil
}
