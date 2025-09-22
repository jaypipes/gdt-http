// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"fmt"
	"strings"

	"github.com/gdt-dev/core/api"
	gdtjson "github.com/gdt-dev/core/assertion/json"
	"github.com/gdt-dev/core/parse"
	"github.com/samber/lo"
	"github.com/theory/jsonpath"
	"gopkg.in/yaml.v3"
)

var validHTTPMethods = []string{
	"DELETE",
	"GET",
	"PATCH",
	"POST",
	"PUT",
}

// InvalidHTTPMethodAt returns a parse error indicating the test author used an
// invalid method field value.
func InvalidHTTPMethodAt(method string, node *yaml.Node) error {
	return &parse.Error{
		Line:   node.Line,
		Column: node.Column,
		Message: fmt.Sprintf(
			"invalid HTTP method specified: %s. valid values: %s",
			method, strings.Join(validHTTPMethods, ","),
		),
	}
}

// EitherShortcutOrHTTPSpecAt returns a parse error indicating the test author
// included both a shortcut (e.g. `http.get` or just `GET`) AND the long-form
// `http` object in the same test spec.
func EitherShortcutOrHTTPSpecAt(node *yaml.Node) error {
	return &parse.Error{
		Line:   node.Line,
		Column: node.Column,
		Message: "either specify a full HTTPSpec in the `http` field or " +
			"specify one of the shortcuts (e.g. `http.get` or `GET`",
	}
}

// MultipleHTTPMethods returns a parse error indicating the test author
// specified multiple HTTP methods either full-form or via shortcuts.
func MultipleHTTPMethods(
	firstMethod string,
	secondMethod string,
	node *yaml.Node,
) error {
	return &parse.Error{
		Line:   node.Line,
		Column: node.Column,
		Message: fmt.Sprintf(
			"multiple HTTP methods specified (%q, %q). "+
				"please specify a single HTTP method for each test spec.",
			firstMethod, secondMethod,
		),
	}
}

func (s *Spec) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	vars := Variables{}
	// We do an initial pass over the shortcut fields, then all the
	// non-shortcut fields after that.
	var hs *HTTPSpec

	// maps/structs are stored in a top-level Node.Content field which is a
	// concatenated slice of Node pointers in pairs of key/values.
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "url":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			hs = &HTTPSpec{}
			hs.URL = url
		case "method":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			method := strings.ToUpper(strings.TrimSpace(valNode.Value))
			if !lo.Contains(validHTTPMethods, s.Method) {
				return InvalidHTTPMethodAt(valNode.Value, valNode)
			}
			if hs != nil {
				if hs.Method != "" {
					return MultipleHTTPMethods(hs.Method, method, valNode)
				}
				hs.Method = method
			} else {
				hs = &HTTPSpec{}
				hs.Method = method
			}
		case "GET", "get", "http.get":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if hs != nil {
				return MultipleHTTPMethods(hs.Method, "GET", valNode)
			}
			hs = &HTTPSpec{}
			hs.Method = "GET"
			hs.URL = url
		case "POST", "post", "http.post":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if hs != nil {
				return MultipleHTTPMethods(hs.Method, "POST", valNode)
			}
			hs = &HTTPSpec{}
			hs.Method = "POST"
			hs.URL = url
		case "PUT", "put", "http.put":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if hs != nil {
				return MultipleHTTPMethods(hs.Method, "PUT", valNode)
			}
			hs = &HTTPSpec{}
			hs.Method = "PUT"
			hs.URL = url
		case "DELETE", "delete", "http.delete":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if hs != nil {
				return MultipleHTTPMethods(hs.Method, "DELETE", valNode)
			}
			hs = &HTTPSpec{}
			hs.Method = "DELETE"
			hs.URL = url
		case "PATCH", "patch", "http.patch":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if hs != nil {
				return MultipleHTTPMethods(hs.Method, "PATCH", valNode)
			}
			hs = &HTTPSpec{}
			hs.Method = "PATCH"
			hs.URL = url
		case "data":
			var data interface{}
			if err := valNode.Decode(&data); err != nil {
				return err
			}
			s.Data = data
		}
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "http":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			if hs != nil {
				return EitherShortcutOrHTTPSpecAt(valNode)
			}
			if err := valNode.Decode(&hs); err != nil {
				return err
			}
			s.HTTP = hs
		case "var":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			var specVars Variables
			if err := valNode.Decode(&specVars); err != nil {
				return err
			}
			vars = lo.Assign(specVars, vars)
		case "assert":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			var e *Expect
			if err := valNode.Decode(&e); err != nil {
				return err
			}
			s.Assert = e
		case "http.get", "http.post", "http.delete", "http.put", "http.patch",
			"GET", "POST", "DELETE", "PUT", "PATCH",
			"get", "post", "delete", "put", "patch",
			"url", "method", "data":
			continue
		default:
			if lo.Contains(api.BaseSpecFields, key) {
				continue
			}
			return parse.UnknownFieldAt(key, keyNode)
		}
	}
	if s.Data != nil {
		hs.Data = s.Data
	}
	s.HTTP = hs
	if len(vars) > 0 {
		s.Var = vars
	}
	return nil
}

