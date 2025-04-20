# echo-dump-body-skipper

A middleware/skipper utility for the [Echo](https://echo.labstack.com/) web framework in Go. It allows you to conditionally skip request/response body dumping for specific routes or requests, improving performance and security in logging scenarios.

## Features
- Customizable skipper logic for Echo middlewares 
    - [echo-otel-middleware](https://github.com/adlandh/echo-otel-middleware)
    - [echo-sentry-middleware](https://github.com/adlandh/echo-sentry-middleware)
- Supports flexible rules for skipping body dumps
- Supports both exact endpoint matching and regular expression pattern matching

## Installation

```sh
go get github.com/adlandh/echo-dump-body-skipper
```

## Usage

```go
e := echo.New()

// Example: Use skipper to skip body dump for health check endpoint
skipper := echodumpbodyskipper.NewEchoDumpBodySkipper(
	echodumpbodyskipper.SkipperConf{
        DumpNoResponseBodyForPaths: []string{"/health"},
})

app.Use(echootelmiddleware.MiddlewareWithConfig(echootelmiddleware.OtelConfig{
    AreHeadersDump: true, // dump request && response headers
    IsBodyDump:     true, // dump request && response body
    // No dump for health check endpoint
    BodySkipper: skipper.Skip,
}))
```
