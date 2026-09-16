package route

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/controlplane-com/libs-go/pkg/common"
)

// HeaderRequestID carries the caller's trace id; it is echoed back as the request's trace id.
const HeaderRequestID = "X-Request-Id"

// requestDurationHistogram is registered once per process with the default registry. go-libs
// may be linked into several binaries, and a package-level promauto var guarantees a single
// registration no matter how many packages call Route.
var requestDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_duration_seconds",
	Help:    "Total http requests since",
	Buckets: []float64{0.01, 0.03, 0.05, 0.1, 0.3, 0.6, 1, 3, 5, 9, 20, 60},
}, []string{"method", "path"})

// Route registers routerHandler at path, wrapping it so that every request gets a trace id and
// remote IP in its context and its duration observed in http_request_duration_seconds{method,path}.
func Route(rtr *mux.Router, path string, routerHandler http.HandlerFunc) *mux.Route {
	return rtr.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		request = InjectCommonRequestContextVars(request)
		addDefaultResponseHeaders(request.Context(), request.Header)
		routerHandler(writer, request)
		requestDurationSeconds := time.Now().Sub(start).Seconds()
		labels := map[string]string{"method": request.Method, "path": path}
		requestDurationHistogram.With(labels).Observe(requestDurationSeconds)
	})
}

// InjectCommonRequestContextVars returns the request with a context carrying the trace id from
// the X-Request-Id header (a fresh UUID when absent) and the caller's remote IP.
func InjectCommonRequestContextVars(r *http.Request) *http.Request {
	requestID := r.Header.Get(HeaderRequestID)
	if requestID == "" {
		requestID = r.Header.Get(strings.ToLower(HeaderRequestID))
	}
	ip := r.Header.Get("X-Forwarded-For")
	if len(ip) == 0 {
		ip = r.RemoteAddr
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, common.TraceIDKey, requestID)
	ctx = InjectCommonContextVars(ctx)
	return r.WithContext(ctx)
}

// InjectCommonContextVars ensures ctx carries a non-empty trace id and a remote IP value
// (empty when unknown), so downstream logging can rely on both keys being present.
func InjectCommonContextVars(ctx context.Context) context.Context {
	requestId := ctx.Value(common.TraceIDKey)
	if requestId == "" || requestId == nil {
		requestId = common.NewUuid()
	}
	ctx = context.WithValue(ctx, common.TraceIDKey, requestId)

	remoteIp := ctx.Value(common.RemoteIPKey)
	if remoteIp == nil {
		remoteIp = ""
	}
	ctx = context.WithValue(ctx, common.RemoteIPKey, remoteIp)
	return ctx
}

func addDefaultResponseHeaders(ctx context.Context, header http.Header) {
	header.Add(HeaderRequestID, ctx.Value(common.TraceIDKey).(string))
}
