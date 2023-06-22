// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"github.com/jaypipes/gdt-core/spec"
)

// Spec describes a test of a single HTTP request and response
type Spec struct {
	spec.Spec
	defaults *Defaults
	// URL being called by HTTP client
	URL string `json:"url,omitempty"`
	// HTTP Method specified by HTTP client
	Method string `json:"method,omitempty"`
	// Shortcut for URL and Method of "GET"
	GET string `json:"GET,omitempty"`
	// Shortcut for URL and Method of "POST"
	POST string `json:"POST,omitempty"`
	// Shortcut for URL and Method of "PUT"
	PUT string `json:"PUT,omitempty"`
	// Shortcut for URL and Method of "PATCH"
	PATCH string `json:"PATCH,omitempty"`
	// Shortcut for URL and Method of "DELETE"
	DELETE string `json:"DELETE,omitempty"`
	// JSON payload to send along in request
	Data interface{} `json:"data,omitempty"`
	// Specification for expected response
	Response *ResponseAssertions `json:"response,omitempty"`
}

// Title returns a good name for the Spec
func (s *Spec) Title() string {
	// If the user did not specify a name for the test spec, just default
	// it to the method and URL
	if s.Name != "" {
		return s.Name
	}
	return s.Method + ":" + s.URL
}
