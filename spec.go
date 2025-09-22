// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"github.com/gdt-dev/core/api"
)

// HTTPSpec is the complex type containing all of the HTTP-specific actions.
// Most users will use the `GET`, `http.get` and similar shortcut fields.
type HTTPSpec struct {
	Action
}

// Spec describes a test of a single HTTP request and response
type Spec struct {
	api.Spec
	// HTTP is the complex type containing all of the HTTP-specific actions and
	// assertions. Most users will use the `URL`/`Method`, `GET`, `http.get`
	// and similar shortcut fields.
	HTTP *HTTPSpec `yaml:"http,omitempty"`
	// Shortcut for `http.url`
	URL string `yaml:"url,omitempty"`
	// Shortcut for `http.method`
	Method string `yaml:"method,omitempty"`
	// Shortcut for `http.get`
	GET string `yaml:"GET,omitempty"`
	// Shortcut for `http.post`
	POST string `yaml:"POST,omitempty"`
	// Shortcut for `http.put`
	PUT string `yaml:"PUT,omitempty"`
	// Shortcut for `http.patch`
	PATCH string `yaml:"PATCH,omitempty"`
	// Shortcut for `http.delete`
	DELETE string `yaml:"DELETE,omitempty"`
	// Shortcut for `http.data`
	Data any `yaml:"data,omitempty"`
	// Assert is the assertions for the HTTP response
	Assert *Expect `yaml:"assert,omitempty"`
	// Var allows the test author to save arbitrary data to the test scenario,
	// facilitating the passing of variables between test specs potentially
	// provided by different gdt Plugins.
	Var Variables `yaml:"var,omitempty"`
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

func (s *Spec) Retry() *api.Retry {
	if s.Spec.Retry != nil {
		// The user may have overridden in the test spec file...
		return s.Spec.Retry
	}
	if s.Method == "GET" {
		// returning nil here means the plugin's default will be used...
		return nil
	}
	// for POST/PUT/DELETE/PATCH, we don't want to retry...
	return api.NoRetry
}

func (s *Spec) Timeout() *api.Timeout {
	// returning nil here means the plugin's default will be used...
	return nil
}

func (s *Spec) SetBase(b api.Spec) {
	s.Spec = b
}

func (s *Spec) Base() *api.Spec {
	return &s.Spec
}
