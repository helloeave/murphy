package murphy

import (
	"main/util/values"
	"net/http"
)

// responseWriter is the http.ResponseWriter given to wrapped handlers. It
// indirects certain operations to provide additional functionality.
type responseWriter struct {
	http.ResponseWriter
	skipResponse bool
}

var _ http.ResponseWriter = &responseWriter{}

func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.skipResponse = (status != http.StatusOK)
}

type HttpContext interface {
	W() http.ResponseWriter
	R() *http.Request
	Now() values.DateTime
}

type httpContext struct {
	w   http.ResponseWriter
	r   *http.Request
	now values.DateTime
}

// Assert that httpContext implements HttpContext interface.
var _ HttpContext = &httpContext{}

func NewHttpContext(w http.ResponseWriter, r *http.Request) HttpContext {
	return &httpContext{
		w:   w,
		r:   r,
		now: values.Now(),
	}
}

func (c *httpContext) W() http.ResponseWriter {
	return c.w
}

func (c *httpContext) R() *http.Request {
	return c.r
}

func (c *httpContext) Now() values.DateTime {
	return c.now
}
