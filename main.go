package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://081d914fbbf5a6c403c19e863c36dfa2@o4507688756445184.ingest.us.sentry.io/4507688758214656",
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if hint.Context != nil {
				if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					event.Request = sentry.NewRequest(req)
				}
			}
			return event
		},
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	app := echo.New()

	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	// Middleware Sentry untuk menangani error
	app.Use(sentryecho.New(sentryecho.Options{
		Repanic:         true,
		WaitForDelivery: true,
		Timeout:         3 * time.Second,
	}))

	// Middleware untuk menambahkan tag ke scope Sentry
	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
				hub.Scope().SetTag("someRandomTag", "maybeYouNeedIt")
			}
			return next(ctx)
		}
	})

	// Handler contoh
	app.GET("/", func(ctx echo.Context) error {
		if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("unwantedQuery", "someQueryDataMaybe")
				hub.CaptureMessage("User provided unwanted query string, but we recovered just fine")
			})
		}
		return ctx.String(http.StatusOK, "Hello, World!")
	})

	// Endpoint untuk menguji error
	app.GET("/test-error", func(ctx echo.Context) error {
		// Buat error yang bisa ditangkap oleh Sentry
		err := fmt.Errorf("this is a test error to check Sentry integration")
		if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("endpoint", "/test-error")
				hub.CaptureException(err)
			})
		}
		return err
	})

	// Endpoint untuk menguji panic
	app.GET("/test-panic", func(ctx echo.Context) error {
		// Buat panic yang bisa ditangkap oleh Sentry
		panic("this is a test panic to check Sentry integration")
	})

	app.Logger.Fatal(app.Start(":2310"))
}
