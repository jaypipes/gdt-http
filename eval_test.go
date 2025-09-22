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

	"github.com/gdt-dev/core/api"
	gdtcontext "github.com/gdt-dev/core/context"
	gdtjsonfix "github.com/gdt-dev/core/fixture/json"
	"github.com/gdt-dev/core/scenario"
	gdthttp "github.com/gdt-dev/http"
	"github.com/gdt-dev/http/test/server"
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

func dataFixture() api.Fixture {
	f, err := os.Open(dataFilePath)
	if err != nil {
		panic(err)
	}
	fix, err := gdtjsonfix.New(f)
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
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(context.TODO(), t)
	require.NotNil(err)
	assert.ErrorIs(err, api.RuntimeError)
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
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.Nil(err)
	require.NotNil(s)

	ctx := gdtcontext.New()
	ctx = setup(ctx)

	s.Run(ctx, t)
}

func TestFailures(t *testing.T) {
	require := require.New(t)

	fp := filepath.Join("testdata", "failures.yaml")
	f, err := os.Open(fp)
	require.Nil(err)
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.Nil(err)
	require.NotNil(s)

	ctx := gdtcontext.New()
	ctx = setup(ctx)

	s.Run(ctx, t)
}

func TestGetBooks(t *testing.T) {
	require := require.New(t)

	fp := filepath.Join("testdata", "get-books.yaml")
	f, err := os.Open(fp)
	require.Nil(err)
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.Nil(err)
	require.NotNil(s)

	ctx := gdtcontext.New()
	ctx = setup(ctx)

	s.Run(ctx, t)
}

func TestPutMultipleBooks(t *testing.T) {
	require := require.New(t)

	fp := filepath.Join("testdata", "put-multiple-books.yaml")
	f, err := os.Open(fp)
	require.Nil(err)
	defer f.Close() // nolint:errcheck

	s, err := scenario.FromReader(f, scenario.WithPath(fp))
	require.Nil(err)
	require.NotNil(s)

	ctx := gdtcontext.New()
	ctx = setup(ctx)

	s.Run(ctx, t)
}
