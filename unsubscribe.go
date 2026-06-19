package mdsend

import "net/mail"

type Unsubscribe struct {
	mail.Address
	URL string
}

func (l Letter) GetListID() (listID string, err error) {
	listID, _ = l.Frontmatter[FieldNameListID].(string)
	return listID, nil
}

func (l Letter) GetUnsubscribe() (u Unsubscribe, err error) {
	m, ok := l.Frontmatter[FieldNameUnsubscribe].(map[string]any)
	if !ok {
		return Unsubscribe{}, nil
	}
	if u.Address, err = newAddressFromMap(m[FieldNameUnsubscribeEmail].(map[string]any)); err != nil {
		return Unsubscribe{}, err
	}
	if url, ok := m[FieldNameUnsubscribeURL].(string); ok {
		u.URL = url
	}
	return
}
