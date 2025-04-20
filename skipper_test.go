package echodumpbodyskipper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestSkipper(t *testing.T) {
	t.Run("exclude ping from resp", func(t *testing.T) {
		skipper := NewEchoDumpBodySkipper(SkipperConf{
			DumpNoResponseBodyForPaths: []string{
				"^\\/ping\\/121",
			},
		})

		router := echo.New()
		router.GET("/ping/:id", func(c echo.Context) error {
			skipReqBody, skipRespBody := skipper.Skip(c)
			require.False(t, skipReqBody)
			require.True(t, skipRespBody)

			return c.String(http.StatusOK, "ok")
		})
		r := httptest.NewRequest("GET", "/ping/121?sdsdds=1212", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

	})

	t.Run("exclude ping from req", func(t *testing.T) {
		skipper := NewEchoDumpBodySkipper(SkipperConf{
			DumpNoRequestBodyForPaths: []string{
				"/ping/:id",
			},
		})

		router := echo.New()
		router.GET("/ping/:id", func(c echo.Context) error {
			skipReqBody, skipRespBody := skipper.Skip(c)
			require.True(t, skipReqBody)
			require.False(t, skipRespBody)

			return c.String(http.StatusOK, "ok")
		})
		r := httptest.NewRequest("GET", "/ping/123", strings.NewReader("test"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		response := w.Result()
		require.Equal(t, http.StatusOK, response.StatusCode)
	})
}
