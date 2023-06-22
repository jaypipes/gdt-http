// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jaypipes/gdt-core/errors"
	"github.com/jaypipes/gdt-core/spec"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func (s *Spec) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.ExpectedMapAt(node)
	}
	// maps/structs are stored in a top-level Node.Content field which is a
	// concatenated slice of Node pointers in pairs of key/values.
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return errors.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "url":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.URL = strings.TrimSpace(valNode.Value)
		case "method":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.Method = strings.TrimSpace(valNode.Value)
		case "GET":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.GET = strings.TrimSpace(valNode.Value)
		case "POST":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.POST = strings.TrimSpace(valNode.Value)
		case "PUT":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.PUT = strings.TrimSpace(valNode.Value)
		case "DELETE":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.DELETE = strings.TrimSpace(valNode.Value)
		case "PATCH":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.PATCH = strings.TrimSpace(valNode.Value)
		case "data":
			var data interface{}
			if err := valNode.Decode(&data); err != nil {
				return err
			}
			s.Data = data
		case "response":
			if valNode.Kind != yaml.MappingNode {
				return errors.ExpectedMapAt(valNode)
			}
			var ra *ResponseAssertions
			if err := valNode.Decode(&ra); err != nil {
				return err
			}
			s.Response = ra
		default:
			if lo.Contains(spec.BaseFields, key) {
				continue
			}
			return errors.UnknownFieldAt(key, keyNode)
		}
	}
	if err := validateResponseAssertions(s.Response); err != nil {
		return err
	}
	if err := validateMethodAndURL(s); err != nil {
		return err
	}
	return nil
}

func validateMethodAndURL(s *Spec) error {
	if s.URL == "" {
		if s.GET != "" {
			s.Method = "GET"
			s.URL = s.GET
			return nil
		} else if s.POST != "" {
			s.Method = "POST"
			s.URL = s.POST
			return nil
		} else if s.PUT != "" {
			s.Method = "PUT"
			s.URL = s.PUT
			return nil
		} else if s.DELETE != "" {
			s.Method = "DELETE"
			s.URL = s.DELETE
			return nil
		} else if s.PATCH != "" {
			s.Method = "PATCH"
			s.URL = s.PATCH
			return nil
		} else {
			return ErrInvalidAliasOrURL
		}
	}
	if s.Method == "" {
		return ErrInvalidAliasOrURL
	}
	return nil
}

func validateResponseAssertions(resp *ResponseAssertions) error {
	if resp == nil {
		return nil
	}
	if resp.JSON == nil {
		return nil
	}
	if resp.JSON.Schema == "" {
		return nil
	}
	// Ensure any JSONSchema URL specified in response.json.schema exists
	schemaURL := resp.JSON.Schema
	if strings.HasPrefix(schemaURL, "http://") || strings.HasPrefix(schemaURL, "https://") {
		// TODO(jaypipes): Support network lookups?
		return UnsupportedJSONSchemaReference(schemaURL)
	}
	// Convert relative filepaths to absolute filepaths rooted in the context's
	// testdir after stripping any "file://" scheme prefix
	schemaURL = strings.TrimPrefix(schemaURL, "file://")
	schemaURL, _ = filepath.Abs(schemaURL)

	f, err := os.Open(schemaURL)
	if err != nil {
		return JSONSchemaFileNotFound(schemaURL)
	}
	defer f.Close()
	if runtime.GOOS == "windows" {
		// Need to do this because of an "optimization" done in the
		// gojsonreference library:
		// https://github.com/xeipuuv/gojsonreference/blob/bd5ef7bd5415a7ac448318e64f11a24cd21e594b/reference.go#L107-L114
		resp.JSON.Schema = "file:///" + schemaURL
	} else {
		resp.JSON.Schema = "file://" + schemaURL
	}
	return nil
}
