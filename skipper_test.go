package echodumpbodyskipper

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	tests := []struct {
		name                      string
		conf                      SkipperConf
		method                    string
		path                      string
		route                     string
		body                      io.Reader
		wantSkipReq, wantSkipResp bool
	}{
		{
			name:         "no config returns false",
			conf:         SkipperConf{},
			method:       http.MethodGet,
			path:         "/ping/121?qs=1",
			route:        "/ping/:id",
			wantSkipReq:  false,
			wantSkipResp: false,
		},
		{
			name: "exclude response body via regex",
			conf: SkipperConf{
				DumpNoResponseBodyForPaths: []string{
					"^/ping/121$",
				},
			},
			method:       http.MethodGet,
			path:         "/ping/121?sdsdds=1212",
			route:        "/ping/:id",
			wantSkipReq:  false,
			wantSkipResp: true,
		},
		{
			name: "exclude request body via endpoint",
			conf: SkipperConf{
				DumpNoRequestBodyForPaths: []string{
					"/ping/:id",
				},
			},
			method:       http.MethodGet,
			path:         "/ping/123",
			route:        "/ping/:id",
			body:         strings.NewReader("test"),
			wantSkipReq:  true,
			wantSkipResp: false,
		},
		{
			name: "exclude both request and response bodies",
			conf: SkipperConf{
				DumpNoRequestBodyForPaths: []string{
					"/ping/:id",
				},
				DumpNoResponseBodyForPaths: []string{
					"^/ping/121$",
				},
			},
			method:       http.MethodGet,
			path:         "/ping/121",
			route:        "/ping/:id",
			body:         strings.NewReader("test"),
			wantSkipReq:  true,
			wantSkipResp: true,
		},
		{
			name: "literal path is anchored - does not match longer URL",
			conf: SkipperConf{
				DumpNoResponseBodyForPaths: []string{
					"/users",
				},
			},
			method:       http.MethodGet,
			path:         "/users/123",
			route:        "/users/:id",
			wantSkipReq:  false,
			wantSkipResp: false,
		},
		{
			name: "literal path matches exactly",
			conf: SkipperConf{
				DumpNoResponseBodyForPaths: []string{
					"/users",
				},
			},
			method:       http.MethodGet,
			path:         "/users",
			route:        "/users",
			wantSkipReq:  false,
			wantSkipResp: true,
		},
		{
			name: "route template is not treated as regex against URL path",
			conf: SkipperConf{
				DumpNoRequestBodyForPaths: []string{
					"/ping/:id",
				},
			},
			method:       http.MethodGet,
			path:         "/ping/:id-debug",
			route:        "/other/:id",
			wantSkipReq:  false,
			wantSkipResp: false,
		},
		{
			name: "wildcard route template matched via endpoint",
			conf: SkipperConf{
				DumpNoRequestBodyForPaths: []string{
					"/files/*",
				},
			},
			method:       http.MethodGet,
			path:         "/files/a/b",
			route:        "/files/*",
			body:         strings.NewReader("test"),
			wantSkipReq:  true,
			wantSkipResp: false,
		},
		{
			name: "regex non-match does not skip",
			conf: SkipperConf{
				DumpNoResponseBodyForPaths: []string{
					"^/pong/123$",
				},
			},
			method:       http.MethodGet,
			path:         "/ping/121",
			route:        "/ping/:id",
			wantSkipReq:  false,
			wantSkipResp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipper, err := New(tt.conf)
			require.NoError(t, err)

			ctx := newContext(tt.method, tt.path, tt.route, tt.body)
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
