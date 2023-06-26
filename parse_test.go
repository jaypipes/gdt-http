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

	gdtcontext "github.com/jaypipes/gdt-core/context"
	"github.com/jaypipes/gdt-core/errors"
	"github.com/jaypipes/gdt-core/scenario"
	"github.com/jaypipes/gdt-core/spec"
	gdttypes "github.com/jaypipes/gdt-core/types"
	gdthttp "github.com/jaypipes/gdt-http"
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

	fp := filepath.Join("testdata", "failures", "bad-defaults.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.NotNil(err)
	assert.ErrorIs(err, errors.ErrInvalidExpectedMap)
	assert.Nil(s)
}

func TestParseFailures(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse-failure.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.NotNil(err)
	assert.ErrorIs(err, errors.ErrInvalid)
	assert.Nil(s)
}

func TestMissingSchema(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "missing-schema.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.NotNil(err)
	assert.ErrorIs(err, errors.ErrInvalid)
	assert.Nil(s)
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.Nil(err)
	assert.NotNil(s)

	code404 := 404
	code200 := 200
	code201 := 201
	len0 := uint(0)
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

	assert.IsType(&scenario.Scenario{}, s)
	sc := s.(*scenario.Scenario)
	expTests := []gdttypes.Spec{
		&gdthttp.Spec{
			Spec: spec.Spec{
				Index: 0,
				Name:  "no such book was found",
			},
			Method: "GET",
			URL:    "/books/nosuchbook",
			GET:    "/books/nosuchbook",
			Response: &gdthttp.ResponseAssertions{
				JSON: &gdthttp.JSONAssertions{
					Length: &len0,
				},
				Status: &code404,
			},
		},
		&gdthttp.Spec{
			Spec: spec.Spec{
				Index: 1,
				Name:  "list all books",
			},
			Method: "GET",
			URL:    "/books",
			GET:    "/books",
			Response: &gdthttp.ResponseAssertions{
				JSON: &gdthttp.JSONAssertions{
					Schema: schemaPath,
				},
				Status: &code200,
			},
		},
		&gdthttp.Spec{
			Spec: spec.Spec{
				Index: 2,
				Name:  "create a new book",
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
			Response: &gdthttp.ResponseAssertions{
				Status: &code201,
				Headers: []string{
					"Location",
				},
			},
		},
		&gdthttp.Spec{
			Spec: spec.Spec{
				Index: 3,
				Name:  "look up that created book",
			},
			Method: "GET",
			URL:    "$LOCATION",
			GET:    "$LOCATION",
			Response: &gdthttp.ResponseAssertions{
				JSON: &gdthttp.JSONAssertions{
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
			Spec: spec.Spec{
				Index: 4,
				Name:  "create two books",
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
			Response: &gdthttp.ResponseAssertions{
				Status: &code200,
			},
		},
	}
	assert.Equal(expTests, sc.Tests)
}
