// Package echodumpbodyskipper is a body skipper function for some middleware for echo framework
package echodumpbodyskipper

import (
	"regexp"

	"github.com/labstack/echo/v4"
)

type BodySkipper func(echo.Context) (skipReqBody bool, skipRespBody bool)

type SkipperConf struct {
	// paths (regular expressions) or endpoints (ex: `/ping/:id`) to exclude from dumping response bodies
	DumpNoResponseBodyForPaths []string

	// paths (regular expressions) or endpoints (ex: `/ping/:id`) to exclude from dumping request bodies (regular expressions)
	DumpNoRequestBodyForPaths []string
}

var regexExcludedPathsReq, regexExcludedPathsResp []*regexp.Regexp

func prepareRegexs(config SkipperConf) {
	regexExcludedPathsResp = make([]*regexp.Regexp, 0, len(config.DumpNoResponseBodyForPaths))
	regexExcludedPathsReq = make([]*regexp.Regexp, 0, len(config.DumpNoRequestBodyForPaths))

	if len(config.DumpNoResponseBodyForPaths) > 0 {
		for _, path := range config.DumpNoResponseBodyForPaths {
			regexExcludedPath, err := regexp.Compile(path)
			if err != nil {
				// Just ignore  continue
				continue
			}

			regexExcludedPathsResp = append(regexExcludedPathsResp, regexExcludedPath)
		}
	}

	if len(config.DumpNoRequestBodyForPaths) > 0 {
		for _, path := range config.DumpNoRequestBodyForPaths {
			regexExcludedPath, err := regexp.Compile(path)
			if err != nil {
				// Just ignore  continue
				continue
			}

			regexExcludedPathsReq = append(regexExcludedPathsReq, regexExcludedPath)
		}
	}
}

func isExcluded(path string, endpoint string, regexs []*regexp.Regexp, endpoints []string) bool {
	if len(endpoints) > 0 {
		for _, endpointExcluded := range endpoints {
			if endpointExcluded == endpoint {
				return true
			}
		}
	}

	if len(regexs) > 0 {
		for _, regexExcludedPath := range regexs {
			if regexExcludedPath.MatchString(path) {
				return true
			}
		}
	}

	return false
}

func NewEchoDumpBodySkipper(config SkipperConf) BodySkipper {
	prepareRegexs(config)

	return func(c echo.Context) (bool, bool) {
		skipReqBody := isExcluded(c.Request().URL.Path, c.Path(), regexExcludedPathsReq, config.DumpNoRequestBodyForPaths)
		skipRespBody := isExcluded(c.Request().URL.Path, c.Path(), regexExcludedPathsResp, config.DumpNoResponseBodyForPaths)

		return skipReqBody, skipRespBody
	}
}
