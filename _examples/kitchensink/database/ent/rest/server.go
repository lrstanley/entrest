// Code generated by ent, DO NOT EDIT.

package rest

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"

	"github.com/lrstanley/entrest/_examples/kitchensink/database/ent"
)

//go:embed openapi.json
var OpenAPI []byte // OpenAPI contains the JSON schema of the API.

// UseEntContext is an http middleware that injects the provided ent.Client into the
// request context.
func UseEntContext(client *ent.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ent.NewContext(r.Context(), client)))
		})
	}
}

type ErrBadRequest struct {
	Err error
}

func (e ErrBadRequest) Error() string {
	return fmt.Sprintf("bad request: %s", e.Err)
}

func (e ErrBadRequest) Unwrap() error {
	return e.Err
}

// IsBadRequest returns true if the unwrapped/underlying error is of type ErrBadRequest.
func IsBadRequest(err error) bool {
	var target *ErrBadRequest
	return errors.As(err, &target)
}