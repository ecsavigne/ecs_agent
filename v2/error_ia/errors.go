//lint:file-ignore ST1005 allowed Capital letter to start
package error_ia

import "errors"

type ErrorIA interface {
	Error() string
}

var (
	ErrorApiKeyNotSet      = errors.New("API key not set")
	ErrorAgentNotFound     = errors.New("Agent not found")
	ErrorQuotaExceeded     = errors.New("API quota exceeded")
	ErrorGettingCompletion = errors.New("Error getting completion")
)
