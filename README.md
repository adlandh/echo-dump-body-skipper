# echo-dump-body-skipper

A skipper utility for the [Echo](https://echo.labstack.com/) web framework in Go. It lets you conditionally skip request/response body dumping for specific routes or requests, improving performance and reducing sensitive data exposure in logs.

## Features
- Customizable skipper logic for Echo middlewares
    - [echo-otel-middleware](https://github.com/adlandh/echo-otel-middleware)
    - [echo-sentry-middleware](https://github.com/adlandh/echo-sentry-middleware)
- Supports both exact endpoint matching and regular expression pattern matching
- Separate skip lists for request and response bodies
- Safe behavior when config is empty; invalid regex patterns are reported as errors at construction time

## Installation

```sh
go get github.com/adlandh/echo-dump-body-skipper/v2
```

## Usage

### Basic

```go
e := echo.New()

// Example: Use skipper to skip body dump for health check endpoint
skipper, err := echodumpbodyskipper.New(
	echodumpbodyskipper.SkipperConf{
        DumpNoResponseBodyForPaths: []string{"/health"},
})
if err != nil {
    log.Fatal(err)
}

e.Use(echootelmiddleware.MiddlewareWithConfig(echootelmiddleware.OtelConfig{
    AreHeadersDump: true, // dump request && response headers
    IsBodyDump:     true, // dump request && response body
    // No dump for health check endpoint
    BodySkipper: skipper,
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

- Entries containing `/:` or `/*` (or ending in `*`) are treated as Echo route templates and matched against `echo.Context.Path()`
- All other entries are treated as regular expressions and matched against `echo.Context.Request().URL.Path`
- Regex entries are auto-anchored with `^...$` if not already anchored, so a literal like `/users` matches `/users` exactly and not `/users/123`
- Query strings are ignored since only the path is checked
- Invalid regex patterns cause `New` to return an error

### Examples

Skip request body for any `/users/:id` route while skipping response body for a specific path:

```go
skipper, err := echodumpbodyskipper.New(
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
skipper, err := echodumpbodyskipper.New(
	echodumpbodyskipper.SkipperConf{
		DumpNoResponseBodyForPaths: []string{
			"/health",
		},
	},
)

e.Use(echootelmiddleware.MiddlewareWithConfig(echootelmiddleware.OtelConfig{
	AreHeadersDump: true,
	IsBodyDump:     true,
	BodySkipper:    skipper,
}))
```

#### echo-sentry-middleware

```go
skipper, err := echodumpbodyskipper.New(
	echodumpbodyskipper.SkipperConf{
		DumpNoRequestBodyForPaths: []string{
			"/auth/:provider/callback",
		},
	},
)

e.Use(echosentry.MiddlewareWithConfig(echosentry.Config{
	BodySkipper: skipper,
}))
```

### Testing

```sh
go test ./...
```
