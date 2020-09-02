package httpdeco

import (
	"net/http"
)

// Decorator decorates http.Handlers.
type Decorator func(http.Handler) http.Handler

// Decorate applies a bunch of decorators to an http.Handler.
func Decorate(h http.Handler, dd ...Decorator) http.Handler {
	result := h

	for _, d := range dd {
		result = d(result)
	}

	return result
}
