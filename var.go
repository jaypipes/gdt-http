// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gdt-dev/core/api"
	gdtjson "github.com/gdt-dev/core/assertion/json"
	"github.com/gdt-dev/core/debug"
	"github.com/theory/jsonpath"
)

type VarEntry struct {
	// From is a string that indicates where the value of the variable will be
	// sourced from. This string is a JSONPath expression that contains
	// instructions on how to extract a particular field from the HTTP response.
	From string `yaml:"from"`
}

// Variables allows the test author to save arbitrary data to the test scenario,
// facilitating the passing of variables between test specs potentially
// provided by different gdt Plugins.
type Variables map[string]VarEntry

// saveVars examines the supplied Variables and what we got back from the
// Action.Do() call and sets any variables in the run data context key.
func saveVars(
	ctx context.Context,
	vars Variables,
	body []byte,
	res *api.Result,
) error {
	if len(body) == 0 {
		return nil
	}
	var bodyMap any
	if err := json.Unmarshal(body, &bodyMap); err != nil {
		return err
	}
	for varName, entry := range vars {
		path := entry.From
		extracted, err := extractFrom(path, bodyMap)
		if err != nil {
			return err
		}
		debug.Printf(ctx, "save.vars: %s -> %v", varName, extracted)
		res.SetData(varName, extracted)
	}
	return nil
}

func extractFrom(path string, out any) (any, error) {
	var normalized any
	switch out := out.(type) {
	case map[string]any:
		normalized = out
	case []map[string]any:
		normalized = out
	default:
		return nil, fmt.Errorf("unhandled extract type %T", out)
	}
	p, err := jsonpath.Parse(path)
	if err != nil {
		// Not terminal because during parse we validate the JSONPath
		// expression is valid.
		return nil, gdtjson.JSONPathNotFound(path, err)
	}
	nodes := p.Select(normalized)
	if len(nodes) == 0 {
		return nil, gdtjson.JSONPathNotFound(path, err)
	}
	got := nodes[0]
	return got, nil
}
