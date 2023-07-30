// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gdt-dev/gdt"
	gdtjson "github.com/gdt-dev/gdt/assertion/json"
	"github.com/gdt-dev/gdt/errors"
	gdttypes "github.com/gdt-dev/gdt/types"
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
	s, err := gdt.From(fp)
	require.NotNil(err)
	assert.ErrorIs(err, errors.ErrExpectedMap)
	require.Nil(s)
}

func TestParseFailures(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse", "fail", "invalid.yaml")

	s, err := gdt.From(fp)
	require.NotNil(err)
	assert.ErrorIs(err, errors.ErrExpectedMap)
	require.Nil(s)
}

func TestMissingSchema(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse", "fail", "missing-schema.yaml")

	s, err := gdt.From(fp)
	require.NotNil(err)
	assert.ErrorIs(err, gdtjson.ErrJSONSchemaFileNotFound)
	require.Nil(s)
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse.yaml")

	suite, err := gdt.From(fp)
	require.Nil(err)
	require.NotNil(suite)

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

	require.Len(suite.Scenarios, 1)
	s := suite.Scenarios[0]

	expTests := []gdttypes.Evaluable{
		&gdthttp.Spec{
			Spec: gdttypes.Spec{
				Index:    0,
				Name:     "no such book was found",
				Defaults: &gdttypes.Defaults{},
			},
			Method: "GET",
			URL:    "/books/nosuchbook",
			GET:    "/books/nosuchbook",
			Response: &gdthttp.Expect{
				JSON: &gdtjson.Expect{
					Len: &len0,
				},
				Status: &code404,
			},
		},
		&gdthttp.Spec{
			Spec: gdttypes.Spec{
				Index:    1,
				Name:     "list all books",
				Defaults: &gdttypes.Defaults{},
			},
			Method: "GET",
			URL:    "/books",
			GET:    "/books",
			Response: &gdthttp.Expect{
				JSON: &gdtjson.Expect{
					Schema: schemaPath,
				},
				Status: &code200,
			},
		},
		&gdthttp.Spec{
			Spec: gdttypes.Spec{
				Index:    2,
				Name:     "create a new book",
				Defaults: &gdttypes.Defaults{},
			},
			Method: "POST",
			URL:    "/books",
			POST:   "/books",
			Data: map[string]interface{}{
				"title":        "For Whom The Bell Tolls",
				"published_on": publishedOn1940,
				"pages":        480,
				"author_id":    "$.authors.by_name[\"Ernest Hemingway\"].id",
				"publisher_id": "$.publishers.by_name[\"Charles Scribner's Sons\"].id",
			},
			Response: &gdthttp.Expect{
				Status: &code201,
				Headers: []string{
					"Location",
				},
			},
		},
		&gdthttp.Spec{
			Spec: gdttypes.Spec{
				Index:    3,
				Name:     "look up that created book",
				Defaults: &gdttypes.Defaults{},
			},
			Method: "GET",
			URL:    "$LOCATION",
			GET:    "$LOCATION",
			Response: &gdthttp.Expect{
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
			Spec: gdttypes.Spec{
				Index:    4,
				Name:     "create two books",
				Defaults: &gdttypes.Defaults{},
			},
			Method: "PUT",
			URL:    "/books",
			PUT:    "/books",
			Data: []interface{}{
				map[string]interface{}{
					"title":        "For Whom The Bell Tolls",
					"published_on": publishedOn1940,
					"pages":        480,
					"author_id":    "$.authors.by_name[\"Ernest Hemingway\"].id",
					"publisher_id": "$.publishers.by_name[\"Charles Scribner's Sons\"].id",
				},
				map[string]interface{}{
					"title":        "To Have and Have Not",
					"published_on": publishedOn1937,
					"pages":        257,
					"author_id":    "$.authors.by_name[\"Ernest Hemingway\"].id",
					"publisher_id": "$.publishers.by_name[\"Charles Scribner's Sons\"].id",
				},
			},
			Response: &gdthttp.Expect{
				Status: &code200,
			},
		},
	}
	assert.Equal(expTests, s.Tests)
}
