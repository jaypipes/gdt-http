// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	gdttypes "github.com/jaypipes/gdt-core/types"
)

// Spec describes a test of a single HTTP request and response
type Spec struct {
	gdttypes.Spec
	// URL being called by HTTP client
	URL string `yaml:"url,omitempty"`
	// HTTP Method specified by HTTP client
	Method string `yaml:"method,omitempty"`
	// Shortcut for URL and Method of "GET"
	GET string `yaml:"GET,omitempty"`
	// Shortcut for URL and Method of "POST"
	POST string `yaml:"POST,omitempty"`
	// Shortcut for URL and Method of "PUT"
	PUT string `yaml:"PUT,omitempty"`
	// Shortcut for URL and Method of "PATCH"
	PATCH string `yaml:"PATCH,omitempty"`
	// Shortcut for URL and Method of "DELETE"
	DELETE string `yaml:"DELETE,omitempty"`
	// JSON payload to send along in request
	Data interface{} `yaml:"data,omitempty"`
	// Specification for expected response
	Response *ResponseAssertions `yaml:"response,omitempty"`
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

func (s *Spec) SetBase(b gdttypes.Spec) {
	s.Spec = b
}

func (s *Spec) Base() *gdttypes.Spec {
	return &s.Spec
}
