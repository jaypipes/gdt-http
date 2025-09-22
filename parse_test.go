// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gdt-dev/core/api"
	gdtjson "github.com/gdt-dev/core/assertion/json"
	"github.com/gdt-dev/core/parse"
	"github.com/gdt-dev/core/scenario"
	gdthttp "github.com/gdt-dev/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func currentDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func TestBadDefaults(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse", "fail", "bad-defaults.yaml")
	f, err := os.Open(fp)
	require.Nil(err)
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.NotNil(err)
	assert.Error(err, &parse.Error{})
	assert.ErrorContains(err, "expected map")
	require.Nil(s)
}

func TestParseFailures(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse", "fail", "invalid.yaml")
	f, err := os.Open(fp)
	require.Nil(err)
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.NotNil(err)
	assert.Error(err, &parse.Error{})
	assert.ErrorContains(err, "expected map")
	require.Nil(s)
}

func TestMissingSchema(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse", "fail", "missing-schema.yaml")
	f, err := os.Open(fp)
	require.Nil(err)
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.NotNil(err)
	assert.ErrorContains(err, "unable to find JSONSchema file")
	require.Nil(s)
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse.yaml")
	f, err := os.Open(fp)
	require.Nil(err)
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.Nil(err)

	code404 := 404
	code200 := 200
	code201 := 201
	len0 := 0
	dateOnly := "2006-01-02"
	publishedOn1940, _ := time.Parse(dateOnly, "1940-10-21")
	publishedOn1937, _ := time.Parse(dateOnly, "1937-10-15")

	pathParts := []string{
		"file://",
		filepath.Join(
			currentDir(),
			"testdata", "schemas", "get_books.json",
		),
	}
	if runtime.GOOS == "windows" {
		// Need to do this because of an "optimization" done in the
		// gojsonreference library:
		// https://github.com/xeipuuv/gojsonreference/blob/bd5ef7bd5415a7ac448318e64f11a24cd21e594b/reference.go#L107-L114
		pathParts[0] = "file:///"
	}
	schemaPath := strings.Join(pathParts, "")

	expTests := []api.Evaluable{
		&gdthttp.Spec{
			Spec: api.Spec{
				Index:    0,
				Name:     "no such book was found",
				Defaults: &api.Defaults{},
			},
			HTTP: &gdthttp.HTTPSpec{
				Action: gdthttp.Action{
					Method: "GET",
					URL:    "/books/nosuchbook",
				},
			},
			Assert: &gdthttp.Expect{
				JSON: &gdtjson.Expect{
					Len: &len0,
				},
				Status: &code404,
			},
		},
		&gdthttp.Spec{
			Spec: api.Spec{
				Index:    1,
				Name:     "list all books",
				Defaults: &api.Defaults{},
			},
			HTTP: &gdthttp.HTTPSpec{
				Action: gdthttp.Action{
					Method: "GET",
					URL:    "/books",
				},
			},
			Assert: &gdthttp.Expect{
				JSON: &gdtjson.Expect{
					Schema: schemaPath,
				},
				Status: &code200,
			},
		},
		&gdthttp.Spec{
			Spec: api.Spec{
				Index:    2,
				Name:     "create a new book",
				Defaults: &api.Defaults{},
			},
			HTTP: &gdthttp.HTTPSpec{
				Action: gdthttp.Action{
					Method: "POST",
					URL:    "/books",
					Data: map[string]any{
						"title":        "For Whom The Bell Tolls",
						"published_on": publishedOn1940,
						"pages":        480,
						"author_id":    "$.authors.by_name[\"Ernest Hemingway\"].id",
						"publisher_id": "$.publishers.by_name[\"Charles Scribner's Sons\"].id",
					},
				},
			},
			Assert: &gdthttp.Expect{
				Status: &code201,
				Headers: []string{
					"Location",
				},
			},
		},
		&gdthttp.Spec{
			Spec: api.Spec{
				Index:    3,
				Name:     "look up that created book",
				Defaults: &api.Defaults{},
			},
			HTTP: &gdthttp.HTTPSpec{
				Action: gdthttp.Action{
					Method: "GET",
					URL:    "$LOCATION",
				},
			},
			Assert: &gdthttp.Expect{
				JSON: &gdtjson.Expect{
					Paths: map[string]string{
						"$.author.name":             "Ernest Hemingway",
						"$.publisher.address.state": "NY",
					},
					PathFormats: map[string]string{
						"$.id": "uuid4",
					},
				},
				Status: &code200,
			},
		},
		&gdthttp.Spec{
			Spec: api.Spec{
				Index:    4,
				Name:     "create two books",
				Defaults: &api.Defaults{},
			},
			HTTP: &gdthttp.HTTPSpec{
				Action: gdthttp.Action{
					Method: "PUT",
					URL:    "/books",
					Data: []any{
						map[string]any{
							"title":        "For Whom The Bell Tolls",
							"published_on": publishedOn1940,
							"pages":        480,
							"author_id":    "$.authors.by_name[\"Ernest Hemingway\"].id",
							"publisher_id": "$.publishers.by_name[\"Charles Scribner's Sons\"].id",
						},
						map[string]any{
							"title":        "To Have and Have Not",
							"published_on": publishedOn1937,
							"pages":        257,
							"author_id":    "$.authors.by_name[\"Ernest Hemingway\"].id",
							"publisher_id": "$.publishers.by_name[\"Charles Scribner's Sons\"].id",
						},
					},
				},
			},
			Assert: &gdthttp.Expect{
				Status: &code200,
			},
		},
	}
	require.Len(s.Tests, len(expTests))
	for x, st := range s.Tests {
		exp := expTests[x].(*gdthttp.Spec)
		sth := st.(*gdthttp.Spec)
		assert.Equal(exp.HTTP, sth.HTTP)
		assert.Equal(exp.Var, sth.Var)
		assert.Equal(exp.Assert, sth.Assert)
	}
}
