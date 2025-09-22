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
	nethttp "net/http"
	"reflect"
	"strings"

	gdtcontext "github.com/gdt-dev/core/context"
	"github.com/gdt-dev/core/debug"
)

// Action describes the the HTTP-specific action that is performed by the test.
type Action struct {
	// URL being called by HTTP client. Used with the `Method` field.
	URL string `yaml:"url,omitempty"`
	// HTTP Method specified by HTTP client. Used with the `URL` shortcut field.
	Method string `yaml:"method,omitempty"`
	// Data is the payload to send along in request
	Data interface{} `yaml:"data,omitempty"`
	// Shortcut for URL and Method of "GET"
	Get string `yaml:"get,omitempty"`
	// Shortcut for URL and Method of "POST"
	Post string `yaml:"post,omitempty"`
	// Shortcut for URL and Method of "PUT"
	Put string `yaml:"put,omitempty"`
	// Shortcut for URL and Method of "PATCH"
	Patch string `yaml:"patch,omitempty"`
	// Shortcut for URL and Method of "DELETE"
	Delete string `yaml:"delete,omitempty"`
}

// Do performs a single HTTP request, returning the HTTP Response and any
// runtime error.
func (a *Action) Do(
	ctx context.Context,
	c *nethttp.Client,
	defaults *Defaults,
) (*nethttp.Response, error) {
	url, err := a.getURL(ctx, defaults)
	if err != nil {
		return nil, err
	}

	debug.Printf(ctx, "http: > %s %s", a.Method, url)
	var reqData io.Reader
	if a.Data != nil {
		if err := a.processRequestData(ctx); err != nil {
			return nil, err
		}
		jsonBody, err := json.Marshal(a.Data)
		if err != nil {
			return nil, err
		}
		b := bytes.NewReader(jsonBody)
		if b.Size() > 0 {
			sendData, _ := io.ReadAll(b)
			debug.Printf(ctx, "http: > %s", sendData)
			b.Seek(0, 0) // nolint:errcheck
		}
		reqData = b
	}

	req, err := nethttp.NewRequest(a.Method, url, reqData)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	debug.Printf(ctx, "http: < %d", resp.StatusCode)
	return resp, err
}

// getURL returns the URL to use for the test's HTTP request. The test's url
// field is first queried to see if it is the special $LOCATION string. If it
// is, then we return the previous HTTP response's Location header. Otherwise,
// we construct the URL from the httpFile's base URL and the test's url field.
func (a *Action) getURL(
	ctx context.Context,
	defaults *Defaults,
) (string, error) {
	if strings.ToUpper(a.URL) == "$LOCATION" {
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
	base := defaults.BaseURLFromContext(ctx)
	return base + a.URL, nil
}

// processRequestData looks through the raw data interface{} that was
// unmarshaled during parse for any string values that look like JSONPath
// expressions. If we find any, we query the fixture registry to see if any
// fixtures have a value that matches the JSONPath expression. See
// gdt.fixtures:jsonFixture for more information on how this works
func (a *Action) processRequestData(ctx context.Context) error {
	if a.Data == nil {
		return nil
	}
	// Get a pointer to the unmarshaled interface{} so we can mutate the
	// contents pointed to
	p := reflect.ValueOf(&a.Data)

	// We're interested in the value pointed to by the interface{}, which is
	// why we do a double Elem() here.
	v := p.Elem().Elem()
	vt := v.Type()

	switch vt.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Elem()
			it := item.Type()
			err := a.preprocessMap(ctx, item, it.Key(), it.Elem())
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		err := a.preprocessMap(ctx, v, vt.Key(), vt.Elem())
		if err != nil {
			return err
		}
	}
	return nil
}

// processRequestDataMap processes a map pointed to by v, transforming any
// string keys or values of the map into the results of calling the fixture
// set's State() method.
func (a *Action) preprocessMap(
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
			err := a.preprocessMapValue(ctx, m, reflect.ValueOf(keyStr), val, val.Type())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Action) preprocessMapValue(
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
		return a.preprocessMap(ctx, v, vt.Key(), vt.Elem())
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