func (s *HTTPSpec) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	// maps/structs are stored in a top-level Node.Content field which is a
	// concatenated slice of Node pointers in pairs of key/values.
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		switch key {
		case "get", "put", "post", "patch", "delete",
			"GET", "PUT", "POST", "PATCH", "DELETE",
			"url", "method", "data":
			// Because Action is an embedded struct and we parse it below, just
			// ignore these fields in the top-level `http:` field for now.
		default:
			return parse.UnknownFieldAt(key, keyNode)
		}
	}
	var a Action
	if err := node.Decode(&a); err != nil {
		return err
	}
	s.Action = a
	return nil
}

func (a *Action) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	// maps/structs are stored in a top-level Node.Content field which is a
	// concatenated slice of Node pointers in pairs of key/values.
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "url":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			a.URL = url
		case "method":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			method := strings.ToUpper(strings.TrimSpace(valNode.Value))
			if !lo.Contains(validHTTPMethods, a.Method) {
				return InvalidHTTPMethodAt(valNode.Value, valNode)
			}
			if a.Method != "" {
				return MultipleHTTPMethods(a.Method, method, valNode)
			}
			a.Method = method
		case "GET", "get", "http.get":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if a.Method != "" {
				return MultipleHTTPMethods(a.Method, "GET", valNode)
			}
			a.Method = "GET"
			a.URL = url
		case "POST", "post", "http.post":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if a.Method != "" {
				return MultipleHTTPMethods(a.Method, "POST", valNode)
			}
			a.Method = "POST"
			a.URL = url
		case "PUT", "put", "http.put":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if a.Method != "" {
				return MultipleHTTPMethods(a.Method, "PUT", valNode)
			}
			a.Method = "PUT"
			a.URL = url
		case "DELETE", "delete", "http.delete":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if a.Method != "" {
				return MultipleHTTPMethods(a.Method, "DELETE", valNode)
			}
			a.Method = "DELETE"
			a.URL = url
		case "PATCH", "patch", "http.patch":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			url := strings.TrimSpace(valNode.Value)
			if a.Method != "" {
				return MultipleHTTPMethods(a.Method, "PATCH", valNode)
			}
			a.Method = "PATCH"
			a.URL = url
		case "data":
			var data interface{}
			if err := valNode.Decode(&data); err != nil {
				return err
			}
			a.Data = data
		}
	}
	return nil
}

// UnmarshalYAML is a custom unmarshaler that ensures that JSONPath expressions
// contained in the VarEntry are valid.
func (e *VarEntry) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	// maps/structs are stored in a top-level Node.Content field which is a
	// concatenated slice of Node pointers in pairs of key/values.
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "from":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			var path string
			if err := valNode.Decode(&path); err != nil {
				return err
			}
			if len(path) == 0 || path[0] != '$' {
				return gdtjson.JSONPathInvalidNoRoot(path, valNode)
			}
			if _, err := jsonpath.Parse(path); err != nil {
				return gdtjson.JSONPathInvalid(path, err, valNode)
			}
			e.From = path
		}
	}
	return nil
}
