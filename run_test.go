// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http_test

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jaypipes/gdt"
	gdtcontext "github.com/jaypipes/gdt-core/context"
	gdterrors "github.com/jaypipes/gdt-core/errors"
	jsonfix "github.com/jaypipes/gdt-core/fixture/json"
	"github.com/jaypipes/gdt-core/scenario"
	gdttypes "github.com/jaypipes/gdt-core/types"
	gdthttp "github.com/jaypipes/gdt-http"
	"github.com/jaypipes/gdt-http/test/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dataFilePath = "testdata/fixtures.json"
)

type dataset struct {
	Authors    interface{}
	Publishers interface{}
	Books      []*server.Book
}

func data() *dataset {
	f, err := os.Open(dataFilePath)
	if err != nil {
		panic(err)
	}
	data := &dataset{}
	if err = json.NewDecoder(f).Decode(&data); err != nil {
		panic(err)
	}
	return data
}

func dataFixture() gdttypes.Fixture {
	f, err := os.Open(dataFilePath)
	if err != nil {
		panic(err)
	}
	fix, err := jsonfix.New(f)
	if err != nil {
		panic(err)
	}
	return fix
}

func TestFixturesNotSetup(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fp := filepath.Join("testdata", "create-then-get.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(ctx, t)
	assert.NotNil(err)
	assert.ErrorIs(err, gdterrors.ErrRuntime)
}

func setup(ctx context.Context) context.Context {
	// Register an HTTP server fixture that spins up the API service on a
	// random port on localhost
	logger := log.New(os.Stdout, "books_api_http: ", log.LstdFlags)
	srv := server.NewControllerWithBooks(logger, data().Books)
	serverFixture := gdthttp.NewServerFixture(srv.Router(), false /* useTLS */)
	ctx = gdtcontext.RegisterFixture(ctx, "books_api", serverFixture)
	ctx = gdtcontext.RegisterFixture(ctx, "books_data", dataFixture())
	return ctx
}

func TestCreateThenGet(t *testing.T) {
	require := require.New(t)

	fp := filepath.Join("testdata", "create-then-get.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())
	ctx = setup(ctx)

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	s.Run(ctx, t)
}

func TestFailures(t *testing.T) {
	require := require.New(t)

	fp := filepath.Join("testdata", "failures.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())
	ctx = setup(ctx)

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	s.Run(ctx, t)
}

func TestGetBooks(t *testing.T) {
	require := require.New(t)

	fp := filepath.Join("testdata", "get-books.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())
	ctx = setup(ctx)

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(ctx, t)
	require.Nil(err)
}

func TestGetBooksUsingGDT(t *testing.T) {
	// This is testing that the plugin registration for the gdt module (and
	// thus the lack of need to manually register the HTTP plugin) is working
	// properly.
	require := require.New(t)

	fp := filepath.Join("testdata", "get-books.yaml")

	s, err := gdt.From(fp)
	require.Nil(err)
	require.NotNil(s)

	ctx := gdt.NewContext()
	ctx = setup(ctx)
	err = s.Run(ctx, t)
	require.Nil(err)
}

func TestPutMultipleBooks(t *testing.T) {
	require := require.New(t)

	fp := filepath.Join("testdata", "put-multiple-books.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdthttp.Plugin())
	ctx = setup(ctx)

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	s.Run(ctx, t)
}
