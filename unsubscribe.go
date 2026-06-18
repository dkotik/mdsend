package mdsend

import "net/mail"

type Unsubscribe struct {
	mail.Address
	URL string
}

func (l Letter) GetUnsubscribe() (u Unsubscribe, err error) {
	// switch from := l.Frontmatter[FieldNameFrom].(type) {
	// case map[string]any:
	// 	return newAddressFromMap(from)
	// case string:
	// 	if strings.TrimSpace(from) == "" {
	// 		return mail.Address{}, ErrNoFromAddress
	// 	}
	// 	address, err := mail.ParseAddress(from)
	// 	if err != nil {
	// 		return mail.Address{}, err
	// 	}
	// 	return *address, nil
	// default:
	// 	return mail.Unsubscribe{}, ErrNoFromAddress
	// }
	return
}
