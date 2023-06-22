// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"fmt"

	gdterrors "github.com/jaypipes/gdt-core/errors"
)

var (
	// ErrInvalidAliasOrURL is returned when the test author failed to provide
	// either a URL and Method or specify one of the aliases like GET, POST, or
	// DELETE
	ErrInvalidAliasOrURL = fmt.Errorf(
		"%w: either specify a URL and Method or specify one "+
			"of GET, POST, PUT, PATCH or DELETE",
		gdterrors.ErrInvalid,
	)
	// ErrExpectedLocationHeader indicates that the user specified the special
	// `$LOCATION` string in the `url` or `GET` fields of the HTTP test spec
	// but there have been no previous HTTP responses to find a Location HTTP
	// Header within.
	ErrExpectedLocationHeader = fmt.Errorf(
		"%w: expected Location HTTP Header in previous response",
		gdterrors.ErrRuntime,
	)
	// ErrJSONSchemaFileNotFound indicates a specified JSONSchema file could
	// not be found.
	ErrJSONSchemaFileNotFound = fmt.Errorf(
		"%w: unable to find JSONSchema file",
		gdterrors.ErrInvalid,
	)
	// ErrUnsupportedJSONSchemaReference indicates that a specified JSONSchema
	// file is referenced as an HTTP(S) URL instead of a file URI.
	ErrUnsupportedJSONSchemaReference = fmt.Errorf(
		"%w: unsupported JSONSchema reference",
		gdterrors.ErrInvalid,
	)
)

// UnsupportedJSONSchemaReference returns ErrUnsupportedJSONSchemaReference for
// a supplied URL.
func UnsupportedJSONSchemaReference(url string) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedJSONSchemaReference, url)
}

// JSONSchemaFileNotFound returns ErrJSONSchemaFileNotFound for a supplied
// path.
func JSONSchemaFileNotFound(path string) error {
	return fmt.Errorf("%w: %s", ErrJSONSchemaFileNotFound, path)
}
