// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	nethttp "net/http"
	"strings"
	"testing"

	gdtjson "github.com/jaypipes/gdt-core/assertion/json"
	"github.com/stretchr/testify/assert"
)

const (
	msgHTTPStatus   = "Expected HTTP response to have status code of %d but got %d"
	msgStringInBody = "Expected HTTP response to contain %s"
	msgHeaderIn     = "Expected HTTP header %s to be in response"
	msgHeaderValue  = "Expected HTTP header with value %s to be in response"
)

// ResponseAssertions contains one or more assertions about an HTTP response
type ResponseAssertions struct {
	// JSON contains the assertions about JSON data in the response
	JSON *gdtjson.Expect `yaml:"json,omitempty"`
	// Headers contains a list of HTTP headers that should be in the response
	Headers []string `yaml:"headers,omitempty"`
	// Strings contains a list of strings that should be present in the
	// response content
	Strings []string `yaml:"strings,omitempty"`
	// Status contains the numeric HTTP status code (e.g. 200 or 404) that
	// should be returned in the HTTP response
	Status *int `yaml:"status,omitempty"`
}

func assertHTTPStatusEqual(t *testing.T, r *nethttp.Response, exp int) {
	t.Helper()
	got := r.StatusCode
	assert.Equal(t, exp, got, msgHTTPStatus, exp, got)
}

func assertStringInBody(t *testing.T, r *nethttp.Response, b []byte, exp string) {
	t.Helper()
	assert.Contains(t, string(b), exp, msgStringInBody, exp)
}

func assertHeader(t *testing.T, r *nethttp.Response, exp string) {
	t.Helper()
	colonPos := strings.IndexRune(exp, ':')
	if colonPos > -1 {
		keyPart := exp[:colonPos]
		valPart := exp[colonPos+1:]
		val := r.Header.Get(keyPart)
		assert.NotEmpty(t, val, msgHeaderIn, exp)
		// If the string being compared is of the form Key: Value,
		// then we check for both existence and the value of the
		// header
		expVal := strings.ToLower(valPart)
		assert.Equal(t, expVal, strings.ToLower(val), msgHeaderValue, exp)
	} else {
		val := r.Header.Get(exp)
		assert.NotEmpty(t, val, msgHeaderIn, exp)
	}
}
