# echo-dump-body-skipper

A skipper utility for the [Echo](https://echo.labstack.com/) web framework in Go. It lets you conditionally skip request/response body dumping for specific routes or requests, improving performance and reducing sensitive data exposure in logs.

## Features
- Customizable skipper logic for Echo middlewares
    - [echo-otel-middleware](https://github.com/adlandh/echo-otel-middleware)
    - [echo-sentry-middleware](https://github.com/adlandh/echo-sentry-middleware)
- Supports both exact endpoint matching and regular expression pattern matching
- Separate skip lists for request and response bodies
- Safe behavior when config is empty or regex patterns are invalid

## Installation

```sh
go get github.com/adlandh/echo-dump-body-skipper/v2
```

## Usage

### Basic

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

### Configuration

`SkipperConf` supports two lists:

- `DumpNoRequestBodyForPaths`: skip request body dumping
- `DumpNoResponseBodyForPaths`: skip response body dumping

Each list can include:

- Exact route patterns (Echo route syntax, e.g. `/ping/:id`)
- Regular expressions (e.g. `^/ping/121$`)

Matching rules:

- Exact route checks compare against `echo.Context.Path()`
- For regex patterns, prefer anchoring with `^...$` for clarity and to avoid unintended matches
- Regex checks compare against `echo.Context.Request().URL.Path`
- Query strings are ignored for regex matching since only the path is checked
- Invalid regex patterns are silently ignored

### Examples

Skip request body for any `/users/:id` route while skipping response body for a specific path:

```go
skipper := echodumpbodyskipper.NewEchoDumpBodySkipper(
	echodumpbodyskipper.SkipperConf{
		DumpNoRequestBodyForPaths: []string{
			"/users/:id",
		},
		DumpNoResponseBodyForPaths: []string{
			"^/users/42$",
		},
	},
)
```

### Middleware Integration

#### echo-otel-middleware

```go
skipper := echodumpbodyskipper.NewEchoDumpBodySkipper(
	echodumpbodyskipper.SkipperConf{
		DumpNoResponseBodyForPaths: []string{
			"/health",
		},
	},
)

e.Use(echootelmiddleware.MiddlewareWithConfig(echootelmiddleware.OtelConfig{
	AreHeadersDump: true,
	IsBodyDump:     true,
	BodySkipper:    skipper.Skip,
}))
```

#### echo-sentry-middleware

```go
skipper := echodumpbodyskipper.NewEchoDumpBodySkipper(
	echodumpbodyskipper.SkipperConf{
		DumpNoRequestBodyForPaths: []string{
			"/auth/:provider/callback",
		},
	},
)

e.Use(echosentry.MiddlewareWithConfig(echosentry.Config{
	BodySkipper: skipper.Skip,
}))
```

### Testing

```sh
go test ./...
```
