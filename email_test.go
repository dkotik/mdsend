package mdsend

import (
	"errors"
	"testing"
)

func TestEmailAddressValidation(t *testing.T) {
	// if t.Short() {
	// 	t.Skip("MX domain look ups take a long time")
	// }
	cases := [...]struct {
		Address string
		Error   error
	}{
		{"gjfgifg@gmail.com", nil},
	}

	var err error
	for i, tc := range cases {
		if err = ValidateEmail(tc.Address); !errors.Is(err, tc.Error) {
			t.Errorf("email #%d failed: %s", i, err)
		}
	}
}
