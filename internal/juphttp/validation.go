package juphttp

import (
	"errors"
	"regexp"
)

var base58ish = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,64}$`)

// ValidatePublicKey performs basic client-side validation for Solana public keys.
func ValidatePublicKey(value string) error {
	if !base58ish.MatchString(value) {
		return errors.New("invalid solana public key")
	}
	return nil
}

// ValidateRawAmount requires raw integer token units, not UI decimal amounts.
func ValidateRawAmount(amount string) error {
	if amount == "" {
		return errors.New("amount is required")
	}
	for _, r := range amount {
		if r < '0' || r > '9' {
			return errors.New("amount must be raw integer token units")
		}
	}
	return nil
}
