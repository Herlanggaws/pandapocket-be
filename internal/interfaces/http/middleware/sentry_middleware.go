package middleware

import (
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// SentryHTTP tags the request path and reports 5xx responses without bodies.
func SentryHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		hub := sentry.GetHubFromContext(c.Request.Context())
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
		}
		hub.Scope().SetTag("endpoint", c.FullPath())
		c.Next()

		if userID := c.GetInt("user_id"); userID > 0 {
			hub.Scope().SetUser(sentry.User{ID: strconv.Itoa(userID)})
		}
		status := c.Writer.Status()
		if status >= 500 {
			endpoint := c.FullPath()
			if endpoint == "" {
				endpoint = c.Request.URL.Path
			}
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("http.status_code", strconv.Itoa(status))
				scope.SetTag("endpoint", endpoint)
				if userID := c.GetInt("user_id"); userID > 0 {
					scope.SetUser(sentry.User{ID: strconv.Itoa(userID)})
				}
				hub.CaptureMessage("HTTP " + strconv.Itoa(status) + " " + c.Request.Method + " " + endpoint)
			})
		}
	}
}
