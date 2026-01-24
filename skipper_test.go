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
			name: "invalid regex is ignored",
			conf: SkipperConf{
				DumpNoResponseBodyForPaths: []string{
					"[",
				},
			},
			method:       http.MethodGet,
			path:         "/ping/121",
			route:        "/ping/:id",
			wantSkipReq:  false,
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
			skipper := NewEchoDumpBodySkipper(tt.conf)

			ctx := newContext(tt.method, tt.path, tt.route, tt.body)
			skipReqBody, skipRespBody := skipper.Skip(ctx)
			require.Equal(t, tt.wantSkipReq, skipReqBody)
			require.Equal(t, tt.wantSkipResp, skipRespBody)
		})
	}
}
