// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"context"

	gdtcontext "github.com/jaypipes/gdt-core/context"
	"github.com/jaypipes/gdt-core/errors"
	"gopkg.in/yaml.v3"
)

type httpDefaults struct {
	BaseURL string `json:"base_url,omitempty"`
}

// Defaults is the known HTTP plugin defaults collection
type Defaults struct {
	httpDefaults
}

func (d *Defaults) UnmarshalYAML(node *yaml.Node) error {
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
		case "http":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			hd := httpDefaults{}
			if err := valNode.Decode(&hd); err != nil {
				return err
			}
			d.httpDefaults = hd
		default:
			continue
		}
	}
	return nil
}

// BaseURLFromContext returns the base URL to use when constructing HTTP
// requests. If the Defaults is non-nil and has a BaseURL value, use that.
// Otherwise we look up a base URL from the context's fixtures.
func (d *Defaults) BaseURLFromContext(ctx context.Context) string {
	// If the httpFile has been manually configured and the configuration
	// contains a base URL, use that. Otherwise, check to see if there is a
	// fixture in the registry that has an "http.base_url" state key and use
	// that if found.
	if d != nil && d.BaseURL != "" {
		return d.BaseURL
	}
	// query the fixture registry to determine if any of them contain an
	// http.base_url state attribute.
	fixtures := gdtcontext.Fixtures(ctx)
	for _, f := range fixtures {
		if f.HasState(StateKeyBaseURL) {
			return f.State(StateKeyBaseURL).(string)
		}
	}
	return ""
}
