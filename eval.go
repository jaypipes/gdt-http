// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package http

import (
	"context"
	"io"
	"io/ioutil"
	nethttp "net/http"

	"github.com/gdt-dev/core/api"
	gdtcontext "github.com/gdt-dev/core/context"
	"github.com/gdt-dev/core/debug"
)

// Run executes the test described by the HTTP test. A new HTTP request and
// response pair is created during this call.
func (s *Spec) Eval(ctx context.Context) (*api.Result, error) {
	c := client(ctx)
	defaults := fromBaseDefaults(s.Defaults)
	runData := &RunData{}

	resp, err := s.HTTP.Do(ctx, c, defaults)
	if err != nil {
		if err == api.ErrTimeoutExceeded {
			return api.NewResult(api.WithFailures(api.ErrTimeoutExceeded)), nil
		}
		return nil, err
	}

	// Make sure we drain and close our response body...
	defer func() {
		io.Copy(ioutil.Discard, resp.Body) // nolint:errcheck
		resp.Body.Close()                  // nolint:errcheck
	}()

	// Only read the response body contents once and pass the byte
	// buffer to the assertion functions
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(body) > 0 {
		debug.Printf(ctx, "http: < %s", string(body))
	}
	a := newAssertions(s.Assert, resp, body)
	if a.OK(ctx) {
		runData.Response = resp
		res := api.NewResult()
		res.SetData(pluginName, runData)
		if err := saveVars(ctx, s.Var, body, res); err != nil {
			return nil, err
		}
		return res, nil
	}
	return api.NewResult(api.WithFailures(a.Failures()...)), nil
}

// client returns the HTTP client to use when executing HTTP requests. If any
// fixture provides a state with key "http.client", the fixture is asked for
// the HTTP client. Otherwise, we use the net/http.DefaultClient
func client(ctx context.Context) *nethttp.Client {
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

// RunData is data stored in the context about the run. It is fetched from the
// gdtcontext.PriorRun() function and evaluated for things like the special
// `$LOCATION` URL value.
type RunData struct {
	Response *nethttp.Response
}

// priorRunData returns any prior run cached data in the context.
func priorRunData(ctx context.Context) *RunData {
	data := gdtcontext.Run(ctx)
	httpData, ok := data[pluginName]
	if !ok {
		return nil
	}
	if data, ok := httpData.(*RunData); ok {
		return data
	}
	return nil
}
