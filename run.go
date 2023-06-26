// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	nethttp "net/http"
	"reflect"
	"strings"
	"testing"

	gdtcontext "github.com/jaypipes/gdt-core/context"
	"github.com/jaypipes/gdt-core/result"
	"github.com/stretchr/testify/require"
)

// RunData is data stored in the context about the run. It is fetched from the
// gdtcontext.PriorRun() function and evaluated for things like the special
// `$LOCATION` URL value.
type RunData struct {
	Response *nethttp.Response
}

// priorRunData returns any prior run cached data in the context.
func priorRunData(ctx context.Context) *RunData {
	prData := gdtcontext.PriorRun(ctx)
	httpData, ok := prData[pluginName]
	if !ok {
		return nil
	}
	if data, ok := httpData.(*RunData); ok {
		return data
	}
	return nil
}

// getURL returns the URL to use for the test's HTTP request. The test's url
// field is first queried to see if it is the special $LOCATION string. If it
// is, then we return the previous HTTP response's Location header. Otherwise,
// we construct the URL from the httpFile's base URL and the test's url field.
func (s *Spec) getURL(ctx context.Context) (string, error) {
	if strings.ToUpper(s.URL) == "$LOCATION" {
		pr := priorRunData(ctx)
		if pr == nil || pr.Response == nil {
			panic("test unit referenced $LOCATION before executing an HTTP request")
		}
		url, err := pr.Response.Location()
		if err != nil {
			return "", ErrExpectedLocationHeader
		}
		return url.String(), nil
	}

	d := fromBaseDefaults(s.Defaults)
	base := d.BaseURLFromContext(ctx)
	return base + s.URL, nil
}

// processRequestData looks through the raw data interface{} that was
// unmarshaled during parse for any string values that look like JSONPath
// expressions. If we find any, we query the fixture registry to see if any
// fixtures have a value that matches the JSONPath expression. See
// gdt.fixtures:jsonFixture for more information on how this works
func (s *Spec) processRequestData(ctx context.Context) {
	if s.Data == nil {
		return
	}
	// Get a pointer to the unmarshaled interface{} so we can mutate the
	// contents pointed to
	p := reflect.ValueOf(&s.Data)

	// We're interested in the value pointed to by the interface{}, which is
	// why we do a double Elem() here.
	v := p.Elem().Elem()
	vt := v.Type()

	switch vt.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Elem()
			it := item.Type()
			s.preprocessMap(ctx, item, it.Key(), it.Elem())
		}
		//	ht.f.preprocessSliceValue(v, vt.Key(), vt.Elem())
	case reflect.Map:
		s.preprocessMap(ctx, v, vt.Key(), vt.Elem())
	}
}

// client returns the HTTP client to use when executing HTTP requests. If any
// fixture provides a state with key "http.client", the fixture is asked for
// the HTTP client. Otherwise, we use the net/http.DefaultClient
func (s *Spec) client(ctx context.Context) *nethttp.Client {
	// query the fixture registry to determine if any of them contain an
	// http.client state attribute.
	fixtures := gdtcontext.Fixtures(ctx)
	for _, f := range fixtures {
		if f.HasState(StateKeyClient) {
			c, ok := f.State(StateKeyClient).(*nethttp.Client)
			if !ok {
				panic("fixture failed to return a *net/http.Client")
			}
			return c
		}
	}
	return nethttp.DefaultClient
}

// processRequestDataMap processes a map pointed to by v, transforming any
// string keys or values of the map into the results of calling the fixture
// set's State() method.
func (s *Spec) preprocessMap(
	ctx context.Context,
	m reflect.Value,
	kt reflect.Type,
	vt reflect.Type,
) error {
	it := m.MapRange()
	for it.Next() {
		if kt.Kind() == reflect.String {
			keyStr := it.Key().String()
			fixtures := gdtcontext.Fixtures(ctx)
			for _, f := range fixtures {
				if !f.HasState(keyStr) {
					continue
				}
				trKeyStr := f.State(keyStr)
				keyStr = trKeyStr.(string)
			}

			val := it.Value()
			err := s.preprocessMapValue(ctx, m, reflect.ValueOf(keyStr), val, val.Type())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Spec) preprocessMapValue(
	ctx context.Context,
	m reflect.Value,
	k reflect.Value,
	v reflect.Value,
	vt reflect.Type,
) error {
	if vt.Kind() == reflect.Interface {
		v = v.Elem()
		vt = v.Type()
	}

	switch vt.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			fmt.Println(item)
		}
		fmt.Printf("map element is an array.\n")
	case reflect.Map:
		return s.preprocessMap(ctx, v, vt.Key(), vt.Elem())
	case reflect.String:
		valStr := v.String()
		fixtures := gdtcontext.Fixtures(ctx)
		for _, f := range fixtures {
			if !f.HasState(valStr) {
				continue
			}
			trValStr := f.State(valStr)
			m.SetMapIndex(k, reflect.ValueOf(trValStr))
		}
	default:
		return nil
	}
	return nil
}

// Run executes the test described by the HTTP test. A new HTTP request and
// response pair is created during this call.
func (s *Spec) Run(ctx context.Context, t *testing.T) error {
	var body io.Reader
	if s.Data != nil {
		s.processRequestData(ctx)
		jsonBody, err := json.Marshal(s.Data)
		require.Nil(t, err)
		body = bytes.NewReader(jsonBody)
	}
	runData := &RunData{}
	var rerr error
	t.Run(s.Title(), func(t *testing.T) {
		url, err := s.getURL(ctx)
		if err != nil {
			rerr = err
			return
		}

		req, err := nethttp.NewRequest(s.Method, url, body)
		if err != nil {
			rerr = err
			return
		}

		// TODO(jaypipes): Allow customization of the HTTP client for proxying,
		// TLS, etc
		c := s.client(ctx)

		resp, err := c.Do(req)
		if err != nil {
			rerr = err
			return
		}

		// Make sure we drain and close our response body...
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()

		if s.Response != nil {
			// Only read the response body contents once and pass the byte
			// buffer to the assertion functions
			b, err := ioutil.ReadAll(resp.Body)
			require.Nil(t, err)

			rspec := s.Response
			if rspec.Status != nil {
				assertHTTPStatusEqual(t, resp, *(rspec.Status))
			}

			if rspec.JSON != nil {
				assertJSON(t, resp, b, rspec.JSON)
			}

			if len(rspec.Strings) > 0 {
				for _, exp := range rspec.Strings {
					assertStringInBody(t, resp, b, exp)
				}
			}

			if len(rspec.Headers) > 0 {
				for _, exp := range rspec.Headers {
					assertHeader(t, resp, exp)
				}
			}
		}
		runData.Response = resp
	})
	return result.New(
		result.WithError(rerr),
		result.WithData(pluginName, runData),
	)
}
