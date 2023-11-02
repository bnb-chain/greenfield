package types

import (
	"net/url"

	"cosmossdk.io/errors"
)

// Verify if input endpoint URL is valid.
func ValidateEndpointURL(endpointURL string) error {
	if endpointURL == "" {
		return errors.Wrap(ErrInvalidEndpointURL, "Endpoint url cannot be empty.")
	}
	url, err := url.Parse(endpointURL)
	if err != nil {
		return errors.Wrap(ErrInvalidEndpointURL, "Endpoint url cannot be parsed.")
	}
	if url.Path != "/" && url.Path != "" {
		return errors.Wrap(ErrInvalidEndpointURL, "Endpoint url cannot have fully qualified paths.")
	}
	return nil
}
