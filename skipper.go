// Package echodumpbodyskipper is a body skipper function for some middleware for echo framework
package echodumpbodyskipper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/labstack/echo/v5"
)

// Skipper reports, for the current request, whether the request body and/or
// response body should be skipped from dumping. It is suitable for use as a
// BodySkipper in middlewares such as echo-otel-middleware and
// echo-sentry-middleware.
type Skipper func(c *echo.Context) (skipReqBody bool, skipRespBody bool)

// SkipperConf configures which paths should have their request and/or response
// bodies excluded from dumping.
type SkipperConf struct {
	// DumpNoResponseBodyForPaths lists patterns whose responses should not be
	// dumped. Entries containing `/:` or `/*` (or ending in `*`) are treated as
	// Echo route templates and matched against c.Path(). All other entries are
	// treated as regular expressions matched against the request URL path; they
	// are auto-anchored with ^...$ so a literal like "/users" matches "/users"
	// exactly and not "/users/123".
	DumpNoResponseBodyForPaths []string

	// DumpNoRequestBodyForPaths lists patterns whose requests should not be
	// dumped. See DumpNoResponseBodyForPaths for matching rules.
	DumpNoRequestBodyForPaths []string
}

// isRouteTemplate reports whether p looks like an Echo route template
// (containing a `:param` or `*` segment) rather than a regular expression.
func isRouteTemplate(p string) bool {
	return strings.Contains(p, "/:") || strings.Contains(p, "/*") || strings.HasSuffix(p, "*")
}

func classifyPaths(field string, paths []string) (map[string]struct{}, []*regexp.Regexp, error) {
	exact := make(map[string]struct{}, len(paths))
	regexes := make([]*regexp.Regexp, 0, len(paths))

	for _, path := range paths {
		if isRouteTemplate(path) {
			exact[path] = struct{}{}
			continue
		}

		pattern := path
		if !strings.HasPrefix(pattern, "^") {
			pattern = "^" + pattern
		}

		if !strings.HasSuffix(pattern, "$") {
			pattern += "$"
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: invalid regex %q: %w", field, path, err)
		}

		regexes = append(regexes, re)
	}

	return exact, regexes, nil
}

func isExcluded(path string, endpoint string, regexps []*regexp.Regexp, endpoints map[string]struct{}) bool {
	if _, ok := endpoints[endpoint]; ok {
		return true
	}

	for _, regexExcludedPath := range regexps {
		if regexExcludedPath.MatchString(path) {
			return true
		}
	}

	return false
}

// New builds a Skipper from the given configuration. When the config is empty,
// the returned Skipper is a no-op that always returns (false, false). An error
// is returned if any non-template pattern in the configuration fails to
// compile as a regular expression.
func New(config SkipperConf) (Skipper, error) {
	if len(config.DumpNoResponseBodyForPaths) == 0 && len(config.DumpNoRequestBodyForPaths) == 0 {
		return func(*echo.Context) (bool, bool) {
			return false, false
		}, nil
	}

	exactResp, regexResp, err := classifyPaths("DumpNoResponseBodyForPaths", config.DumpNoResponseBodyForPaths)
	if err != nil {
		return nil, err
	}

	exactReq, regexReq, err := classifyPaths("DumpNoRequestBodyForPaths", config.DumpNoRequestBodyForPaths)
	if err != nil {
		return nil, err
	}

	return func(c *echo.Context) (bool, bool) {
		urlPath := c.Request().URL.Path
		route := c.Path()
		skipReqBody := isExcluded(urlPath, route, regexReq, exactReq)
		skipRespBody := isExcluded(urlPath, route, regexResp, exactResp)

		return skipReqBody, skipRespBody
	}, nil
}
