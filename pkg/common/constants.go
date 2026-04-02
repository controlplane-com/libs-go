package common

const (
	FieldTraceID        = "traceId"
	TraceIDKey          = TraceIDType("key")
	RemoteIPKey         = RemoteIPType("key")
	PerformanceSpanKey  = PerformanceSpanType("key")
	HeaderWorkloadLink  = "X-Cpln-Workload-Link"
	HeaderIdentityLink  = "X-Cpln-Identity-Link"
	AuthnTypeBearer     = "Bearer"
)

type DataContextType string
type TraceIDType string
type RemoteIPType string
type PerformanceSpanType string
