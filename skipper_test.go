package echodumpbodyskipper

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func newContext(method, path, route string, body io.Reader) echo.Context {
	router := echo.New()
	req := httptest.NewRequest(method, path, body)
	rec := httptest.NewRecorder()
	ctx := router.NewContext(req, rec)
	ctx.SetPath(route)

	return ctx
}

func TestSkipper(t *testing.T) {
	t.Run("no config returns false", func(t *testing.T) {
		skipper := NewEchoDumpBodySkipper(SkipperConf{})

		ctx := newContext(http.MethodGet, "/ping/121?qs=1", "/ping/:id", nil)
		skipReqBody, skipRespBody := skipper.Skip(ctx)
		require.False(t, skipReqBody)
		require.False(t, skipRespBody)
	})

	t.Run("exclude response body via regex", func(t *testing.T) {
		skipper := NewEchoDumpBodySkipper(SkipperConf{
			DumpNoResponseBodyForPaths: []string{
				"^/ping/121$",
			},
		})

		ctx := newContext(http.MethodGet, "/ping/121?sdsdds=1212", "/ping/:id", nil)
		skipReqBody, skipRespBody := skipper.Skip(ctx)
		require.False(t, skipReqBody)
		require.True(t, skipRespBody)
	})

	t.Run("exclude request body via endpoint", func(t *testing.T) {
		skipper := NewEchoDumpBodySkipper(SkipperConf{
			DumpNoRequestBodyForPaths: []string{
				"/ping/:id",
			},
		})

		ctx := newContext(http.MethodGet, "/ping/123", "/ping/:id", strings.NewReader("test"))
		skipReqBody, skipRespBody := skipper.Skip(ctx)
		require.True(t, skipReqBody)
		require.False(t, skipRespBody)
	})

	t.Run("invalid regex is ignored", func(t *testing.T) {
		skipper := NewEchoDumpBodySkipper(SkipperConf{
			DumpNoResponseBodyForPaths: []string{
				"[",
			},
		})

		ctx := newContext(http.MethodGet, "/ping/121", "/ping/:id", nil)
		skipReqBody, skipRespBody := skipper.Skip(ctx)
		require.False(t, skipReqBody)
		require.False(t, skipRespBody)
	})

	t.Run("exclude both request and response bodies", func(t *testing.T) {
		skipper := NewEchoDumpBodySkipper(SkipperConf{
			DumpNoRequestBodyForPaths: []string{
				"/ping/:id",
			},
			DumpNoResponseBodyForPaths: []string{
				"^/ping/121$",
			},
		})

		ctx := newContext(http.MethodGet, "/ping/121", "/ping/:id", strings.NewReader("test"))
		skipReqBody, skipRespBody := skipper.Skip(ctx)
		require.True(t, skipReqBody)
		require.True(t, skipRespBody)
	})

	t.Run("regex non-match does not skip", func(t *testing.T) {
		skipper := NewEchoDumpBodySkipper(SkipperConf{
			DumpNoResponseBodyForPaths: []string{
				"^/pong/123$",
			},
		})

		ctx := newContext(http.MethodGet, "/ping/121", "/ping/:id", nil)
		skipReqBody, skipRespBody := skipper.Skip(ctx)
		require.False(t, skipReqBody)
		require.False(t, skipRespBody)
	})
}
