package lncfg

import "github.com/pkt-cash/pktd/btcutil/er"

// Validator is a generic interface for validating sub configurations.
type Validator interface {
	// Validate returns an error if a particular configuration is invalid or
	// insane.
	Validate() er.R
}

// Validate accepts a variadic list of Validators and checks that each one
// passes its Validate method. An error is returned from the first Validator
// that fails.
func Validate(validators ...Validator) er.R {
	for _, validator := range validators {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}
