// Package echodumpbodyskipper is a body skipper function for some middleware for echo framework
package echodumpbodyskipper

import (
	"regexp"

	"github.com/labstack/echo/v4"
)

type BodySkipper struct {
	Skip                                          func(echo.Context) (skipReqBody bool, skipRespBody bool)
	regexExcludedPathsReq, regexExcludedPathsResp []*regexp.Regexp
}

type SkipperConf struct {
	// paths (regular expressions) or endpoints (ex: `/ping/:id`) to exclude from dumping response bodies
	DumpNoResponseBodyForPaths []string

	// paths (regular expressions) or endpoints (ex: `/ping/:id`) to exclude from dumping request bodies
	DumpNoRequestBodyForPaths []string
}

func (b *BodySkipper) prepareRegexps(config SkipperConf) {
	b.regexExcludedPathsResp = make([]*regexp.Regexp, 0, len(config.DumpNoResponseBodyForPaths))
	b.regexExcludedPathsReq = make([]*regexp.Regexp, 0, len(config.DumpNoRequestBodyForPaths))

	if len(config.DumpNoResponseBodyForPaths) > 0 {
		for _, path := range config.DumpNoResponseBodyForPaths {
			regexExcludedPath, err := regexp.Compile(path)
			if err != nil {
				// if the pattern is invalid / not regular expression - just ignore it
				continue
			}

			b.regexExcludedPathsResp = append(b.regexExcludedPathsResp, regexExcludedPath)
		}
	}

	if len(config.DumpNoRequestBodyForPaths) > 0 {
		for _, path := range config.DumpNoRequestBodyForPaths {
			regexExcludedPath, err := regexp.Compile(path)
			if err != nil {
				// if the pattern is invalid / not regular expression - just ignore it
				continue
			}

			b.regexExcludedPathsReq = append(b.regexExcludedPathsReq, regexExcludedPath)
		}
	}
}

func isExcluded(path string, endpoint string, regexps []*regexp.Regexp, endpoints []string) bool {
	if len(endpoints) == 0 {
		return false
	}

	for _, endpointExcluded := range endpoints {
		if endpointExcluded == endpoint {
			return true
		}
	}

	if len(regexps) == 0 {
		return false
	}

	for _, regexExcludedPath := range regexps {
		if regexExcludedPath.MatchString(path) {
			return true
		}
	}

	return false
}

func NewEchoDumpBodySkipper(config SkipperConf) *BodySkipper {
	b := &BodySkipper{}

	if len(config.DumpNoResponseBodyForPaths) == 0 && len(config.DumpNoRequestBodyForPaths) == 0 {
		b.Skip = func(c echo.Context) (bool, bool) {
			return false, false
		}

		return b
	}

	b.prepareRegexps(config)

	b.Skip = func(c echo.Context) (bool, bool) {
		skipReqBody := isExcluded(c.Request().URL.Path, c.Path(), b.regexExcludedPathsReq, config.DumpNoRequestBodyForPaths)
		skipRespBody := isExcluded(c.Request().URL.Path, c.Path(), b.regexExcludedPathsResp, config.DumpNoResponseBodyForPaths)

		return skipReqBody, skipRespBody
	}

	return b
}
