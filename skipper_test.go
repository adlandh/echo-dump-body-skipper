package echodumpbodyskipper

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func newContext(method, path, route string, body io.Reader) *echo.Context {
	router := echo.New()
	req := httptest.NewRequest(method, path, body)
	rec := httptest.NewRecorder()
	ctx := router.NewContext(req, rec)
	ctx.SetPath(route)

	return ctx
}

func TestSkipper(t *testing.T) {
	req := func(paths ...string) SkipperConf { return SkipperConf{DumpNoRequestBodyForPaths: paths} }
	resp := func(paths ...string) SkipperConf { return SkipperConf{DumpNoResponseBodyForPaths: paths} }
	both := func(reqs, resps []string) SkipperConf {
		return SkipperConf{DumpNoRequestBodyForPaths: reqs, DumpNoResponseBodyForPaths: resps}
	}

	tests := []struct {
		name                      string
		conf                      SkipperConf
		path, route               string
		wantSkipReq, wantSkipResp bool
	}{
		{"no config returns false", SkipperConf{}, "/ping/121?qs=1", "/ping/:id", false, false},
		{"exclude response body via regex", resp("^/ping/121$"), "/ping/121?sdsdds=1212", "/ping/:id", false, true},
		{"exclude request body via endpoint", req("/ping/:id"), "/ping/123", "/ping/:id", true, false},
		{"exclude both request and response bodies", both([]string{"/ping/:id"}, []string{"^/ping/121$"}), "/ping/121", "/ping/:id", true, true},
		{"literal path is anchored - does not match longer URL", resp("/users"), "/users/123", "/users/:id", false, false},
		{"literal path matches exactly", resp("/users"), "/users", "/users", false, true},
		{"route template is not treated as regex against URL path", req("/ping/:id"), "/ping/:id-debug", "/other/:id", false, false},
		{"wildcard route template matched via endpoint", req("/files/*"), "/files/a/b", "/files/*", true, false},
		{"regex non-match does not skip", resp("^/pong/123$"), "/ping/121", "/ping/:id", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipper, err := New(tt.conf)
			require.NoError(t, err)

			ctx := newContext(http.MethodGet, tt.path, tt.route, nil)
			skipReqBody, skipRespBody := skipper(ctx)
			require.Equal(t, tt.wantSkipReq, skipReqBody)
			require.Equal(t, tt.wantSkipResp, skipRespBody)
		})
	}
}

func TestNew_InvalidRegexReturnsError(t *testing.T) {
	tests := []struct {
		name     string
		conf     SkipperConf
		wantField string
	}{
		{
			name:      "invalid request pattern",
			conf:      SkipperConf{DumpNoRequestBodyForPaths: []string{"["}},
			wantField: "DumpNoRequestBodyForPaths",
		},
		{
			name:      "invalid response pattern",
			conf:      SkipperConf{DumpNoResponseBodyForPaths: []string{"("}},
			wantField: "DumpNoResponseBodyForPaths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.conf)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantField)
		})
	}
}
