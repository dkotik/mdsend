package mdsend

import (
	"errors"
	"testing"

	"github.com/dkotik/mdsend/internal"
)

func TestEmailAddressValidation(t *testing.T) {
	// if t.Short() {
	// 	t.Skip("MX domain look ups take a long time")
	// }
	for _, address := range internal.MockAddresses {
		t.Run(address.Address, func(t *testing.T) {
			var err error
			if err = ValidateEmailFormat(address.Address); err != nil {
				t.Errorf("email failed: %s", err)
			}
		})
	}

	// TODO: enrich validation with meaningful errors
	casesWithError := [...]struct {
		Address string
		Error   error
	}{
		{"gjfgifg@gmail.com", nil},
	}

	var err error
	for i, tc := range casesWithError {
		if err = ValidateEmailFormat(tc.Address); !errors.Is(err, tc.Error) {
			t.Errorf("email #%d failed: %s", i, err)
		}
	}
}
